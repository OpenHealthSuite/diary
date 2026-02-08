package server

import (
	"encoding/csv"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/openhealthsuite/diary/internal/server/generated"
	"github.com/openhealthsuite/diary/internal/storage/strggen"
)

// CreateFoodLog implements generated.ServerInterface.
func (g *ServerState) CreateFoodLog(c *gin.Context) {
	var req generated.CreateFoodLogEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid request body"})
		return
	}

	userId, ok := c.Get("userId")
	if !ok {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}

	// Generate a new UUID for the entry

	id := uuid.New()

	// Insert food log entry

	tsStart := pgtype.Timestamp{Time: req.Time.Start, Valid: true}
	tsEnd := pgtype.Timestamp{Time: req.Time.End, Valid: true}

	_, err := g.storage.GetQuerier().CreateFoodLogEntry(c, strggen.CreateFoodLogEntryParams{
		UserID:    userId.(string),
		ID:        id,
		Name:      req.Name,
		Labels:    req.Labels,
		TimeStart: tsStart,
		TimeEnd:   tsEnd,
	})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}

	// Insert metrics
	for k, v := range req.Metrics {
		_ = g.storage.GetQuerier().CreateFoodLogEntryMetric(c, strggen.CreateFoodLogEntryMetricParams{
			UserID:         userId.(string),
			FoodlogentryID: id,
			MetricKey:      k,
			MetricValue:    int32(v),
		})
	}

	c.String(200, id.String())
}

// DeleteFoodLog implements generated.ServerInterface.
func (g *ServerState) DeleteFoodLog(c *gin.Context, itemId string) {
	userId, ok := c.Get("userId")
	if !ok {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}

	id, err := uuid.Parse(itemId)
	if err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid itemId"})
		return
	}

	// Delete metrics first (optional, for referential integrity)
	_ = g.storage.GetQuerier().DeleteFoodLogEntryMetrics(c, struct {
		UserID         string
		FoodlogentryID uuid.UUID
	}{UserID: userId.(string), FoodlogentryID: id})

	// Delete entry
	err = g.storage.GetQuerier().DeleteFoodLogEntry(c, struct {
		UserID string
		ID     uuid.UUID
	}{UserID: userId.(string), ID: id})
	if err != nil {
		c.JSON(404, generated.Error{Code: 404, Message: "Food log entry not found"})
		return
	}
	c.Status(204)
}

// ExportFoodLogs implements generated.ServerInterface.
func (g *ServerState) ExportFoodLogs(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}

	// Get all entries
	entries, err := g.storage.GetQuerier().ExportFoodLogEntries(c, userId.(string))
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}
	// Get all metrics
	metrics, err := g.storage.GetQuerier().ExportFoodLogEntryMetrics(c, userId.(string))
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}
	// Map metrics by entry ID
	metricsMap := make(map[uuid.UUID]map[string]float32)
	for _, m := range metrics {
		if metricsMap[m.FoodlogentryID] == nil {
			metricsMap[m.FoodlogentryID] = make(map[string]float32)
		}
		metricsMap[m.FoodlogentryID][m.MetricKey] = float32(m.MetricValue)
	}

	// Write CSV response
	c.Header("Content-Disposition", "attachment; filename=food_logs.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"id", "name", "timeStart", "timeEnd", "metrics", "labels"})

	// Write rows
	for _, e := range entries {
		metricsJSON, _ := json.Marshal(metricsMap[e.ID])
		labelsJSON, _ := json.Marshal(e.Labels)

		writer.Write([]string{
			e.ID.String(),
			e.Name,
			e.TimeStart.Time.Format("2006-01-02T15:04:05Z"),
			e.TimeEnd.Time.Format("2006-01-02T15:04:05Z"),
			string(metricsJSON),
			string(labelsJSON),
		})
	}
}

// GetFoodLog implements generated.ServerInterface.
func (g *ServerState) GetFoodLog(c *gin.Context, itemId string) {
	userId, ok := c.Get("userId")
	if !ok {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}

	// Parse UUID
	id, err := uuid.Parse(itemId)
	if err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid itemId"})
		return
	}

	entry, err := g.storage.GetQuerier().GetFoodLogEntry(c, struct {
		UserID string
		ID     uuid.UUID
	}{UserID: userId.(string), ID: id})
	if err != nil {
		c.JSON(404, generated.Error{Code: 404, Message: "Food log entry not found"})
		return
	}

	metrics, err := g.storage.GetQuerier().GetFoodLogEntryMetrics(c, struct {
		UserID         string
		FoodlogentryID uuid.UUID
	}{UserID: userId.(string), FoodlogentryID: id})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}
	metricsMap := make(map[string]float32)
	for _, m := range metrics {
		metricsMap[m.MetricKey] = float32(m.MetricValue)
	}

	resp := generated.FoodLogEntry{
		Id:     entry.ID.String(),
		Name:   entry.Name,
		Labels: entry.Labels,
		Time: generated.TimeRange{
			Start: entry.TimeStart.Time,
			End:   entry.TimeEnd.Time,
		},
		Metrics: metricsMap,
	}
	c.JSON(200, resp)
}

