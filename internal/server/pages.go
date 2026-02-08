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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/openhealthsuite/diary/internal/storage/strggen"
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
	Metrics    map[string]float32
}

func (sts *DiaryServerState) getUserId(c *gin.Context) string {
	userId, _ := c.Get("userId")
	return userId.(string)
}

func (sts *DiaryServerState) getStorage() *ServerState {
	return sts.GeneratedInterface.(*ServerState)
}

func (sts *DiaryServerState) fetchMetricsConfig(c *gin.Context) MetricsConfig {
	userId := sts.getUserId(c)
	cfg, err := sts.getStorage().storage.GetQuerier().GetUserConfig(c, strggen.GetUserConfigParams{
		UserID: userId,
		ID:     "metrics",
	})
	if err != nil {
		// Return default metrics
		return MetricsConfig{
			"calories": {Label: "Calories", Priority: 0},
		}
	}

	var metrics MetricsConfig
	if err := json.Unmarshal(cfg.ConfigValue, &metrics); err != nil {
		return MetricsConfig{
			"calories": {Label: "Calories", Priority: 0},
		}
	}
	return metrics
}

func (sts *DiaryServerState) saveMetricsConfig(c *gin.Context, metrics MetricsConfig) error {
	userId := sts.getUserId(c)
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return sts.getStorage().storage.GetQuerier().StoreUserConfig(c, strggen.StoreUserConfigParams{
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

func (sts *DiaryServerState) fetchLogsForDate(c *gin.Context, date time.Time) []LogDisplay {
	userId := sts.getUserId(c)

	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 1)

	entries, err := sts.getStorage().storage.GetQuerier().QueryFoodLogEntries(c, strggen.QueryFoodLogEntriesParams{
		UserID:    userId,
		TimeStart: pgtype.Timestamp{Time: startDate, Valid: true},
		TimeEnd:   pgtype.Timestamp{Time: endDate, Valid: true},
	})
	if err != nil {
		return []LogDisplay{}
	}

	// Collect entry IDs
	ids := make([]uuid.UUID, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.ID)
	}

	// Query all metrics for these entries
	metricsRows, _ := sts.getStorage().storage.GetQuerier().QueryFoodLogEntryMetrics(c, strggen.QueryFoodLogEntryMetricsParams{
		UserID:  userId,
		Column2: ids,
	})

	// Map metrics by entry ID
	metricsMap := make(map[uuid.UUID]map[string]float32)
	for _, m := range metricsRows {
		if metricsMap[m.FoodlogentryID] == nil {
			metricsMap[m.FoodlogentryID] = make(map[string]float32)
		}
		metricsMap[m.FoodlogentryID][m.MetricKey] = float32(m.MetricValue)
	}

	var result []LogDisplay
	for _, entry := range entries {
		metrics := metricsMap[entry.ID]
		if metrics == nil {
			metrics = make(map[string]float32)
		}

		result = append(result, LogDisplay{
			ID:         entry.ID.String(),
			Name:       entry.Name,
			TimeStart:  entry.TimeStart.Time,
			TimeEnd:    entry.TimeEnd.Time,
			TimeString: entry.TimeStart.Time.Format("15:04:05"),
			Metrics:    metrics,
		})
	}

	// Sort by time
	sort.Slice(result, func(i, j int) bool {
		return result[i].TimeStart.Before(result[j].TimeStart)
	})

	return result
}

// Page Handlers

func (sts *DiaryServerState) handleLogs(c *gin.Context) {
	date := parseDateParam(c, "date", time.Now())
	logs := sts.fetchLogsForDate(c, date)
	metrics := sts.fetchMetricsConfig(c)
	topMetric := getTopMetric(metrics)

	// Calculate total for top metric
	var topMetricTotal float32 = 0
	if topMetric != nil {
		for _, log := range logs {
			if val, ok := log.Metrics[topMetric.Key]; ok {
				topMetricTotal += val
			}
		}
	}

	data := gin.H{
		"CurrentPath":    "/",
		"CurrentDay":     date.Format("2006-01-02"),
		"PrevDay":        date.AddDate(0, 0, -1).Format("2006-01-02"),
		"NextDay":        date.AddDate(0, 0, 1).Format("2006-01-02"),
		"Logs":           logs,
		"Metrics":        metrics,
		"TopMetric":      topMetric,
		"TopMetricTotal": topMetricTotal,
	}

	c.HTML(http.StatusOK, "pages/home", data)
}

func (sts *DiaryServerState) handleNewLogForm(c *gin.Context) {
	date := parseDateParam(c, "date", time.Now())
	logs := sts.fetchLogsForDate(c, date)
	metrics := sts.fetchMetricsConfig(c)
	topMetric := getTopMetric(metrics)

	// Calculate total for top metric
	var topMetricTotal float32 = 0
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
		"Metrics":        metrics,
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

func (sts *DiaryServerState) handleConfig(c *gin.Context) {
	metrics := sts.fetchMetricsConfig(c)

	data := gin.H{
		"CurrentPath":    "/config",
		"Metrics":        metrics,
		"LogoutEndpoint": sts.Config.SingoutEndpoint,
	}

	c.HTML(http.StatusOK, "pages/config", data)
}

func (sts *DiaryServerState) handleEditLogForm(c *gin.Context) {
	logId := c.Param("id")
	userId := sts.getUserId(c)

	logUuid, err := uuid.Parse(logId)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid log ID")
		return
	}

	entry, err := sts.getStorage().storage.GetQuerier().GetFoodLogEntry(c, strggen.GetFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})
	if err != nil {
		c.String(http.StatusNotFound, "Log not found")
		return
	}

	// Fetch metrics for this entry
	metricsRows, _ := sts.getStorage().storage.GetQuerier().GetFoodLogEntryMetrics(c, strggen.GetFoodLogEntryMetricsParams{
		UserID:         userId,
		FoodlogentryID: logUuid,
	})

	logMetrics := make(map[string]float32)
	for _, m := range metricsRows {
		logMetrics[m.MetricKey] = float32(m.MetricValue)
	}

	duration := entry.TimeEnd.Time.Sub(entry.TimeStart.Time).Minutes()
	if duration < 1 {
		duration = 1
	}

	metrics := sts.fetchMetricsConfig(c)

	data := gin.H{
		"IsEdit":     true,
		"LogID":      entry.ID.String(),
		"LogName":    entry.Name,
		"LogDate":    entry.TimeStart.Time.Format("2006-01-02"),
		"LogTime":    entry.TimeStart.Time.Format("15:04"),
		"LogMetrics": logMetrics,
		"Metrics":    metrics,
		"Duration":   int(duration),
		"CurrentDay": entry.TimeStart.Time.Format("2006-01-02"),
	}

	c.HTML(http.StatusOK, "components/log_form_modal", data)
}

