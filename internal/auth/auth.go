package auth

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/server/generated"
)

func UserAuthenticationMiddleware(cfg *config.ServerConfiguration) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/api/ping" {
			ctx.Next()
			return
		}
		// Skip auth for static files
		if strings.HasPrefix(ctx.Request.URL.Path, "/static/") {
			ctx.Next()
			return
		}
		if cfg.UserId != "" {
			ctx.Set("userId", cfg.UserId)
			ctx.Next()
			return
		}

		if ctx.Request.Header.Get(cfg.UserIdHeader) != "" {
			ctx.Set("userId", ctx.Request.Header.Get(cfg.UserIdHeader))
			ctx.Next()
			return
		}
		ctx.AbortWithStatusJSON(403, generated.Error{Code: 403, Message: ErrNoUserIdentification.Error()})
	}
}

var (
	ErrNoUserIdentification = errors.New("missing user identification")
)

func GetUserId(c *gin.Context) (*string, error) {
	userId, ok := c.Get("userId")
	strid, ok2 := userId.(string)
	if !ok || !ok2 {
		return nil, ErrNoUserIdentification
	}
	return &strid, nil
}