// PurgeFoodLogs implements generated.ServerInterface.
func (g *ServerState) PurgeFoodLogs(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}

	// Delete all entries
	err := g.storage.GetQuerier().PurgeFoodLogEntries(c, userId.(string))
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}
	c.Status(204)
}

// QueryFoodLogs implements generated.ServerInterface.
func (g *ServerState) QueryFoodLogs(c *gin.Context, params generated.QueryFoodLogsParams) {
	userId, ok := c.Get("userId")
	if !ok {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}

	// Query entries
	entries, err := g.storage.GetQuerier().QueryFoodLogEntries(c, strggen.QueryFoodLogEntriesParams{
		UserID:    userId.(string),
		TimeStart: pgtype.Timestamp{Time: params.StartDate, Valid: true},
		TimeEnd:   pgtype.Timestamp{Time: params.EndDate, Valid: true},
	})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}

	// Collect entry IDs
	ids := make([]uuid.UUID, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.ID)
	}
	// Query all metrics for these entries
	metrics, err := g.storage.GetQuerier().QueryFoodLogEntryMetrics(c, strggen.QueryFoodLogEntryMetricsParams{
		UserID:  userId.(string),
		Column2: ids,
	})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}
	// Map metrics by entry ID
	metricsMap := make(map[uuid.UUID]map[string]float32)
	for _, m := range metrics {
		if metricsMap[m.FoodlogentryID] == nil {
			metricsMap[m.FoodlogentryID] = make(map[string]float32)
		}
		metricsMap[m.FoodlogentryID][m.MetricKey] = float32(m.MetricValue)
	}

	// Compose response
	resp := make([]generated.FoodLogEntry, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, generated.FoodLogEntry{
			Id:     e.ID.String(),
			Name:   e.Name,
			Labels: e.Labels,
			Time: generated.TimeRange{
				Start: e.TimeStart.Time,
				End:   e.TimeEnd.Time,
			},
			Metrics: metricsMap[e.ID],
		})
	}
	c.JSON(200, resp)
}

// UpdateFoodLog implements generated.ServerInterface.
func (g *ServerState) UpdateFoodLog(c *gin.Context, itemId string) {
	userId, ok := c.Get("userId")
	if !ok {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}

	var req generated.EditFoodLogEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid request body"})
		return
	}

	id, err := uuid.Parse(itemId)
	if err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid itemId"})
		return
	}

	// Fetch existing entry
	entry, err := g.storage.GetQuerier().GetFoodLogEntry(c, struct {
		UserID string
		ID     uuid.UUID
	}{UserID: userId.(string), ID: id})
	if err != nil {
		c.JSON(404, generated.Error{Code: 404, Message: "Food log entry not found"})
		return
	}

	// Prepare updated fields
	name := entry.Name
	if req.Name != nil {
		name = *req.Name
	}
	labels := entry.Labels
	if req.Labels != nil {
		labels = *req.Labels
	}
	tsStart := entry.TimeStart
	tsEnd := entry.TimeEnd
	if req.Time != nil {
		tsStart = pgtype.Timestamp{Time: req.Time.Start, Valid: true}
		tsEnd = pgtype.Timestamp{Time: req.Time.End, Valid: true}
	}

	// Update entry
	_, err = g.storage.GetQuerier().UpdateFoodLogEntry(c, strggen.UpdateFoodLogEntryParams{
		UserID:    userId.(string),
		ID:        id,
		Name:      name,
		Labels:    labels,
		TimeStart: tsStart,
		TimeEnd:   tsEnd,
	})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}

	// Update metrics if provided
	if req.Metrics != nil {
		// Delete old metrics
		_ = g.storage.GetQuerier().DeleteFoodLogEntryMetrics(c, struct {
			UserID         string
			FoodlogentryID uuid.UUID
		}{UserID: userId.(string), FoodlogentryID: id})
		// Insert new metrics
		for k, v := range *req.Metrics {
			_ = g.storage.GetQuerier().CreateFoodLogEntryMetric(c, strggen.CreateFoodLogEntryMetricParams{
				UserID:         userId.(string),
				FoodlogentryID: id,
				MetricKey:      k,
				MetricValue:    int32(v),
			})
		}
	}

	// Compose response
	metrics := make(map[string]float32)
	if req.Metrics != nil {
		for k, v := range *req.Metrics {
			metrics[k] = v
		}
	} else {
		// Fetch current metrics
		mlist, _ := g.storage.GetQuerier().GetFoodLogEntryMetrics(c, struct {
			UserID         string
			FoodlogentryID uuid.UUID
		}{UserID: userId.(string), FoodlogentryID: id})
		for _, m := range mlist {
			metrics[m.MetricKey] = float32(m.MetricValue)
		}
	}
	resp := generated.FoodLogEntry{
		Id:     id.String(),
		Name:   name,
		Labels: labels,
		Time: generated.TimeRange{
			Start: tsStart.Time,
			End:   tsEnd.Time,
		},
		Metrics: metrics,
	}
	c.JSON(200, resp)
}
