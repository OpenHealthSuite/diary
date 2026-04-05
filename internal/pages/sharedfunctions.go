package pages

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/foodlogs"
	"github.com/openhealthsuite/diary/internal/metrics"
)

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
