package api

import (
	"encoding/csv"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/openhealthsuite/diary/internal/api/generated"
	"github.com/openhealthsuite/diary/internal/auth"
	"github.com/openhealthsuite/diary/internal/storage/types"
)

// CreateFoodLog implements generated.ServerInterface.
func (g *ServerState) CreateFoodLog(c *gin.Context) {
	var req generated.CreateFoodLogEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid request body"})
		return
	}

	uidptr, err := auth.GetUserId(c)
	if err != nil {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}
	userId := *uidptr

	id, err := g.Storage.CreateFoodLogEntry(c, types.CreateFoodLogEntryParams{
		UserID:    userId,
		Name:      req.Name,
		Labels:    req.Labels,
		TimeStart: req.Time.Start,
		TimeEnd:   req.Time.End,
		Metrics:   req.Metrics,
	})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}

	c.String(200, id.String())
}

// DeleteFoodLog implements generated.ServerInterface.
func (g *ServerState) DeleteFoodLog(c *gin.Context, itemId string) {
	uidptr, err := auth.GetUserId(c)
	if err != nil {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}
	userId := *uidptr

	id, err := uuid.Parse(itemId)
	if err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid itemId"})
		return
	}

	// Delete entry
	err = g.Storage.DeleteFoodLogEntry(c, struct {
		UserID string
		ID     uuid.UUID
	}{UserID: userId, ID: id})
	if err != nil {
		c.JSON(404, generated.Error{Code: 404, Message: "Food log entry not found"})
		return
	}
	c.Status(204)
}

// ExportFoodLogs implements generated.ServerInterface.
func (g *ServerState) ExportFoodLogs(c *gin.Context) {
	uidptr, err := auth.GetUserId(c)
	if err != nil {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}
	userId := *uidptr

	// Get all entries
	entries, err := g.Storage.ExportFoodLogEntries(c, userId)
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
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
		metricsJSON, _ := json.Marshal(e.Metrics)
		labelsJSON, _ := json.Marshal(e.Labels)

		writer.Write([]string{
			e.ID.String(),
			e.Name,
			e.TimeStart.Format("2006-01-02T15:04:05Z"),
			e.TimeEnd.Format("2006-01-02T15:04:05Z"),
			string(metricsJSON),
			string(labelsJSON),
		})
	}
}

// GetFoodLog implements generated.ServerInterface.
func (g *ServerState) GetFoodLog(c *gin.Context, itemId string) {
	uidptr, err := auth.GetUserId(c)
	if err != nil {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}
	userId := *uidptr

	// Parse UUID
	id, err := uuid.Parse(itemId)
	if err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid itemId"})
		return
	}

	entry, err := g.Storage.GetFoodLogEntry(c, struct {
		UserID string
		ID     uuid.UUID
	}{UserID: userId, ID: id})
	if err != nil {
		c.JSON(404, generated.Error{Code: 404, Message: "Food log entry not found"})
		return
	}

	resp := generated.FoodLogEntry{
		Id:     entry.ID.String(),
		Name:   entry.Name,
		Labels: entry.Labels,
		Time: generated.TimeRange{
			Start: entry.TimeStart,
			End:   entry.TimeEnd,
		},
		Metrics: entry.Metrics,
	}
	c.JSON(200, resp)
}

// PurgeFoodLogs implements generated.ServerInterface.
func (g *ServerState) PurgeFoodLogs(c *gin.Context) {
	uidptr, err := auth.GetUserId(c)
	if err != nil {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}
	userId := *uidptr

	// Delete all entries
	err = g.Storage.PurgeFoodLogEntries(c, userId)
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}
	c.Status(204)
}

// QueryFoodLogs implements generated.ServerInterface.
func (g *ServerState) QueryFoodLogs(c *gin.Context, params generated.QueryFoodLogsParams) {
	uidptr, err := auth.GetUserId(c)
	if err != nil {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}
	userId := *uidptr

	// Query entries
	entries, err := g.Storage.QueryFoodLogEntries(c, types.QueryFoodLogEntriesParams{
		UserID:    userId,
		TimeStart: params.StartDate,
		TimeEnd:   params.EndDate,
	})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}

	// Compose response
	resp := make([]generated.FoodLogEntry, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, generated.FoodLogEntry{
			Id:     e.ID.String(),
			Name:   e.Name,
			Labels: e.Labels,
			Time: generated.TimeRange{
				Start: e.TimeStart,
				End:   e.TimeEnd,
			},
			Metrics: e.Metrics,
		})
	}
	c.JSON(200, resp)
}

// UpdateFoodLog implements generated.ServerInterface.
func (g *ServerState) UpdateFoodLog(c *gin.Context, itemId string) {
	uidptr, err := auth.GetUserId(c)
	if err != nil {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}
	userId := *uidptr

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
	entry, err := g.Storage.GetFoodLogEntry(c, struct {
		UserID string
		ID     uuid.UUID
	}{UserID: userId, ID: id})
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
		tsStart = req.Time.Start
		tsEnd = req.Time.End
	}

	metrics := entry.Metrics
	if req.Metrics != nil {
		metrics = *req.Metrics
	}

	// Update entry
	err = g.Storage.UpdateFoodLogEntry(c, types.UpdateFoodLogEntryParams{
		UserID:    userId,
		ID:        id,
		Name:      name,
		Labels:    labels,
		TimeStart: tsStart,
		TimeEnd:   tsEnd,
		Metrics:   metrics,
	})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}

	resp := generated.FoodLogEntry{
		Id:     id.String(),
		Name:   name,
		Labels: labels,
		Time: generated.TimeRange{
			Start: tsStart,
			End:   tsEnd,
		},
		Metrics: metrics,
	}
	c.JSON(200, resp)
}
