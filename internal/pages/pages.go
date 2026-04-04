package pages

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/openhealthsuite/diary/internal/auth"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/foodlogs"
	"github.com/openhealthsuite/diary/internal/metrics"
	"github.com/openhealthsuite/diary/internal/storage/types"
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

	return nil
}

func parseDateParam(c *gin.Context, param string, defaultVal time.Time) time.Time {
	dateStr := c.Query(param)
	if dateStr == "" {
		return defaultVal
	}
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return defaultVal
	}
	return parsed
}

func parseDateTimeParam(c *gin.Context, param string, defaultVal time.Time) time.Time {
	dateStr := c.Query(param)
	if dateStr == "" {
		return defaultVal
	}
	parsed, err := time.Parse("2006-01-02T15:04:05", dateStr)
	if err != nil {
		return defaultVal
	}
	return parsed
}

func (sts *PagesState) handleLogs(c *gin.Context) {
	date := parseDateParam(c, "date", time.Now())
	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	logs, err := sts.FoodLogs.GetLogsForDate(c, *user_id, date)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metrics, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	topMetric := sts.Metrics.GetTopMetric(*metrics)

	// Calculate total for top metric
	var topMetricTotal float64 = 0
	if topMetric != nil {
		for _, log := range logs {
			if val, ok := log.Metrics[topMetric.Key]; ok {
				topMetricTotal += val
			}
		}
	}

	// Check for tutorial step
	var tutorialStep *int
	if stepStr := c.Query("tutorialStep"); stepStr != "" {
		step := 0
		fmt.Sscanf(stepStr, "%d", &step)
		tutorialStep = &step
	}

	data := gin.H{
		"CurrentPath":    "/logs",
		"CurrentDay":     date.Format("2006-01-02"),
		"PrevDay":        date.AddDate(0, 0, -1).Format("2006-01-02"),
		"NextDay":        date.AddDate(0, 0, 1).Format("2006-01-02"),
		"Logs":           logs,
		"Metrics":        *metrics,
		"TopMetric":      topMetric,
		"TopMetricTotal": topMetricTotal,
		"TutorialStep":   tutorialStep,
	}

	c.HTML(http.StatusOK, "pages/home", data)
}

func totalTopMetrics(topMetric *metrics.TopMetric, logs []foodlogs.UserFoodLog) float64 {
	var topMetricTotal float64 = 0
	if topMetric != nil {
		for _, log := range logs {
			if val, ok := log.Metrics[topMetric.Key]; ok {
				topMetricTotal += val
			}
		}
	}
	return topMetricTotal
}

func (sts *PagesState) handleNewLogForm(c *gin.Context) {
	date := parseDateTimeParam(c, "ts", time.Now())
	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	logs, err := sts.FoodLogs.GetLogsForDate(c, *user_id, date)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metrics, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, time.UTC)

	topMetric := sts.Metrics.GetTopMetric(*metrics)
	data := gin.H{
		"CurrentPath":    "/logs",
		"CurrentDay":     date.Format("2006-01-02"),
		"PrevDay":        date.AddDate(0, 0, -1).Format("2006-01-02"),
		"NextDay":        date.AddDate(0, 0, 1).Format("2006-01-02"),
		"Logs":           logs,
		"Metrics":        *metrics,
		"TopMetric":      topMetric,
		"TopMetricTotal": totalTopMetrics(topMetric, logs),
		"LogFormModal": gin.H{
			"IsEdit":     false,
			"Log":        nil,
			"LogDate":    date.Format("2006-01-02"),
			"LogTime":    date.Format("15:04"),
			"Metrics":    metrics,
			"Duration":   1,
			"CurrentDay": c.Query("date"),
		},
	}

	c.HTML(http.StatusOK, "pages/home", data)
}

func (sts *PagesState) handleEditLogForm(c *gin.Context) {
	logId := c.Param("id")
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(403, err)
		return
	}
	userId := *userIdPtr

	logUuid, err := uuid.Parse(logId)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid log ID")
		return
	}

	entry, err := sts.Storage.GetFoodLogEntry(c, types.GetFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})
	if err != nil {
		c.String(http.StatusNotFound, "Log not found")
		return
	}

	date := entry.TimeStart

	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	logs, err := sts.FoodLogs.GetLogsForDate(c, *user_id, date)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metrics, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	topMetric := sts.Metrics.GetTopMetric(*metrics)

	// Calculate total for top metric
	duration := entry.TimeEnd.Sub(entry.TimeStart).Minutes()
	if duration < 1 {
		duration = 1
	}

	data := gin.H{
		"CurrentPath":    "/logs",
		"CurrentDay":     date.Format("2006-01-02"),
		"PrevDay":        date.AddDate(0, 0, -1).Format("2006-01-02"),
		"NextDay":        date.AddDate(0, 0, 1).Format("2006-01-02"),
		"Logs":           logs,
		"Metrics":        *metrics,
		"TopMetric":      topMetric,
		"TopMetricTotal": totalTopMetrics(topMetric, logs),
		"LogFormModal": gin.H{
			"IsEdit":     true,
			"LogID":      entry.ID.String(),
			"LogName":    entry.Name,
			"LogDate":    entry.TimeStart.Format("2006-01-02"),
			"LogTime":    entry.TimeStart.Format("15:04"),
			"LogMetrics": entry.Metrics,
			"Metrics":    metrics,
			"Duration":   int(duration),
			"CurrentDay": entry.TimeStart.Format("2006-01-02"),
		},
	}

	c.HTML(http.StatusOK, "pages/home", data)
}