func (sts *DiaryServerState) handleCreateLog(c *gin.Context) {
	userId := sts.getUserId(c)

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
	metricsConfig := sts.fetchMetricsConfig(c)
	metrics := make(map[string]float32)
	for key := range metricsConfig {
		if val := c.PostForm("metric_" + key); val != "" {
			var f float64
			if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
				metrics[key] = float32(f)
			}
		}
	}

	id := uuid.New()

	// Create the entry
	_, err := sts.getStorage().storage.GetQuerier().CreateFoodLogEntry(c, strggen.CreateFoodLogEntryParams{
		UserID:    userId,
		ID:        id,
		Name:      name,
		Labels:    []string{},
		TimeStart: pgtype.Timestamp{Time: startTime, Valid: true},
		TimeEnd:   pgtype.Timestamp{Time: endTime, Valid: true},
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create log")
		return
	}

	// Insert metrics
	for k, v := range metrics {
		_ = sts.getStorage().storage.GetQuerier().CreateFoodLogEntryMetric(c, strggen.CreateFoodLogEntryMetricParams{
			UserID:         userId,
			FoodlogentryID: id,
			MetricKey:      k,
			MetricValue:    int32(v),
		})
	}

	// Set the date for the logs partial
	c.Request.URL.RawQuery = "date=" + dateStr

	// Return updated logs list
	sts.handleLogs(c)
}

func (sts *DiaryServerState) handleUpdateLog(c *gin.Context) {
	logId := c.Param("id")
	userId := sts.getUserId(c)

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
	metricsConfig := sts.fetchMetricsConfig(c)
	metrics := make(map[string]float32)
	for key := range metricsConfig {
		if val := c.PostForm("metric_" + key); val != "" {
			var f float64
			if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
				metrics[key] = float32(f)
			}
		}
	}

	// Update entry
	_, err = sts.getStorage().storage.GetQuerier().UpdateFoodLogEntry(c, strggen.UpdateFoodLogEntryParams{
		UserID:    userId,
		ID:        logUuid,
		Name:      name,
		Labels:    []string{},
		TimeStart: pgtype.Timestamp{Time: startTime, Valid: true},
		TimeEnd:   pgtype.Timestamp{Time: endTime, Valid: true},
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update log")
		return
	}

	// Delete old metrics and insert new ones
	_ = sts.getStorage().storage.GetQuerier().DeleteFoodLogEntryMetrics(c, strggen.DeleteFoodLogEntryMetricsParams{
		UserID:         userId,
		FoodlogentryID: logUuid,
	})
	for k, v := range metrics {
		_ = sts.getStorage().storage.GetQuerier().CreateFoodLogEntryMetric(c, strggen.CreateFoodLogEntryMetricParams{
			UserID:         userId,
			FoodlogentryID: logUuid,
			MetricKey:      k,
			MetricValue:    int32(v),
		})
	}

	// Set the date for the logs partial
	c.Request.URL.RawQuery = "date=" + dateStr

	// Return updated logs list
	sts.handleLogs(c)
}

