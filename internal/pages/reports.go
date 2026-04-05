package pages

import (
	"net/http"

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
		"CurrentPath": "/reports",
		"Metrics":     *metrics,
		"TopMetric":   topMetric,
	}

	c.HTML(http.StatusOK, "pages/reports", data)
}
