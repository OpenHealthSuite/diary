package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct {
	server  *httptest.Server
	key     *ecdsa.PrivateKey
	subject string
	email   string
	issuer  string

	mu               sync.Mutex
	authorizeHandler func(w http.ResponseWriter, r *http.Request)
	authzQuery       url.Values
	issuedNonce      string
	codeVerifier     string
	codeChallenge    string
	subjectOverride  string
	nonceOverride    string
	audienceOverride string
	signKeyOverride  *ecdsa.PrivateKey
	omitIdToken      bool
	tokenError       string
}

func newFakeProvider(t *testing.T) *fakeProvider {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	p := &fakeProvider{key: key, subject: "dex-user-1", email: "someone@example.com"}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", p.discovery)
	mux.HandleFunc("/keys", p.keys)
	mux.HandleFunc("/auth", p.authorize)
	mux.HandleFunc("/token", p.token)

	p.server = httptest.NewServer(mux)
	t.Cleanup(p.server.Close)
	p.issuer = p.server.URL
	return p
}

func (p *fakeProvider) discovery(w http.ResponseWriter, r *http.Request) {
	writeJson(w, map[string]any{
		"issuer":                                p.issuer,
		"authorization_endpoint":                p.issuer + "/auth",
		"token_endpoint":                        p.issuer + "/token",
		"jwks_uri":                              p.issuer + "/keys",
		"end_session_endpoint":                  p.issuer + "/logout",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"ES256"},
		"scopes_supported":                      []string{"openid", "profile", "email"},
	})
}

func (p *fakeProvider) keys(w http.ResponseWriter, r *http.Request) {
	jwk := jose.JSONWebKey{
		Key:       p.key.Public(),
		Algorithm: string(jose.ES256),
		Use:       "sig",
		KeyID:     "test-key-1",
	}
	writeJson(w, map[string]any{"keys": []jose.JSONWebKey{jwk}})
}

func (p *fakeProvider) authorize(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	handler := p.authorizeHandler
	p.authzQuery = r.URL.Query()
	p.issuedNonce = r.URL.Query().Get("nonce")
	p.codeChallenge = r.URL.Query().Get("code_challenge")
	p.mu.Unlock()

	if handler != nil {
		handler(w, r)
		return
	}
	state := r.URL.Query().Get("state")
	redirectUrl := r.URL.Query().Get("redirect_uri")
	if redirectUrl == "" {
		http.Error(w, "no redirect_uri", http.StatusBadRequest)
		return
	}
	target, err := url.Parse(redirectUrl)
	if err != nil {
		http.Error(w, "bad redirect_uri", http.StatusBadRequest)
		return
	}
	query := target.Query()
	query.Set("code", "valid-code")
	query.Set("state", state)
	target.RawQuery = query.Encode()
	http.Redirect(w, r, target.String(), http.StatusSeeOther)
}

func (p *fakeProvider) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientId := r.Form.Get("client_id")
	if basicId, _, ok := r.BasicAuth(); ok && clientId == "" {
		clientId = basicId
	}

	p.mu.Lock()
	p.codeVerifier = r.Form.Get("code_verifier")
	p.mu.Unlock()

	if p.tokenError != "" {
		writeJson(w, map[string]string{"error": p.tokenError})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if p.omitIdToken {
		writeJson(w, map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
		return
	}

	subject := p.subject
	if p.subjectOverride != "" {
		subject = p.subjectOverride
	}
	audience := p.audienceOverride
	if audience == "" {
		audience = clientId
	}
	nonce := p.issuedNonce
	if p.nonceOverride != "" {
		nonce = p.nonceOverride
	}

	now := time.Now()
	idToken, err := p.signIdToken(map[string]any{
		"iss":   p.issuer,
		"aud":   audience,
		"sub":   subject,
		"email": p.email,
		"nonce": nonce,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJson(w, map[string]any{
		"access_token": "test-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     idToken,
	})
}

func (p *fakeProvider) signIdToken(claims map[string]any) (string, error) {
	signingKey := p.signKeyOverride
	if signingKey == nil {
		signingKey = p.key
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: signingKey},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test-key-1"),
	)
	if err != nil {
		return "", err
	}
	return jwt.Signed(signer).Claims(claims).Serialize()
}

func writeJson(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func requireRedisUrl(t *testing.T) string {
	t.Helper()
	require.NotEmpty(t, testRedisUrl, "the test redis container did not start")
	return testRedisUrl
}

func newOidcTestConfig(p *fakeProvider, redisUrl string) *config.ServerConfiguration {
	return &config.ServerConfiguration{
		Oauth2Issuer:       p.issuer,
		Oauth2ClientId:     "ofd-client",
		Oauth2ClientSecret: "ofd-secret",
		SessionCookieName:  "openfooddiary-session",
		SessionMaxAge:      3600,
		SessionSecure:      false,
		SessionRedisUrl:    redisUrl,
		SessionSecret:      "test-secret",
	}
}

func newOidcTestEngine(t *testing.T, cfg *config.ServerConfiguration) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	authr, err := NewAuthenticator(cfg)
	require.NoError(t, err)

	r := gin.New()
	r.Use(authr.SessionMiddleware())
	r.Use(authr.Middleware())
	authr.RegisterRoutes(r)
	r.GET("/logs", func(ctx *gin.Context) {
		uid, err := GetUserId(ctx)
		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}
		ctx.String(200, "logs-for:"+*uid)
	})
	r.GET("/config", func(ctx *gin.Context) { ctx.String(200, "config-page") })
	return r
}

func newBrowserClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func Test_Oidc_HappyPath(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	client := newBrowserClient(t)

	resp := followRedirects(client, server.URL+"/logs")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "logs-for:"+provider.subject, readBody(t, resp))

	resp = followRedirects(client, server.URL+"/auth/login")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "logs-for:"+provider.subject, readBody(t, resp))
}