func (sts *DiaryServerState) handleDeleteLog(c *gin.Context) {
	logId := c.Param("id")
	userId := sts.getUserId(c)

	logUuid, err := uuid.Parse(logId)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid log ID")
		return
	}

	// Get the log first to know what date to refresh
	entry, _ := sts.getStorage().storage.GetQuerier().GetFoodLogEntry(c, strggen.GetFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})

	// Delete metrics first
	_ = sts.getStorage().storage.GetQuerier().DeleteFoodLogEntryMetrics(c, strggen.DeleteFoodLogEntryMetricsParams{
		UserID:         userId,
		FoodlogentryID: logUuid,
	})

	// Delete entry
	err = sts.getStorage().storage.GetQuerier().DeleteFoodLogEntry(c, strggen.DeleteFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete log")
		return
	}

	// Set the date for the logs partial
	if entry.ID != uuid.Nil {
		c.Request.URL.RawQuery = "date=" + entry.TimeStart.Time.Format("2006-01-02")
	}

	// Return updated logs list
	sts.handleLogs(c)
}

// Metrics Config Handlers

func (sts *DiaryServerState) handleMetricsPartial(c *gin.Context) {
	metrics := sts.fetchMetricsConfig(c)

	data := gin.H{
		"Metrics": metrics,
	}

	c.HTML(http.StatusOK, "partials/config/metrics", data)
}

func (sts *DiaryServerState) handleSaveMetrics(c *gin.Context) {
	// This handles inline edits to metric labels
	metrics := sts.fetchMetricsConfig(c)

	// Update from form data
	for key := range metrics {
		if label := c.PostForm(key); label != "" {
			cfg := metrics[key]
			cfg.Label = label
			metrics[key] = cfg
		}
	}

	sts.saveMetricsConfig(c, metrics)
	sts.handleMetricsPartial(c)
}

func (sts *DiaryServerState) handleCreateMetric(c *gin.Context) {
	newMetricLabel := c.PostForm("new_metric")
	if newMetricLabel == "" {
		sts.handleMetricsPartial(c)
		return
	}

	metrics := sts.fetchMetricsConfig(c)

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
	sts.handleMetricsPartial(c)
}

func (sts *DiaryServerState) handleDeleteMetric(c *gin.Context) {
	key := c.Param("key")

	metrics := sts.fetchMetricsConfig(c)
	delete(metrics, key)

	sts.saveMetricsConfig(c, metrics)
	sts.handleMetricsPartial(c)
}

// Purge and Upload Modals

func (sts *DiaryServerState) handlePurgeModal(c *gin.Context) {
	c.HTML(http.StatusOK, "partials/config/purge_logs", nil)
}

func (sts *DiaryServerState) handlePurgeLogs(c *gin.Context) {
	userId := sts.getUserId(c)

	err := sts.getStorage().storage.GetQuerier().PurgeFoodLogEntries(c, userId)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to purge logs")
		return
	}

	c.String(http.StatusOK, "")
}

func (sts *DiaryServerState) handleUploadModal(c *gin.Context) {
	c.HTML(http.StatusOK, "partials/config/bulk_upload", nil)
}

// Tutorial Handlers

func (sts *DiaryServerState) handleTutorial(c *gin.Context) {
	metrics := sts.fetchMetricsConfig(c)

	data := gin.H{
		"Step":    0,
		"Metrics": metrics,
	}

	c.HTML(http.StatusOK, "components/tutorial", data)
}

func (sts *DiaryServerState) handleTutorialStep(c *gin.Context) {
	stepStr := c.Param("step")
	step := 0
	fmt.Sscanf(stepStr, "%d", &step)

	metrics := sts.fetchMetricsConfig(c)

	data := gin.H{
		"Step":    step,
		"Metrics": metrics,
	}

	c.HTML(http.StatusOK, "components/tutorial", data)
}