func (sts *PagesState) handleConfig(c *gin.Context) {
	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metrics, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	tutorialMode := c.Query("tutorial") == "1"

	data := gin.H{
		"CurrentPath":    "/config",
		"Metrics":        *metrics,
		"LogoutEndpoint": sts.Config.SignoutEndpoint,
		"TutorialMode":   tutorialMode,
	}

	c.HTML(http.StatusOK, "pages/config", data)
}

func (sts *PagesState) handleCreateLog(c *gin.Context) {
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(403, err)
		return
	}
	userId := *userIdPtr

	name := c.PostForm("name")
	dateStr := c.PostForm("date")
	timeStr := c.PostForm("time")
	durationStr := c.PostForm("duration")

	if name == "" {
		c.String(http.StatusBadRequest, "Name is required")
		return
	}

	startTime, _ := time.Parse("2006-01-02T15:04", dateStr+"T"+timeStr)
	duration := 1
	if d, err := strconv.Atoi(durationStr); err == nil {
		duration = d
	}
	endTime := startTime.Add(time.Duration(duration) * time.Minute)

	// Parse metrics from form
	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metricsConfig, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metrics := make(map[string]float64)
	for key := range *metricsConfig {
		if val := c.PostForm("metric_" + key); val != "" {
			var f float64
			if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
				metrics[key] = float64(f)
			}
		}
	}

	// Create the entry
	_, err = sts.Storage.CreateFoodLogEntry(c, types.CreateFoodLogEntryParams{
		UserID:    userId,
		Name:      name,
		Labels:    []string{},
		TimeStart: startTime,
		TimeEnd:   endTime,
		Metrics:   metrics,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create log")
		return
	}

	// Set the date for the logs partial
	c.Request.URL.RawQuery = "date=" + dateStr

	// Return updated logs list
	c.Header("HX-Redirect", "/logs?date="+dateStr)
	sts.handleLogs(c)
}

func (sts *PagesState) handleUpdateLog(c *gin.Context) {
	logId := c.Param("id")
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(403, err)
		return
	}
	userId := *userIdPtr

	logUuid, err := uuid.Parse(logId)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid log ID")
		return
	}

	name := c.PostForm("name")
	dateStr := c.PostForm("date")
	timeStr := c.PostForm("time")
	durationStr := c.PostForm("duration")

	if name == "" {
		c.String(http.StatusBadRequest, "Name is required")
		return
	}

	startTime, _ := time.Parse("2006-01-02T15:04", dateStr+"T"+timeStr)
	duration := 1
	if d, err := strconv.Atoi(durationStr); err == nil {
		duration = d
	}
	endTime := startTime.Add(time.Duration(duration) * time.Minute)

	// Parse metrics from form
	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metricsConfig, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metrics := make(map[string]float64)
	for key := range *metricsConfig {
		if val := c.PostForm("metric_" + key); val != "" {
			var f float64
			if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
				metrics[key] = float64(f)
			}
		}
	}

	// Update entry
	err = sts.Storage.UpdateFoodLogEntry(c, types.UpdateFoodLogEntryParams{
		UserID:    userId,
		ID:        logUuid,
		Name:      name,
		Labels:    []string{},
		TimeStart: startTime,
		TimeEnd:   endTime,
		Metrics:   metrics,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update log")
		return
	}

	// Set the date for the logs partial
	c.Request.URL.RawQuery = "date=" + dateStr

	// Return updated logs list
	c.Header("HX-Redirect", "/logs?date="+dateStr)
	sts.handleLogs(c)
}

func (sts *PagesState) handleDeleteLog(c *gin.Context) {
	logId := c.Param("id")
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(403, err)
		return
	}
	userId := *userIdPtr

	logUuid, err := uuid.Parse(logId)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid log ID")
		return
	}

	// Get the log first to know what date to refresh
	entry, _ := sts.Storage.GetFoodLogEntry(c, types.GetFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})

	// Delete metrics first
	err = sts.Storage.DeleteFoodLogEntry(c, types.DeleteFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})

	// Set the date for the logs partial
	if entry.ID != uuid.Nil {
		c.Request.URL.RawQuery = "date=" + entry.TimeStart.Format("2006-01-02")
	}

	// Return updated logs list
	c.Header("HX-Redirect", "/logs?date="+entry.TimeStart.Format("2006-01-02"))
	sts.handleLogs(c)
}

