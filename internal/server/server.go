package server

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/server/generated"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../tools/oapi_codegen/server.cfg.yaml ../../api/swagger.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../tools/oapi_codegen/types.cfg.yaml ../../api/swagger.yaml

type DiaryServer interface {
	RunServer() error
}

type DiaryServerState struct {
	GeneratedInterface generated.ServerInterface
}

func NewServer(cfg *config.ServerConfiguration) (DiaryServer, error) {
	return &DiaryServerState{
		GeneratedInterface: NewGeneratedInterface(ServerState{}),
	}, nil
}

type ServerState struct {
}

func NewGeneratedInterface(srvst ServerState) generated.ServerInterface {
	return &srvst
}

// TestEndpoint implements generated.ServerInterface.
func (g *ServerState) TestEndpoint(c *gin.Context) {
	c.String(200, "OK")
}

func (sts *DiaryServerState) RunServer() error {
	r := gin.Default()

	r.Static("/static", "./web/static")

	r.GET("/", serveDatabrowserSPA)
	r.GET("/assets/*path", serveDatabrowserSPA)

	generated.RegisterHandlers(r, sts.GeneratedInterface)

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
