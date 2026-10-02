package auth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

var exemptPaths = []string{
	"/api/ping",
	"/static/",
	"/favicon.ico",
	"/auth/login",
	"/auth/callback",
	"/auth/logout",
}

func isExempt(path string) bool {
	for _, prefix := range exemptPaths {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func loginRedirect(ctx *gin.Context) {
	target := ctx.Request.URL.RequestURI()
	if isSafeRedirect(target) {
		ctx.Redirect(http.StatusSeeOther, "/auth/login?redirect="+url.QueryEscape(target))
		return
	}
	ctx.Redirect(http.StatusSeeOther, "/auth/login")
}

func isSafeRedirect(target string) bool {
	if !strings.HasPrefix(target, "/") || strings.HasPrefix(target, "//") {
		return false
	}
	return !strings.ContainsAny(target, "\r\n")
}
