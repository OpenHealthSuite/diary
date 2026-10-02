package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetUserId_HappyPath(xt *testing.T) {
	gc, _ := gin.CreateTestContext(httptest.NewRecorder())
	testuserid := "some-user-id"
	gc.Set("userId", testuserid)
	uid, err := GetUserId(gc)
	require.NoError(xt, err)
	assert.Equal(xt, testuserid, *uid)
}

func Test_GetUserId_Missing(xt *testing.T) {
	gc, _ := gin.CreateTestContext(httptest.NewRecorder())
	uid, err := GetUserId(gc)
	require.Error(xt, err)
	assert.Nil(xt, uid)
	assert.Equal(xt, ErrNoUserIdentification, err)
}

func Test_GetUserId_NotString(xt *testing.T) {
	gc, _ := gin.CreateTestContext(httptest.NewRecorder())
	gc.Set("userId", 123)
	uid, err := GetUserId(gc)
	require.Error(xt, err)
	assert.Nil(xt, uid)
	assert.Equal(xt, ErrNoUserIdentification, err)
}

func Test_NewAuthenticator_PrefersSingleUser(xt *testing.T) {
	cfg := newTestConfig()
	cfg.UserId = "pinned-user"
	cfg.Oauth2Issuer = "http://dex.invalid"
	cfg.Oauth2ClientId = "client"
	cfg.SessionRedisUrl = "redis://localhost:6379"

	authr, err := NewAuthenticator(cfg)
	require.NoError(xt, err)
	assert.Equal(xt, "", authr.LogoutEndpoint())
	assert.Nil(xt, authr.SessionMiddleware())
}

func Test_NewAuthenticator_RequiresSomething(xt *testing.T) {
	authr, err := NewAuthenticator(newTestConfig())
	require.ErrorIs(xt, err, ErrNoAuthConfigured)
	assert.Nil(xt, authr)
}

func Test_NewAuthenticator_Oauth2NeedsRedis(xt *testing.T) {
	cfg := newTestConfig()
	cfg.Oauth2Issuer = "http://dex.invalid"
	cfg.Oauth2ClientId = "client"

	authr, err := NewAuthenticator(cfg)
	require.Error(xt, err)
	assert.Nil(xt, authr)
}

func Test_NewAuthenticator_Oauth2NeedsIssuerAndClientId(xt *testing.T) {
	cfg := newTestConfig()
	cfg.Oauth2ClientId = "client"
	cfg.SessionRedisUrl = "redis://localhost:6379"

	_, err := NewAuthenticator(cfg)
	require.ErrorIs(xt, err, ErrNoAuthConfigured)

	cfg = newTestConfig()
	cfg.Oauth2Issuer = "http://dex.invalid"
	cfg.SessionRedisUrl = "redis://localhost:6379"

	_, err = NewAuthenticator(cfg)
	require.ErrorIs(xt, err, ErrNoAuthConfigured)
}

func Test_NewAuthenticator_UnreachableIssuer(xt *testing.T) {
	cfg := newTestConfig()
	cfg.Oauth2Issuer = "http://127.0.0.1:1/dex"
	cfg.Oauth2ClientId = "client"
	cfg.SessionRedisUrl = "redis://localhost:6379"

	authr, err := NewAuthenticator(cfg)
	require.Error(xt, err)
	assert.Nil(xt, authr)
}

func Test_SingleUserMiddleware_PinsUserId(xt *testing.T) {
	authr := NewSingleUserAuthenticator("pinned-user")

	r := gin.New()
	r.Use(authr.Middleware())
	authr.RegisterRoutes(r)
	r.GET("/thing", func(ctx *gin.Context) {
		uid, err := GetUserId(ctx)
		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}
		ctx.String(200, *uid)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/thing", nil))

	assert.Equal(xt, 200, rec.Code)
	assert.Equal(xt, "pinned-user", rec.Body.String())
}

func Test_SingleUserMiddleware_IgnoresAnySuppliedIdentity(xt *testing.T) {
	authr := NewSingleUserAuthenticator("pinned-user")

	r := gin.New()
	r.Use(authr.Middleware())
	r.GET("/thing", func(ctx *gin.Context) {
		uid, err := GetUserId(ctx)
		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}
		ctx.String(200, *uid)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/thing", nil)
	req.Header.Set("x-openfooddiary-userid", "someone-else")
	r.ServeHTTP(rec, req)

	assert.Equal(xt, 200, rec.Code)
	assert.Equal(xt, "pinned-user", rec.Body.String())
}