func Test_Oidc_LoginIssuesStateNonceAndPkce(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	client := newBrowserClient(t)
	resp := followRedirects(client, server.URL+"/logs")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	provider.mu.Lock()
	authz := provider.authzQuery
	challenge := provider.codeChallenge
	provider.mu.Unlock()

	assert.NotEmpty(t, authz.Get("state"))
	assert.NotEmpty(t, authz.Get("nonce"))
	assert.Equal(t, "S256", authz.Get("code_challenge_method"))
	assert.NotEmpty(t, challenge)
	assert.Contains(t, authz.Get("scope"), "openid")

	provider.mu.Lock()
	verifier := provider.codeVerifier
	provider.mu.Unlock()
	assert.Equal(t, challenge, pkceChallengeS256(verifier))
}

func Test_Oidc_RedirectsToLoginAndRemembersTarget(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	client := newBrowserClient(t)

	resp, err := browserGet(client, server.URL+"/config")
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)

	location, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "/auth/login", location.Path)
	assert.Equal(t, "/config", location.Query().Get("redirect"))

	final := followRedirects(client, server.URL+"/auth/login")
	require.Equal(t, http.StatusOK, final.StatusCode)
	assert.Equal(t, "logs-for:"+provider.subject, readBody(t, final))

	back := followRedirects(client, server.URL+"/config")
	assert.Equal(t, http.StatusOK, back.StatusCode)
	assert.Equal(t, "config-page", readBody(t, back))
}

func Test_Oidc_ApiGets401(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/logs")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, readBody(t, resp))
}

func Test_Oidc_PingIsExempt(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	r.GET("/api/ping", func(ctx *gin.Context) { ctx.String(200, "OK") })
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/ping")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, readBody(t, resp))
}

func Test_Oidc_HtmxGets401(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL+"/logs", nil)
	require.NoError(t, err)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("Accept", "text/html")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_Oidc_ApiAuthenticated(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	cfg := newOidcTestConfig(provider, redisUrl)
	r := newOidcTestEngine(t, cfg)
	server := httptest.NewServer(r)
	defer server.Close()

	client := newBrowserClient(t)
	resp := followRedirects(client, server.URL+"/logs")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, err := http.NewRequest("GET", server.URL+"/logs", nil)
	require.NoError(t, err)
	for _, cookie := range client.Jar.Cookies(resp.Request.URL) {
		req.AddCookie(cookie)
	}
	apiResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer apiResp.Body.Close()
	assert.Equal(t, http.StatusOK, apiResp.StatusCode)
}

func Test_Oidc_LogoutClearsSession(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	client := newBrowserClient(t)
	resp := followRedirects(client, server.URL+"/logs")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	logout, err := browserGet(client, server.URL+"/auth/logout")
	require.NoError(t, err)
	logout.Body.Close()
	assert.Equal(t, http.StatusSeeOther, logout.StatusCode)
	assert.Equal(t, provider.issuer+"/logout", logout.Header.Get("Location"))

	after, err := browserGet(client, server.URL+"/logs")
	require.NoError(t, err)
	after.Body.Close()
	assert.Equal(t, http.StatusSeeOther, after.StatusCode)
}

