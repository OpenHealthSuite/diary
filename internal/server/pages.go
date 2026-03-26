package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/openhealthsuite/diary/internal/auth"
	"github.com/openhealthsuite/diary/internal/storage/types"
)

// MetricConfig represents a single metric configuration
type MetricConfig struct {
	Label    string `json:"label"`
	Priority int    `json:"priority"`
}

// MetricsConfig is a map of metric key to config
type MetricsConfig map[string]MetricConfig

// TopMetric represents the highest priority metric for display
type TopMetric struct {
	Key   string
	Label string
}

// LogDisplay represents a food log for template display
type LogDisplay struct {
	ID         string
	Name       string
	TimeStart  time.Time
	TimeEnd    time.Time
	TimeString string
	Metrics    map[string]float64
}

func (sts *DiaryServerState) getStorage() *ServerState {
	return sts.GeneratedInterface.(*ServerState)
}

func (sts *DiaryServerState) fetchMetricsConfig(c *gin.Context) (*MetricsConfig, error) {
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		return nil, err
	}
	userId := *userIdPtr
	cfg, err := sts.getStorage().storage.GetUserConfig(c, types.GetUserConfigParams{
		UserID: userId,
		ID:     "metrics",
	})
	if err != nil {
		// Return default metrics
		return &MetricsConfig{
			"calories": {Label: "Calories", Priority: 0},
		}, nil
	}

	var metrics MetricsConfig
	if err := json.Unmarshal(cfg.ConfigValue, &metrics); err != nil {
		return &MetricsConfig{
			"calories": {Label: "Calories", Priority: 0},
		}, nil
	}
	return &metrics, nil
}

func (sts *DiaryServerState) saveMetricsConfig(c *gin.Context, metrics MetricsConfig) error {
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		return err
	}
	userId := *userIdPtr
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return sts.getStorage().storage.StoreUserConfig(c, types.StoreUserConfigParams{
		UserID:      userId,
		ID:          "metrics",
		ConfigValue: data,
	})
}

func getTopMetric(metrics MetricsConfig) *TopMetric {
	if len(metrics) == 0 {
		return nil
	}

	type kv struct {
		Key string
		Val MetricConfig
	}

	var sorted []kv
	for k, v := range metrics {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Val.Priority < sorted[j].Val.Priority
	})

	return &TopMetric{
		Key:   sorted[0].Key,
		Label: sorted[0].Val.Label,
	}
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

func (sts *DiaryServerState) fetchLogsForDate(c *gin.Context, date time.Time) ([]LogDisplay, error) {
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		return nil, err
	}
	userId := *userIdPtr

	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 1)

	entries, err := sts.getStorage().storage.QueryFoodLogEntries(c, types.QueryFoodLogEntriesParams{
		UserID:    userId,
		TimeStart: startDate,
		TimeEnd:   endDate,
	})
	if err != nil {
		return []LogDisplay{}, nil
	}

	var result []LogDisplay
	for _, entry := range entries {
		result = append(result, LogDisplay{
			ID:         entry.ID.String(),
			Name:       entry.Name,
			TimeStart:  entry.TimeStart,
			TimeEnd:    entry.TimeEnd,
			TimeString: entry.TimeStart.Format("15:04:05"),
			Metrics:    entry.Metrics,
		})
	}

	// Sort by time
	sort.Slice(result, func(i, j int) bool {
		return result[i].TimeStart.Before(result[j].TimeStart)
	})

	return result, nil
}

// Page Handlers

