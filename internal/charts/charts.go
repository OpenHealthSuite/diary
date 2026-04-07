package charts

import (
	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/foodlogs"
	"github.com/openhealthsuite/diary/internal/metrics"
	strgtyp "github.com/openhealthsuite/diary/internal/storage/types"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type ChartsState struct {
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
	sts := &ChartsState{
		Config:   config,
		Storage:  storage,
		FoodLogs: foodLogs,
		Metrics:  metrics,
	}

	sts.registerRoutes(r)

	return nil
}

func (sts *ChartsState) registerRoutes(r *gin.Engine) {
	r.GET("/charts/test.svg", sts.handleTestSvg)
}

func (sts *ChartsState) handleTestSvg(c *gin.Context) {
	p := plot.New()

	values := plotter.Values{3, 5, 2, 7}
	bars, _ := plotter.NewBarChart(values, vg.Points(20))
	p.Add(bars)

	c.Writer.Header().Set("Content-Type", "image/svg+xml")
	wt, _ := p.WriterTo(4*vg.Inch, 3*vg.Inch, "svg")
	wt.WriteTo(c.Writer)
}
