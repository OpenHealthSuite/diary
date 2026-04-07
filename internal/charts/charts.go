package charts

import (
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/auth"
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

func writePlotToResponse(c *gin.Context, p *plot.Plot) {
	c.Writer.Header().Set("Content-Type", "image/svg+xml")
	wt, _ := p.WriterTo(4*vg.Inch, 3*vg.Inch, "svg")
	wt.WriteTo(c.Writer)
}

func (sts *ChartsState) registerRoutes(r *gin.Engine) {
	r.GET("/charts/test.svg", sts.handleTestSvg)

	r.GET("/charts/:metric/total.svg", sts.handleMetricTotals)
}

func (sts *ChartsState) handleTestSvg(c *gin.Context) {
	p := plot.New()

	values := plotter.Values{3, 5, 2, 7}
	bars, _ := plotter.NewBarChart(values, vg.Points(20))
	p.Add(bars)

	writePlotToResponse(c, p)
}

func parseDateParam(c *gin.Context, param string, defaultVal time.Time) time.Time {
	dateStr := c.Query(param)
	if dateStr == "" {
		return time.Date(defaultVal.Year(), defaultVal.Month(), defaultVal.Day(), 0, 0, 0, 0, defaultVal.Location())
	}
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Date(defaultVal.Year(), defaultVal.Month(), defaultVal.Day(), 0, 0, 0, 0, defaultVal.Location())
	}
	return parsed
}

func (sts *ChartsState) handleMetricTotals(c *gin.Context) {
	metric := c.Param("metric")
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(403, err)
		return
	}
	stdt := parseDateParam(c, "startdate", time.Now().Add((time.Hour*24*6)*-1))
	eddt := parseDateParam(c, "enddate", time.Now())

	eddt = eddt.Add(time.Hour * 23)
	eddt = eddt.Add(time.Minute * 59)
	eddt = eddt.Add(time.Second * 59)

	lgs, err := sts.FoodLogs.GetLogsBetweenDateTimes(c, *userIdPtr, stdt, eddt)

	days := []string{}
	weekdays := []string{}
	totals := []float64{}

	for _, lg := range lgs {
		idx := slices.Index(days, lg.TimeStart.Format("2006-01-02"))
		if idx > -1 {
			totals[idx] += lg.Metrics[metric]
		} else {
			days = append(days, lg.TimeStart.Format("2006-01-02"))
			weekdays = append(weekdays, lg.TimeStart.Format("Mon"))
			totals = append(totals, lg.Metrics[metric])
		}
	}

	p := plot.New()

	values := plotter.Values{}

	for _, v := range totals {
		values = append(values, v)
	}

	bars, _ := plotter.NewBarChart(values, vg.Points(15))

	p.Add(bars)
	p.NominalX(weekdays...)
	p.Add(plotter.NewGrid())

	writePlotToResponse(c, p)
}
