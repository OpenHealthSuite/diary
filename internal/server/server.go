package server

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/auth"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/foodlogs"
	"github.com/openhealthsuite/diary/internal/metrics"
	"github.com/openhealthsuite/diary/internal/server/generated"
	"github.com/openhealthsuite/diary/internal/storage"
	strgtyp "github.com/openhealthsuite/diary/internal/storage/types"
	"github.com/rs/zerolog/log"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../tools/oapi_codegen/server.cfg.yaml ../../api/swagger.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../tools/oapi_codegen/types.cfg.yaml ../../api/swagger.yaml

type DiaryServer interface {
	RunServer() (*chan os.Signal, error)
}

func NewServer(cfg *config.ServerConfiguration) (DiaryServer, error) {
	strg, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, err
	}
	mtr, err := metrics.NewMetricsProvider(strg)
	if err != nil {
		return nil, err
	}
	lgs, err := foodlogs.NewFoodLogService(strg)
	if err != nil {
		return nil, err
	}
	sts := &ServerState{
		Config:   cfg,
		storage:  strg,
		metrics:  mtr,
		foodlogs: lgs,
	}
	err = sts.setupHttpServer()
	if err != nil {
		return nil, err
	}
	return sts, nil
}

func (sts *ServerState) setupHttpServer() error {
	r := gin.Default()

	// Load templates
	r.SetFuncMap(templateFuncs())
	tmpldr := "web/template"
	if sts.Config.TemplateDirectory != "" {
		tmpldr = sts.Config.TemplateDirectory
	}
	tmpl, err := loadTemplates(tmpldr)
	if err != nil {
		return err
	}
	r.SetHTMLTemplate(tmpl)

	stscdir := "web/static"
	if sts.Config.StaticDirectory != "" {
		stscdir = sts.Config.StaticDirectory
	}
	// Static files
	r.StaticFile("/favicon.ico", stscdir+"/favicon.ico")
	r.Static("/static", stscdir)

	// User identification middleware
	r.Use(auth.UserAuthenticationMiddleware(sts.Config))

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

	// Config routes
	r.POST("/config/metrics", sts.handleSaveMetrics)
	r.POST("/config/metrics/new", sts.handleCreateMetric)
	r.DELETE("/config/metrics/:key", sts.handleDeleteMetric)
	r.GET("/config/purge", sts.handlePurgePage)
	r.DELETE("/config/purge", sts.handlePurgeLogs)
	r.GET("/config/upload", sts.handleUploadPage)

	// Tutorial
	r.POST("/tutorial/log", sts.handleTutorialLog)

	// API routes (existing)
	generated.RegisterHandlers(r, sts.AsGeneratedInterface())

	sts.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", sts.Config.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return nil
}

type ServerState struct {
	Config   *config.ServerConfiguration
	storage  strgtyp.Storage
	metrics  metrics.MetricsProvider
	foodlogs foodlogs.FoodLogService

	httpServer *http.Server
}

func (srvst *ServerState) AsGeneratedInterface() generated.ServerInterface {
	return srvst
}

// TestEndpoint implements generated.ServerInterface.
func (g *ServerState) TestEndpoint(c *gin.Context) {
	_, err := g.storage.GetTestData(c)
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

func (sts *ServerState) RunServer() (*chan os.Signal, error) {

	quit := make(chan os.Signal, 1)

	go func() {
		<-quit
		log.Info().Msg("receive interrupt signal")
		if err := sts.httpServer.Close(); err != nil {
			log.Error().Err(err).Msg("Server Close error")
		}
	}()

	go func() {
		if err := sts.httpServer.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Info().Msg("Server closed under request")
			} else {
				log.Error().Err(err).Msg("Server closed unexpectedly")
			}
		}

	}()
	return &quit, nil
}