func (sts *DiaryServerState) handleLogs(c *gin.Context) {
	date := parseDateParam(c, "date", time.Now())
	logs, err := sts.fetchLogsForDate(c, date)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metrics, err := sts.fetchMetricsConfig(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	topMetric := getTopMetric(*metrics)

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
		"CurrentPath":    "/",
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

func (sts *DiaryServerState) handleNewLogForm(c *gin.Context) {
	date := parseDateParam(c, "date", time.Now())
	logs, err := sts.fetchLogsForDate(c, date)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metrics, err := sts.fetchMetricsConfig(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	topMetric := getTopMetric(*metrics)

	// Calculate total for top metric
	var topMetricTotal float64 = 0
	if topMetric != nil {
		for _, log := range logs {
			if val, ok := log.Metrics[topMetric.Key]; ok {
				topMetricTotal += val
			}
		}
	}

	// Set time to current time but with the specified date
	now := time.Now()
	date = time.Date(date.Year(), date.Month(), date.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)

	data := gin.H{
		"CurrentPath":    "/",
		"CurrentDay":     date.Format("2006-01-02"),
		"PrevDay":        date.AddDate(0, 0, -1).Format("2006-01-02"),
		"NextDay":        date.AddDate(0, 0, 1).Format("2006-01-02"),
		"Logs":           logs,
		"Metrics":        *metrics,
		"TopMetric":      topMetric,
		"TopMetricTotal": topMetricTotal,
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

func (sts *DiaryServerState) handleEditLogForm(c *gin.Context) {
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

	entry, err := sts.getStorage().storage.GetFoodLogEntry(c, types.GetFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})
	if err != nil {
		c.String(http.StatusNotFound, "Log not found")
		return
	}

	date := entry.TimeStart
	logs, err := sts.fetchLogsForDate(c, date)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	metrics, err := sts.fetchMetricsConfig(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	topMetric := getTopMetric(*metrics)

	// Calculate total for top metric
	var topMetricTotal float64 = 0
	if topMetric != nil {
		for _, log := range logs {
			if val, ok := log.Metrics[topMetric.Key]; ok {
				topMetricTotal += val
			}
		}
	}

	duration := entry.TimeEnd.Sub(entry.TimeStart).Minutes()
	if duration < 1 {
		duration = 1
	}

	data := gin.H{
		"CurrentPath":    "/",
		"CurrentDay":     date.Format("2006-01-02"),
		"PrevDay":        date.AddDate(0, 0, -1).Format("2006-01-02"),
		"NextDay":        date.AddDate(0, 0, 1).Format("2006-01-02"),
		"Logs":           logs,
		"Metrics":        *metrics,
		"TopMetric":      topMetric,
		"TopMetricTotal": topMetricTotal,
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

func (sts *DiaryServerState) handleConfig(c *gin.Context) {
	metrics, err := sts.fetchMetricsConfig(c)
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

func (sts *DiaryServerState) handleCreateLog(c *gin.Context) {
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
	metricsConfig, err := sts.fetchMetricsConfig(c)
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
	_, err = sts.getStorage().storage.CreateFoodLogEntry(c, types.CreateFoodLogEntryParams{
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

func (sts *DiaryServerState) handleUpdateLog(c *gin.Context) {
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
	metricsConfig, err := sts.fetchMetricsConfig(c)
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
	err = sts.getStorage().storage.UpdateFoodLogEntry(c, types.UpdateFoodLogEntryParams{
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

func (sts *DiaryServerState) handleDeleteLog(c *gin.Context) {
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
	entry, _ := sts.getStorage().storage.GetFoodLogEntry(c, types.GetFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})

	// Delete metrics first
	err = sts.getStorage().storage.DeleteFoodLogEntry(c, types.DeleteFoodLogEntryParams{
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

func (sts *DiaryServerState) handleSaveMetrics(c *gin.Context) {
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

	metricsPrt, err := sts.fetchMetricsConfig(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	metrics := *metricsPrt

	if cfg, exists := metrics[key]; exists {
		cfg.Label = label
		metrics[key] = cfg
	}

	sts.saveMetricsConfig(c, metrics)
	if !tutorialMode {
		c.Header("HX-Redirect", "/config")
	}
	sts.handleConfig(c)
}

func (sts *DiaryServerState) handleCreateMetric(c *gin.Context) {
	newMetricLabel := c.PostForm("new_metric")
	tutorialMode := c.Query("tutorial") == "1"

	if newMetricLabel == "" {
		if !tutorialMode {
			c.Header("HX-Redirect", "/config")
		}
		sts.handleConfig(c)
		return
	}

	metricsPrt, err := sts.fetchMetricsConfig(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	metrics := *metricsPrt

	// Generate key from label
	key := strings.ToLower(strings.ReplaceAll(newMetricLabel, " ", "_"))

	// Find max priority
	maxPriority := 0
	for _, m := range metrics {
		if m.Priority > maxPriority {
			maxPriority = m.Priority
		}
	}

	metrics[key] = MetricConfig{
		Label:    newMetricLabel,
		Priority: maxPriority + 1,
	}

	sts.saveMetricsConfig(c, metrics)
	if !tutorialMode {
		c.Header("HX-Redirect", "/config")
	}
	sts.handleConfig(c)
}

func (sts *DiaryServerState) handleDeleteMetric(c *gin.Context) {
	key := c.Param("key")
	tutorialMode := c.Query("tutorial") == "1"

	metricsPrt, err := sts.fetchMetricsConfig(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	metrics := *metricsPrt

	delete(metrics, key)

	sts.saveMetricsConfig(c, metrics)
	if !tutorialMode {
		c.Header("HX-Redirect", "/config")
	}
	sts.handleConfig(c)
}

// Purge and Upload Pages

func (sts *DiaryServerState) handlePurgePage(c *gin.Context) {
	metricsPrt, err := sts.fetchMetricsConfig(c)
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

func (sts *DiaryServerState) handlePurgeLogs(c *gin.Context) {
	userIdPtr, err := auth.GetUserId(c)
	if err != nil {
		c.AbortWithError(403, err)
		return
	}
	userId := *userIdPtr

	err = sts.getStorage().storage.PurgeFoodLogEntries(c, userId)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to purge logs")
		return
	}

	c.Header("HX-Redirect", "/config")
	sts.handleConfig(c)
}

func (sts *DiaryServerState) handleUploadPage(c *gin.Context) {
	metricsPrt, err := sts.fetchMetricsConfig(c)
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

func (sts *DiaryServerState) handleTutorialLog(c *gin.Context) {
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
	metricsConfig, err := sts.fetchMetricsConfig(c)
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
	_, err = sts.getStorage().storage.CreateFoodLogEntry(c, types.CreateFoodLogEntryParams{
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