func Test_Oidc_RejectsStateMismatch(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	provider.authorizeHandler = func(w http.ResponseWriter, r *http.Request) {
		target, err := url.Parse(r.URL.Query().Get("redirect_uri"))
		if err != nil {
			http.Error(w, "bad redirect_uri", http.StatusBadRequest)
			return
		}
		query := target.Query()
		query.Set("code", "valid-code")
		query.Set("state", "tampered")
		target.RawQuery = query.Encode()
		http.Redirect(w, r, target.String(), http.StatusSeeOther)
	}
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	resp := followRedirects(newBrowserClient(t), server.URL+"/logs")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_Oidc_RejectsNonceMismatch(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	provider.nonceOverride = "not-the-nonce-we-sent"
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	resp := followRedirects(newBrowserClient(t), server.URL+"/logs")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_Oidc_RejectsForeignAudience(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	provider.audienceOverride = "some-other-client"
	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	resp := followRedirects(newBrowserClient(t), server.URL+"/logs")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_Oidc_RejectsUntrustedSigner(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	rogueKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	provider.signKeyOverride = rogueKey

	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	resp := followRedirects(newBrowserClient(t), server.URL+"/logs")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_Oidc_RejectsMissingIdToken(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	provider.omitIdToken = true

	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	resp := followRedirects(newBrowserClient(t), server.URL+"/logs")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_Oidc_OnlyRequestsOpenIdScope(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	authr, err := NewOidcAuthenticator(newOidcTestConfig(provider, redisUrl))
	require.NoError(t, err)
	assert.Equal(t, []string{"openid"}, authr.(*oidcAuthenticator).oauth2Config.Scopes)
}

func Test_Oidc_UserIdIsTheSubject(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	provider.subject = "sub-abc-123"
	provider.email = "ignored@example.com"

	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	resp := followRedirects(newBrowserClient(t), server.URL+"/logs")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "logs-for:sub-abc-123", readBody(t, resp))
}

func Test_Oidc_RejectsEmptySubject(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	provider.subject = ""

	r := newOidcTestEngine(t, newOidcTestConfig(provider, redisUrl))
	server := httptest.NewServer(r)
	defer server.Close()

	resp := followRedirects(newBrowserClient(t), server.URL+"/logs")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_Oidc_SessionSecretIsStableAcrossRestarts(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	cfg := newOidcTestConfig(provider, redisUrl)

	first := newOidcTestEngine(t, cfg)
	serverA := httptest.NewServer(first)
	client := newBrowserClient(t)
	resp := followRedirects(client, serverA.URL+"/logs")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	serverA.Close()

	serverB := httptest.NewServer(newOidcTestEngine(t, cfg))
	defer serverB.Close()

	afterRestart := followRedirects(client, serverB.URL+"/logs")
	assert.Equal(t, http.StatusOK, afterRestart.StatusCode)
	assert.Equal(t, "logs-for:"+provider.subject, readBody(t, afterRestart))
}

func Test_Oidc_LogoutEndpointReported(t *testing.T) {
	redisUrl := requireRedisUrl(t)
	provider := newFakeProvider(t)
	authr, err := NewOidcAuthenticator(newOidcTestConfig(provider, redisUrl))
	require.NoError(t, err)
	assert.Equal(t, "/auth/logout", authr.LogoutEndpoint())
}

func browserGet(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	return client.Do(req)
}

func followRedirects(client *http.Client, startUrl string) *http.Response {
	current := startUrl
	for i := 0; i < 10; i++ {
		resp, err := browserGet(client, current)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusFound {
			return resp
		}
		location := resp.Header.Get("Location")
		resp.Body.Close()
		if location == "" {
			return resp
		}
		base, err := url.Parse(current)
		if err != nil {
			panic(err)
		}
		next, err := base.Parse(location)
		if err != nil {
			panic(err)
		}
		current = next.String()
	}
	panic("too many redirects")
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	buf := make([]byte, 0, 512)
	tmp := make([]byte, 512)
	for {
		n, err := resp.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return string(buf)
}

func pkceChallengeS256(verifier string) string {
	if verifier == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
