package pages

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/foodlogs"
	"github.com/openhealthsuite/diary/internal/metrics"
	strgtyp "github.com/openhealthsuite/diary/internal/storage/types"
)

type PagesState struct {
	Config   *config.ServerConfiguration
	Storage  strgtyp.Storage
	Metrics  metrics.MetricsProvider
	FoodLogs foodlogs.FoodLogService
}

func Setup(
	config *config.ServerConfiguration,
	storage strgtyp.Storage,
	metrics metrics.MetricsProvider,
	foodLogs foodlogs.FoodLogService,
	r *gin.Engine,
) error {
	sts := &PagesState{
		Config:   config,
		Storage:  storage,
		FoodLogs: foodLogs,
		Metrics:  metrics,
	}

	err := sts.setupTemplateRendering(r)
	if err != nil {
		return err
	}

	sts.registerStatic(r)

	sts.registerRoutes(r)

	return nil
}

func (sts *PagesState) setupTemplateRendering(r *gin.Engine) error {
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
	return nil
}

func (sts *PagesState) registerStatic(r *gin.Engine) {
	stscdir := "web/static"
	if sts.Config.StaticDirectory != "" {
		stscdir = sts.Config.StaticDirectory
	}
	r.StaticFile("/favicon.ico", stscdir+"/favicon.ico")
	r.Static("/static", stscdir)
}

func (sts *PagesState) registerRoutes(r *gin.Engine) {
	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusSeeOther, "/logs")
	})
	r.GET("/config", sts.handleConfig)

	r.GET("/logs", sts.handleLogs)
	r.GET("/logs/new", sts.handleNewLogForm)
	r.GET("/logs/:id", sts.handleEditLogForm)

	r.POST("/logs", sts.handleCreateLog)
	r.PUT("/logs/:id", sts.handleUpdateLog)
	r.DELETE("/logs/:id", sts.handleDeleteLog)

	r.POST("/config/metrics", sts.handleSaveMetrics)
	r.POST("/config/metrics/new", sts.handleCreateMetric)
	r.DELETE("/config/metrics/:key", sts.handleDeleteMetric)
	r.GET("/config/purge", sts.handlePurgePage)
	r.DELETE("/config/purge", sts.handlePurgeLogs)
	r.GET("/config/upload", sts.handleUploadPage)

	r.POST("/tutorial/log", sts.handleTutorialLog)
}
