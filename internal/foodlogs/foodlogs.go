package foodlogs

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	storage "github.com/openhealthsuite/diary/internal/storage/types"
)

type UserFoodLog struct {
	ID        string
	UserID    string
	Name      string
	Metrics   map[string]float64
	TimeStart time.Time
	TimeEnd   time.Time
}

func FromStorage(stg storage.UserFoodlogentry) UserFoodLog {
	return UserFoodLog{
		ID:        stg.ID.String(),
		UserID:    stg.UserID,
		Name:      stg.Name,
		Metrics:   stg.Metrics,
		TimeStart: stg.TimeStart,
		TimeEnd:   stg.TimeEnd,
	}
}

func ToStorage(ufl UserFoodLog) (*storage.UserFoodlogentry, error) {
	uuid, err := uuid.Parse(ufl.ID)
	if err != nil {
		return nil, err
	}
	return &storage.UserFoodlogentry{
		ID:        uuid,
		UserID:    ufl.UserID,
		Name:      ufl.Name,
		Metrics:   ufl.Metrics,
		TimeStart: ufl.TimeStart,
		TimeEnd:   ufl.TimeEnd,
	}, nil
}

type FoodLogService interface {
	// Gets logs for a single day
	GetLogsForDate(c context.Context, userId string, date time.Time) ([]UserFoodLog, error)
}

type internalFoodLogService struct {
	storage storage.Storage
}

func NewFoodLogService(strg storage.Storage) (FoodLogService, error) {
	return &internalFoodLogService{
		storage: strg,
	}, nil
}

func (sts *internalFoodLogService) GetLogsForDate(c context.Context, userId string, date time.Time) ([]UserFoodLog, error) {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 1)

	entries, err := sts.storage.QueryFoodLogEntries(c, storage.QueryFoodLogEntriesParams{
		UserID:    userId,
		TimeStart: startDate,
		TimeEnd:   endDate,
	})
	if err != nil {
		return []UserFoodLog{}, nil
	}

	var result []UserFoodLog
	for _, entry := range entries {
		result = append(result, FromStorage(entry))
	}

	// Sort by time
	sort.Slice(result, func(i, j int) bool {
		return result[i].TimeStart.Before(result[j].TimeStart)
	})

	return result, nil
}
