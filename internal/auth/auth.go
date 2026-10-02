package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/api/generated"
	"github.com/openhealthsuite/diary/internal/config"
)

const UserIdContextKey = "userId"

type Authenticator interface {
	SessionMiddleware() gin.HandlerFunc
	Middleware() gin.HandlerFunc
	RegisterRoutes(r *gin.Engine)
	LogoutEndpoint() string
}

func NewAuthenticator(cfg *config.ServerConfiguration) (Authenticator, error) {
	if cfg.UserId != "" {
		return NewSingleUserAuthenticator(cfg.UserId), nil
	}
	return NewOidcAuthenticator(cfg)
}

var (
	ErrNoUserIdentification = errors.New("missing user identification")
	ErrNoAuthConfigured     = errors.New("no authentication configured: set either OPENFOODDIARY_USERID or the OPENFOODDIARY_OAUTH2_* variables")
)

func GetUserId(c *gin.Context) (*string, error) {
	userId, ok := c.Get(UserIdContextKey)
	strid, ok2 := userId.(string)
	if !ok || !ok2 {
		return nil, ErrNoUserIdentification
	}
	return &strid, nil
}

func reject(ctx *gin.Context, status int32, err error) {
	ctx.AbortWithStatusJSON(int(status), generated.Error{Code: status, Message: err.Error()})
}
