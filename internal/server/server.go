package server

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/config"
)

type DiaryServer interface {
	RunServer() error
}

type DiaryServerState struct {
}

func NewServer(cfg *config.ServerConfiguration) (DiaryServer, error) {
	return &DiaryServerState{}, nil
}

func (sts *DiaryServerState) RunServer() error {
	r := gin.Default()

	r.Static("/static", "./web/static")

	r.GET("/", serveDatabrowserSPA)
	r.GET("/*path", serveDatabrowserSPA)

	r.StaticFile("/favicon.ico", "./web/static/favicon.ico")

	return r.Run()
}

// Serve the React SPA index.html
func serveDatabrowserSPA(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/assets") {
		c.File(strings.Replace(c.Request.URL.Path, "/assets", "./web/dist/assets", 1))
		return
	}
	c.File("./web/dist/index.html")
}
