package types

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	GetTestData(ctx context.Context) (*string, error)

	CreateFoodLogEntry(ctx context.Context, arg CreateFoodLogEntryParams) (*uuid.UUID, error)

	UpdateFoodLogEntry(ctx context.Context, arg UpdateFoodLogEntryParams) error

	DeleteFoodLogEntry(ctx context.Context, arg DeleteFoodLogEntryParams) error

	ExportFoodLogEntries(ctx context.Context, userID string) ([]UserFoodlogentry, error)

	GetFoodLogEntry(ctx context.Context, arg GetFoodLogEntryParams) (*UserFoodlogentry, error)

	QueryFoodLogEntries(ctx context.Context, arg QueryFoodLogEntriesParams) ([]UserFoodlogentry, error)

	PurgeFoodLogEntries(ctx context.Context, userID string) error

	GetUserConfig(ctx context.Context, arg GetUserConfigParams) (*UserConfig, error)
	StoreUserConfig(ctx context.Context, arg StoreUserConfigParams) error
}

var (
	ErrNotFound = fmt.Errorf("not found")
)

type GetUserConfigParams struct {
	UserID string
	ID     string
}

type StoreUserConfigParams struct {
	UserID      string
	ID          string
	ConfigValue []byte
}

type CreateFoodLogEntryParams struct {
	UserID    string
	Name      string
	Labels    []string
	TimeStart time.Time
	TimeEnd   time.Time
	Metrics   map[string]float64
}

type DeleteFoodLogEntryParams struct {
	UserID string
	ID     uuid.UUID
}

type GetFoodLogEntryParams struct {
	UserID string
	ID     uuid.UUID
}

type QueryFoodLogEntriesParams struct {
	UserID    string
	TimeStart time.Time
	TimeEnd   time.Time
}

type UpdateFoodLogEntryParams struct {
	ID uuid.UUID

	UserID    string
	Name      string
	Labels    []string
	Metrics   map[string]float64
	TimeStart time.Time
	TimeEnd   time.Time
}

type UserConfig struct {
	UserID      string
	ID          string
	ConfigValue []byte
}

type UserFoodlogentry struct {
	ID uuid.UUID

	UserID    string
	Name      string
	Labels    []string
	Metrics   map[string]float64
	TimeStart time.Time
	TimeEnd   time.Time
}

type UserFoodlogentryMetric struct {
	UserID         string
	FoodlogentryID uuid.UUID
	MetricKey      string
	MetricValue    float64
}
