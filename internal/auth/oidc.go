package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

const (
	loginPath    = "/auth/login"
	callbackPath = "/auth/callback"
	logoutPath   = "/auth/logout"

	sessionKeyState    = "oidc_state"
	sessionKeyNonce    = "oidc_nonce"
	sessionKeyVerifier = "oidc_code_verifier"
	sessionKeyRedirect = "oidc_redirect"
	sessionKeyUserId   = "user_id"
)

type oidcAuthenticator struct {
	verifier          *oidc.IDTokenVerifier
	oauth2Config      *oauth2.Config
	sessionMiddleware gin.HandlerFunc
	redirectURL       string
	cookieSecure      bool
	maxAge            int
	providerLogoutURL string
}

func NewOidcAuthenticator(cfg *config.ServerConfiguration) (Authenticator, error) {
	if cfg.Oauth2ClientId == "" || cfg.Oauth2Issuer == "" {
		return nil, ErrNoAuthConfigured
	}
	if cfg.SessionRedisUrl == "" {
		return nil, errors.New("OPENFOODDIARY_SESSION_REDIS_URL is required for the oauth2 login flow")
	}
	if cfg.SessionCookieName == "" {
		return nil, errors.New("OPENFOODDIARY_SESSION_COOKIE_NAME must not be empty")
	}

	sessionMiddleware, err := newSessionMiddleware(cfg)
	if err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(context.Background(), cfg.Oauth2Issuer)
	if err != nil {
		return nil, fmt.Errorf("could not reach the oidc issuer %q: %w", cfg.Oauth2Issuer, err)
	}

	return &oidcAuthenticator{
		verifier: provider.Verifier(&oidc.Config{
			ClientID:          cfg.Oauth2ClientId,
			SkipIssuerCheck:   cfg.Oauth2SkipIssuerVerification,
			SkipClientIDCheck: cfg.Oauth2SkipIssuerVerification,
		}),
		oauth2Config: &oauth2.Config{
			ClientID:     cfg.Oauth2ClientId,
			ClientSecret: cfg.Oauth2ClientSecret,
			RedirectURL:  cfg.Oauth2RedirectUrl,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID},
		},
		sessionMiddleware: sessionMiddleware,
		redirectURL:       cfg.Oauth2RedirectUrl,
		cookieSecure:      cfg.SessionSecure,
		maxAge:            cfg.SessionMaxAge,
		providerLogoutURL: endSessionEndpoint(provider),
	}, nil
}

func (a *oidcAuthenticator) SessionMiddleware() gin.HandlerFunc {
	return a.sessionMiddleware
}

func (a *oidcAuthenticator) redirectURLFor(ctx *gin.Context) string {
	if a.redirectURL != "" {
		return a.redirectURL
	}
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := ctx.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}
	host := ctx.Request.Host
	if forwarded := ctx.GetHeader("X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}
	if host == "" {
		return ""
	}
	return scheme + "://" + host + callbackPath
}

func (a *oidcAuthenticator) RegisterRoutes(r *gin.Engine) {
	r.GET(loginPath, a.handleLogin)
	r.GET(callbackPath, a.handleCallback)
	r.GET(logoutPath, a.handleLogout)
}

func (a *oidcAuthenticator) LogoutEndpoint() string {
	return logoutPath
}

func (a *oidcAuthenticator) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if isExempt(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}
		if userId, ok := a.sessionUserId(ctx); ok {
			ctx.Set(UserIdContextKey, userId)
			ctx.Next()
			return
		}
		loginRedirect(ctx)
		ctx.Abort()
	}
}

func (a *oidcAuthenticator) sessionUserId(ctx *gin.Context) (string, bool) {
	raw := sessions.Default(ctx).Get(sessionKeyUserId)
	userId, ok := raw.(string)
	if !ok || userId == "" {
		return "", false
	}
	return userId, true
}

func (a *oidcAuthenticator) handleLogin(ctx *gin.Context) {
	if _, ok := a.sessionUserId(ctx); ok {
		ctx.Redirect(http.StatusSeeOther, postLoginTarget(ctx))
		return
	}

	state, err := randomToken()
	if err != nil {
		abortAuth(ctx, err)
		return
	}
	nonce, err := randomToken()
	if err != nil {
		abortAuth(ctx, err)
		return
	}
	verifier := oauth2.GenerateVerifier()

	redirectUrl := a.redirectURLFor(ctx)
	if redirectUrl == "" {
		abortAuth(ctx, errors.New("could not determine the oauth2 callback url, set OPENFOODDIARY_OAUTH2_REDIRECT_URL"))
		return
	}

	sess := sessions.Default(ctx)
	sess.Set(sessionKeyState, state)
	sess.Set(sessionKeyNonce, nonce)
	sess.Set(sessionKeyVerifier, verifier)
	sess.Set(sessionKeyRedirect, postLoginTarget(ctx))
	sess.Options(a.cookieOptions())
	if err := sess.Save(); err != nil {
		abortAuth(ctx, err)
		return
	}

	oauth2Config := *a.oauth2Config
	oauth2Config.RedirectURL = redirectUrl

	ctx.Redirect(http.StatusSeeOther, oauth2Config.AuthCodeURL(
		state,
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(verifier),
	))
}

