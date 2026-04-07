package pages

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/auth"
)

func (sts *PagesState) handleReports(c *gin.Context) {
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
	topMetric := sts.Metrics.GetTopMetric(*metrics)

	data := gin.H{
		"CurrentPath":   "/reports",
		"Metrics":       *metrics,
		"TopMetric":     topMetric,
		"ThirtyDaysAgo": time.Now().Add((time.Hour * 24 * 29) * -1).Format("2006-01-02"),
	}

	c.HTML(http.StatusOK, "pages/reports", data)
}
