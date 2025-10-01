package server

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/server/generated"
	"github.com/openhealthsuite/diary/internal/storage"
	"github.com/rs/zerolog/log"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../tools/oapi_codegen/server.cfg.yaml ../../api/swagger.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../tools/oapi_codegen/types.cfg.yaml ../../api/swagger.yaml

type DiaryServer interface {
	RunServer() error
}

type DiaryServerState struct {
	GeneratedInterface generated.ServerInterface
	Config             *config.ServerConfiguration
}

func NewServer(cfg *config.ServerConfiguration) (DiaryServer, error) {
	strg, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, err
	}
	return &DiaryServerState{
		GeneratedInterface: NewGeneratedInterface(ServerState{
			storage: strg,
		}),
		Config: cfg,
	}, nil
}

type ServerState struct {
	storage storage.Storage
}

func NewGeneratedInterface(srvst ServerState) generated.ServerInterface {
	return &srvst
}

// TestEndpoint implements generated.ServerInterface.
func (g *ServerState) TestEndpoint(c *gin.Context) {
	_, err := g.storage.GetQuerier().GetTestData(c)
	if err != nil {
		log.Error().Err(err).Msg("error pinging database")
		c.String(500, "KO")
		return
	}
	c.String(200, "OK")
}

func (sts *DiaryServerState) RunServer() error {
	r := gin.Default()

	r.Static("/static", "./web/static")
	r.StaticFile("/favicon.ico", "./web/static/favicon.ico")

	r.GET("/", serveDatabrowserSPA)
	r.GET("/assets/*path", serveDatabrowserSPA)

	r.Use(func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			serveDatabrowserSPA(ctx)
			ctx.Abort()
			return
		}
		ctx.Next()
	})

	r.Use(func(ctx *gin.Context) {
		if sts.Config.UserId != "" {
			ctx.Set("userId", sts.Config.UserId)
			return
		}

		if ctx.Request.Header.Get(sts.Config.UserIdHeader) != "" {
			ctx.Set("userId", ctx.Request.Header.Get(sts.Config.UserIdHeader))
			return
		}
		ctx.String(403, "Missing User Identification")
	})

	generated.RegisterHandlers(r, sts.GeneratedInterface)

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