// Metrics Config Handlers

func (sts *PagesState) handleSaveMetrics(c *gin.Context) {
	key := c.PostForm("key")
	label := c.PostForm("label")
	tutorialMode := c.Query("tutorial") == "1"

	if key == "" || label == "" {
		if !tutorialMode {
			c.Header("HX-Redirect", "/config")
		}
		sts.handleConfig(c)
		return
	}

	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metricsPrt, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	metrics := *metricsPrt

	if cfg, exists := metrics[key]; exists {
		cfg.Label = label
		metrics[key] = cfg
	}

	sts.Metrics.UpdateUserMetrics(c, *user_id, metrics)
	if !tutorialMode {
		c.Header("HX-Redirect", "/config")
	}
	sts.handleConfig(c)
}

func (sts *PagesState) handleCreateMetric(c *gin.Context) {
	newMetricLabel := c.PostForm("new_metric")
	tutorialMode := c.Query("tutorial") == "1"

	if newMetricLabel == "" {
		if !tutorialMode {
			c.Header("HX-Redirect", "/config")
		}
		sts.handleConfig(c)
		return
	}
	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metricsPrt, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	metri := *metricsPrt

	// Generate key from label
	key := strings.ToLower(strings.ReplaceAll(newMetricLabel, " ", "_"))

	// Find max priority
	maxPriority := 0
	for _, m := range metri {
		if m.Priority > maxPriority {
			maxPriority = m.Priority
		}
	}

	metri[key] = metrics.MetricConfig{
		Label:    newMetricLabel,
		Priority: maxPriority + 1,
	}

	sts.Metrics.UpdateUserMetrics(c, *user_id, metri)
	if !tutorialMode {
		c.Header("HX-Redirect", "/config")
	}
	sts.handleConfig(c)
}

func (sts *PagesState) handleDeleteMetric(c *gin.Context) {
	key := c.Param("key")
	tutorialMode := c.Query("tutorial") == "1"

	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metricsPrt, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	metrics := *metricsPrt

	delete(metrics, key)

	sts.Metrics.UpdateUserMetrics(c, *user_id, metrics)
	if !tutorialMode {
		c.Header("HX-Redirect", "/config")
	}
	sts.handleConfig(c)
}

// Purge and Upload Pages

func (sts *PagesState) handlePurgePage(c *gin.Context) {
	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metricsPrt, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	metrics := *metricsPrt

	data := gin.H{
		"CurrentPath":    "/config",
		"Metrics":        metrics,
		"LogoutEndpoint": sts.Config.SignoutEndpoint,
		"PurgeModal":     true,
	}

	c.HTML(http.StatusOK, "pages/config", data)
}

func (sts *PagesState) handlePurgeLogs(c *gin.Context) {
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(403, err)
		return
	}
	userId := *userIdPtr

	err = sts.Storage.PurgeFoodLogEntries(c, userId)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to purge logs")
		return
	}

	c.Header("HX-Redirect", "/config")
	sts.handleConfig(c)
}

func (sts *PagesState) handleUploadPage(c *gin.Context) {
	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metricsPrt, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	metrics := *metricsPrt

	data := gin.H{
		"CurrentPath":    "/config",
		"Metrics":        metrics,
		"LogoutEndpoint": sts.Config.SignoutEndpoint,
		"UploadModal":    true,
	}

	c.HTML(http.StatusOK, "pages/config", data)
}

// Tutorial Handler

func (sts *PagesState) handleTutorialLog(c *gin.Context) {
	// Create the log entry and redirect to tutorial step 3
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(403, err)
		return
	}
	userId := *userIdPtr

	name := c.PostForm("name")
	dateStr := c.PostForm("date")
	timeStr := c.PostForm("time")
	durationStr := c.PostForm("duration")

	if name == "" {
		c.Redirect(http.StatusSeeOther, "/logs?tutorialStep=2")
		return
	}

	startTime, _ := time.Parse("2006-01-02T15:04", dateStr+"T"+timeStr)
	duration := 1
	if d, err := strconv.Atoi(durationStr); err == nil {
		duration = d
	}
	endTime := startTime.Add(time.Duration(duration) * time.Minute)

	// Parse metrics from form
	user_id, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metricsConfig, err := sts.Metrics.GetUserMetrics(c, *user_id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	metrics := make(map[string]float64)
	for key := range *metricsConfig {
		if val := c.PostForm("metric_" + key); val != "" {
			var f float64
			if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
				metrics[key] = float64(f)
			}
		}
	}

	// Create the entry
	_, err = sts.Storage.CreateFoodLogEntry(c, types.CreateFoodLogEntryParams{
		UserID:    userId,
		Name:      name,
		Labels:    []string{},
		TimeStart: startTime,
		TimeEnd:   endTime,
		Metrics:   metrics,
	})
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/logs?tutorialStep=2")
		return
	}

	c.Redirect(http.StatusSeeOther, "/logs?tutorialStep=3")
}
