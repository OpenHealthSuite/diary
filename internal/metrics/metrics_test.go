package metrics_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/metrics"
	"github.com/openhealthsuite/diary/internal/storage"
	"github.com/openhealthsuite/diary/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(xt *testing.T) {

	tsg, err := storage.NewStorage(&config.ServerConfiguration{
		SqliteFile: ":memory:",
	})
	require.NoError(xt, err)

	mcn, err := metrics.NewMetricsProvider(tsg)

	require.NoError(xt, err)

	xt.Run("Nothing stored - gets default", func(t *testing.T) {
		dlf, err := mcn.GetUserMetrics(t.Context(), uuid.NewString())
		require.NoError(t, err)
		assert.Equal(t, metrics.DefaultMetrics, *dlf)
	})

	xt.Run("Already stored - gets", func(t *testing.T) {
		userid := uuid.NewString()

		mtrcs := metrics.MetricsConfig{
			"watcher": metrics.MetricConfig{
				Label:    "WooHoo",
				Priority: 3,
			},
		}
		bts, err := json.Marshal(mtrcs)
		require.NoError(t, err)
		err = tsg.StoreUserConfig(t.Context(), types.StoreUserConfigParams{
			UserID:      userid,
			ID:          "metrics",
			ConfigValue: bts,
		})
		require.NoError(t, err)

		dlf, err := mcn.GetUserMetrics(t.Context(), userid)
		require.NoError(t, err)
		assert.Equal(t, mtrcs, *dlf)
	})

	xt.Run("Save and get", func(t *testing.T) {
		userid := uuid.NewString()

		mtrcs := metrics.MetricsConfig{
			"wooooooooo": metrics.MetricConfig{
				Label:    "Mandelin",
				Priority: 4,
			},
		}

		err := mcn.UpdateUserMetrics(t.Context(), userid, mtrcs)

		require.NoError(t, err)

		dlf, err := mcn.GetUserMetrics(t.Context(), userid)
		require.NoError(t, err)
		assert.Equal(t, mtrcs, *dlf)
	})

	xt.Run("Top Metric", func(t *testing.T) {

		mtrcs := metrics.MetricsConfig{
			"two": metrics.MetricConfig{
				Label:    "Two",
				Priority: 2,
			},
			"zero": metrics.MetricConfig{
				Label:    "Zero",
				Priority: 0,
			},
			"three": metrics.MetricConfig{
				Label:    "Three",
				Priority: 3,
			},
			"one": metrics.MetricConfig{
				Label:    "One",
				Priority: 1,
			},
		}

		tp := mcn.GetTopMetric(mtrcs)

		assert.Equal(t, "zero", tp.Key)
		assert.Equal(t, "Zero", tp.Label)
	})
}
