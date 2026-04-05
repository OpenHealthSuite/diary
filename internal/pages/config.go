package pages

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/auth"
	"github.com/openhealthsuite/diary/internal/metrics"
)

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

	key := strings.ToLower(strings.ReplaceAll(newMetricLabel, " ", "_"))

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
