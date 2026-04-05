package pages

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/openhealthsuite/diary/internal/auth"
	"github.com/openhealthsuite/diary/internal/storage/types"
)

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

	var topMetricTotal float64 = 0
	if topMetric != nil {
		for _, log := range logs {
			if val, ok := log.Metrics[topMetric.Key]; ok {
				topMetricTotal += val
			}
		}
	}

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

	c.Request.URL.RawQuery = "date=" + dateStr

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

	c.Request.URL.RawQuery = "date=" + dateStr

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

	entry, _ := sts.Storage.GetFoodLogEntry(c, types.GetFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})

	err = sts.Storage.DeleteFoodLogEntry(c, types.DeleteFoodLogEntryParams{
		UserID: userId,
		ID:     logUuid,
	})

	if entry.ID != uuid.Nil {
		c.Request.URL.RawQuery = "date=" + entry.TimeStart.Format("2006-01-02")
	}

	c.Header("HX-Redirect", "/logs?date="+entry.TimeStart.Format("2006-01-02"))
	sts.handleLogs(c)
}

func (sts *PagesState) handleTutorialLog(c *gin.Context) {
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
