package metrics

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/openhealthsuite/diary/internal/storage/types"
	storage "github.com/openhealthsuite/diary/internal/storage/types"
	"github.com/rs/zerolog/log"
)

// MetricConfig represents a single metric configuration
type MetricConfig struct {
	Label    string `json:"label"`
	Priority int    `json:"priority"`
}

// MetricsConfig is a map of metric key to config
type MetricsConfig map[string]MetricConfig

type MetricsProvider interface {
	GetUserMetrics(ctx context.Context, user_id string) (*MetricsConfig, error)
	UpdateUserMetrics(ctx context.Context, user_id string, metrics MetricsConfig) error
	GetTopMetric(metrics MetricsConfig) *TopMetric
}

type internalMetricsProvider struct {
	strg storage.Storage
}

func NewMetricsProvider(strg storage.Storage) (MetricsProvider, error) {
	return &internalMetricsProvider{
		strg: strg,
	}, nil
}

var DefaultMetrics = MetricsConfig{
	"calories": {Label: "Calories", Priority: 0},
}

// GetUserMetrics implements [MetricsProvider].
func (i *internalMetricsProvider) GetUserMetrics(ctx context.Context, user_id string) (*MetricsConfig, error) {
	cfg, err := i.strg.GetUserConfig(ctx, types.GetUserConfigParams{
		UserID: user_id,
		ID:     "metrics",
	})
	if err == storage.ErrNotFound {
		return &DefaultMetrics, nil
	}
	if err != nil {
		return nil, err
	}

	var metrics MetricsConfig
	if err := json.Unmarshal(cfg.ConfigValue, &metrics); err != nil {
		log.Error().Err(err).Msg("error unmarshalling stored metrics")
		return &DefaultMetrics, nil
	}
	return &metrics, nil
}

// UpdateUserMetrics implements [MetricsProvider].
func (i *internalMetricsProvider) UpdateUserMetrics(ctx context.Context, user_id string, metrics MetricsConfig) error {

	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return i.strg.StoreUserConfig(ctx, types.StoreUserConfigParams{
		UserID:      user_id,
		ID:          "metrics",
		ConfigValue: data,
	})
}

// TopMetric represents the highest priority metric for display
type TopMetric struct {
	Key   string
	Label string
}

func (i *internalMetricsProvider) GetTopMetric(metrics MetricsConfig) *TopMetric {
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
