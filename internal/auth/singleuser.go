package auth

import (
	"github.com/gin-gonic/gin"
)

type singleUserAuthenticator struct {
	userId string
}

func NewSingleUserAuthenticator(userId string) Authenticator {
	return &singleUserAuthenticator{userId: userId}
}

func (a *singleUserAuthenticator) SessionMiddleware() gin.HandlerFunc {
	return nil
}

func (a *singleUserAuthenticator) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(UserIdContextKey, a.userId)
		ctx.Next()
	}
}

func (a *singleUserAuthenticator) RegisterRoutes(r *gin.Engine) {}

func (a *singleUserAuthenticator) LogoutEndpoint() string {
	return ""
}
