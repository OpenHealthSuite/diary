package server

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
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

func loadTemplates(templatesDir string) (*template.Template, error) {
	tmpl := template.New("").Funcs(templateFuncs())

	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			_, err = tmpl.ParseFiles(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return tmpl, err
}

func (sts *DiaryServerState) RunServer() error {
	r := gin.Default()

	// Load templates
	r.SetFuncMap(templateFuncs())
	tmpl, err := loadTemplates("web/template")
	if err != nil {
		return err
	}
	r.SetHTMLTemplate(tmpl)

	// Static files
	r.StaticFile("/favicon.ico", "./web/static/favicon.ico")
	r.Static("/static", "./web/static")

	// User identification middleware
	r.Use(func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/api/ping" {
			ctx.Next()
			return
		}
		// Skip auth for static files
		if strings.HasPrefix(ctx.Request.URL.Path, "/static/") {
			ctx.Next()
			return
		}
		if sts.Config.UserId != "" {
			ctx.Set("userId", sts.Config.UserId)
			ctx.Next()
			return
		}

		if ctx.Request.Header.Get(sts.Config.UserIdHeader) != "" {
			ctx.Set("userId", ctx.Request.Header.Get(sts.Config.UserIdHeader))
			ctx.Next()
			return
		}
		ctx.String(403, "Missing User Identification")
	})

	// Page routes
	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusSeeOther, "/logs")
	})
	r.GET("/config", sts.handleConfig)

	r.GET("/logs", sts.handleLogs)
	r.GET("/logs/new", sts.handleNewLogForm)
	r.GET("/logs/:id", sts.handleEditLogForm)

	// THESE ALL NEED TO HANDLE AND REDIRECT
	r.POST("/logs", sts.handleCreateLog)
	r.PUT("/logs/:id", sts.handleUpdateLog)
	r.DELETE("/logs/:id", sts.handleDeleteLog)

	// Config partials
	r.GET("/config/metrics", sts.handleMetricsPartial)
	r.POST("/htmx/config/metrics", sts.handleSaveMetrics)
	r.POST("/htmx/config/metrics/new", sts.handleCreateMetric)
	r.DELETE("/htmx/config/metrics/:key", sts.handleDeleteMetric)
	r.GET("/config/purge", sts.handlePurgeModal)
	r.DELETE("/htmx/logs/purge", sts.handlePurgeLogs)
	r.GET("/config/upload", sts.handleUploadModal)

	// Tutorial
	r.GET("/tutorial", sts.handleTutorial)
	r.GET("/tutorial/:step", sts.handleTutorialStep)

	// API routes (existing)
	generated.RegisterHandlers(r, sts.GeneratedInterface)

	return r.Run()
}