func (a *oidcAuthenticator) handleCallback(ctx *gin.Context) {
	sess := sessions.Default(ctx)

	expectedState, _ := sess.Get(sessionKeyState).(string)
	expectedNonce, _ := sess.Get(sessionKeyNonce).(string)
	verifier, _ := sess.Get(sessionKeyVerifier).(string)
	redirect, _ := sess.Get(sessionKeyRedirect).(string)

	sess.Delete(sessionKeyState)
	sess.Delete(sessionKeyNonce)
	sess.Delete(sessionKeyVerifier)
	sess.Delete(sessionKeyRedirect)

	if err := ctx.Request.ParseForm(); err != nil {
		abortAuth(ctx, err)
		return
	}
	if providerErr := ctx.Request.Form.Get("error"); providerErr != "" {
		abortAuth(ctx, fmt.Errorf("identity provider returned an error: %s", providerErr))
		return
	}
	if expectedState == "" || ctx.Request.Form.Get("state") != expectedState {
		abortAuth(ctx, errors.New("oauth2 state mismatch, possible cross-site request forgery"))
		return
	}
	code := ctx.Request.Form.Get("code")
	if code == "" {
		abortAuth(ctx, errors.New("no authorization code in the oauth2 callback"))
		return
	}

	redirectUrl := a.redirectURLFor(ctx)
	if redirectUrl == "" {
		abortAuth(ctx, errors.New("could not determine the oauth2 callback url, set OPENFOODDIARY_OAUTH2_REDIRECT_URL"))
		return
	}
	oauth2Config := *a.oauth2Config
	oauth2Config.RedirectURL = redirectUrl

	token, err := oauth2Config.Exchange(ctx.Request.Context(), code, oauth2.VerifierOption(verifier))
	if err != nil {
		abortAuth(ctx, fmt.Errorf("could not exchange the authorization code: %w", err))
		return
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok || rawIdToken == "" {
		abortAuth(ctx, errors.New("no id_token in the oauth2 token response"))
		return
	}

	idToken, err := a.verifier.Verify(ctx.Request.Context(), rawIdToken)
	if err != nil {
		abortAuth(ctx, fmt.Errorf("could not verify the id_token: %w", err))
		return
	}
	if idToken.Nonce != expectedNonce {
		abortAuth(ctx, errors.New("id_token nonce mismatch"))
		return
	}

	if idToken.Subject == "" {
		abortAuth(ctx, errors.New("the id_token has no sub claim"))
		return
	}

	sess.Set(sessionKeyUserId, idToken.Subject)
	sess.Options(a.cookieOptions())
	if err := sess.Save(); err != nil {
		abortAuth(ctx, err)
		return
	}

	if !isSafeRedirect(redirect) {
		redirect = "/logs"
	}
	ctx.Redirect(http.StatusSeeOther, redirect)
}

func (a *oidcAuthenticator) handleLogout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Clear()
	sess.Options(sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	if err := sess.Save(); err != nil {
		abortAuth(ctx, err)
		return
	}
	if a.providerLogoutURL != "" {
		ctx.Redirect(http.StatusSeeOther, a.providerLogoutURL)
		return
	}
	ctx.Redirect(http.StatusSeeOther, loginPath)
}

func (a *oidcAuthenticator) cookieOptions() sessions.Options {
	return sessions.Options{
		Path:     "/",
		MaxAge:   a.maxAge,
		HttpOnly: true,
		Secure:   a.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func postLoginTarget(ctx *gin.Context) string {
	if redirect := ctx.Query("redirect"); isSafeRedirect(redirect) {
		return redirect
	}
	if referer := ctx.Request.Referer(); referer != "" {
		if u, err := url.Parse(referer); err == nil && isSafeRedirect(u.Path) {
			return u.Path
		}
	}
	return "/logs"
}

func endSessionEndpoint(provider *oidc.Provider) string {
	claims := struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}{}
	if err := provider.Claims(&claims); err != nil {
		return ""
	}
	return claims.EndSessionEndpoint
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func abortAuth(ctx *gin.Context, err error) {
	log.Error().Err(err).Str("path", ctx.Request.URL.Path).Msg("auth flow failed")
	reject(ctx, http.StatusUnauthorized, errors.New("authentication failed"))
}
