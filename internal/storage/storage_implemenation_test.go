package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openhealthsuite/diary/internal/config"
	pgstr "github.com/openhealthsuite/diary/internal/storage/postgres"
	sqlite "github.com/openhealthsuite/diary/internal/storage/sqlite3"
	"github.com/openhealthsuite/diary/internal/storage/types"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(ctx context.Context, t *testing.T) (types.Storage, func()) {
	dbName := "test"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"docker.io/postgres:18",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Error().Err(err).Msg("error starting container")
		t.FailNow()
	}
	cfg := config.ServerConfiguration{
		PostgresConnectionString: postgresContainer.MustConnectionString(ctx),
	}

	strg, err := pgstr.SetupPostgres(&cfg)

	if err != nil {
		log.Error().Err(err).Msg("error setting up storage")
		t.FailNow()
	}
	return strg, func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed to terminate container")
		}
	}
}

func setupSqlite(ctx context.Context, t *testing.T) (types.Storage, func()) {

	cfg := config.ServerConfiguration{
		PostgresConnectionString: ":memory:",
	}

	strg, err := sqlite.SetupSqlite(&cfg)

	if err != nil {
		log.Error().Err(err).Msg("error setting up storage")
		t.FailNow()
	}
	return strg, func() {

	}
}

func Test_StorageImplementations(xt *testing.T) {
	ctx := context.Background()

	impl := []struct {
		name  string
		strg  types.Storage
		clnup func()
	}{}
	pg, pgc := setupPostgres(ctx, xt)
	impl = append(impl, struct {
		name  string
		strg  types.Storage
		clnup func()
	}{"postgres", pg, pgc})
	sql, sqlc := setupSqlite(ctx, xt)
	impl = append(impl, struct {
		name  string
		strg  types.Storage
		clnup func()
	}{"sqlite", sql, sqlc})

	for _, imp := range impl {
		strg := imp.strg
		defer imp.clnup()
		xt.Run(imp.name, func(t *testing.T) {
			t.Run("Can get test data", func(t *testing.T) {
				res, err := strg.GetTestData(ctx)
				require.NoError(t, err)
				assert.Equal(t, "Database is Live", *res)
			})

			t.Run("CRUD of log", func(t *testing.T) {
				userid := "jimbloggs"
				ts := mustParseTime("2006-01-02 15:04:05", "2020-11-12 08:00:00")
				te := mustParseTime("2006-01-02 15:04:05", "2020-11-12 09:12:00")
				id, err := strg.CreateFoodLogEntry(ctx, types.CreateFoodLogEntryParams{
					UserID:    userid,
					Name:      "My Food Log",
					Labels:    []string{},
					TimeStart: ts,
					TimeEnd:   te,
					Metrics: map[string]float64{
						"cals": 123,
						"hap":  456,
					},
				})
				require.NoError(t, err)

				strd, err := strg.GetFoodLogEntry(ctx, types.GetFoodLogEntryParams{
					UserID: userid,
					ID:     *id,
				})

				require.NoError(t, err)
				assert.Equal(t, "My Food Log", strd.Name)
				assert.Equal(t, ts, strd.TimeStart)
				assert.Equal(t, te, strd.TimeEnd)
				assert.Equal(t, float64(123), strd.Metrics["cals"])
				assert.Equal(t, float64(456), strd.Metrics["hap"])

				err = strg.UpdateFoodLogEntry(ctx, types.UpdateFoodLogEntryParams{
					ID:     *id,
					UserID: userid,
					Name:   "My Updated Name",
					Labels: []string{},
					Metrics: map[string]float64{
						"cals": 765,
					},
					TimeStart: ts.Add(time.Hour),
					TimeEnd:   te.Add(2 * time.Hour),
				})
				require.NoError(t, err)

				strd, err = strg.GetFoodLogEntry(ctx, types.GetFoodLogEntryParams{
					UserID: userid,
					ID:     *id,
				})

				require.NoError(t, err)
				assert.Equal(t, "My Updated Name", strd.Name)
				assert.Equal(t, ts.Add(time.Hour), strd.TimeStart)
				assert.Equal(t, te.Add(2*time.Hour), strd.TimeEnd)
				assert.Equal(t, float64(765), strd.Metrics["cals"])
				assert.Equal(t, float64(0), strd.Metrics["hap"])

				err = strg.DeleteFoodLogEntry(ctx, types.DeleteFoodLogEntryParams{
					UserID: userid,
					ID:     *id,
				})
				require.NoError(t, err)

				strd, err = strg.GetFoodLogEntry(ctx, types.GetFoodLogEntryParams{
					UserID: userid,
					ID:     *id,
				})

				assert.Error(t, err)
			})

			t.Run("Test export", func(t *testing.T) {
				userid := "exportbloggs"
				ts := mustParseTime("2006-01-02 15:04:05", "2020-11-12 08:00:00")
				te := mustParseTime("2006-01-02 15:04:05", "2020-11-12 09:12:00")
				id1, err := strg.CreateFoodLogEntry(ctx, types.CreateFoodLogEntryParams{
					UserID:    userid,
					Name:      "My Food Log 1",
					Labels:    []string{},
					TimeStart: ts,
					TimeEnd:   te,
					Metrics: map[string]float64{
						"cals": 123,
						"hap":  456,
					},
				})
				require.NoError(t, err)
				id2, err := strg.CreateFoodLogEntry(ctx, types.CreateFoodLogEntryParams{
					UserID:    userid,
					Name:      "My Food Log 2",
					Labels:    []string{},
					TimeStart: ts.Add(time.Hour),
					TimeEnd:   te.Add(time.Hour),
					Metrics: map[string]float64{
						"cals": 234,
						"hap":  567,
					},
				})
				require.NoError(t, err)
				id3, err := strg.CreateFoodLogEntry(ctx, types.CreateFoodLogEntryParams{
					UserID:    userid,
					Name:      "My Food Log 3",
					Labels:    []string{},
					TimeStart: ts.Add(time.Hour * 2),
					TimeEnd:   te.Add(time.Hour * 2),
					Metrics: map[string]float64{
						"cals": 345,
						"hap":  678,
					},
				})
				require.NoError(t, err)

				exprt, err := strg.ExportFoodLogEntries(ctx, userid)

				require.NoError(t, err)
				require.Len(t, exprt, 3)
				assert.Equal(t, *id1, exprt[0].ID)
				assert.Equal(t, float64(123), exprt[0].Metrics["cals"])
				assert.Equal(t, *id2, exprt[1].ID)
				assert.Equal(t, float64(234), exprt[1].Metrics["cals"])
				assert.Equal(t, *id3, exprt[2].ID)
				assert.Equal(t, float64(678), exprt[2].Metrics["hap"])
			})

			t.Run("Test query", func(t *testing.T) {
				userid := "querybloggs"
				ts := mustParseTime("2006-01-02 15:04:05", "2020-11-12 08:00:00")
				te := mustParseTime("2006-01-02 15:04:05", "2020-11-12 09:12:00")
				id1, err := strg.CreateFoodLogEntry(ctx, types.CreateFoodLogEntryParams{
					UserID:    userid,
					Name:      "My Food Log 1",
					Labels:    []string{},
					TimeStart: ts,
					TimeEnd:   te,
					Metrics: map[string]float64{
						"cals": 123,
						"hap":  456,
					},
				})
				require.NoError(t, err)
				id2, err := strg.CreateFoodLogEntry(ctx, types.CreateFoodLogEntryParams{
					UserID:    userid,
					Name:      "My Food Log 2",
					Labels:    []string{},
					TimeStart: ts.Add(time.Hour),
					TimeEnd:   te.Add(time.Hour),
					Metrics: map[string]float64{
						"cals": 234,
						"hap":  567,
					},
				})
				require.NoError(t, err)
				id3, err := strg.CreateFoodLogEntry(ctx, types.CreateFoodLogEntryParams{
					UserID:    userid,
					Name:      "My Food Log 3",
					Labels:    []string{},
					TimeStart: ts.Add(time.Hour * 2),
					TimeEnd:   te.Add(time.Hour * 2),
					Metrics: map[string]float64{
						"cals": 345,
						"hap":  678,
					},
				})
				require.NoError(t, err)

				exprt, err := strg.QueryFoodLogEntries(ctx, types.QueryFoodLogEntriesParams{
					UserID:    userid,
					TimeStart: ts.Add(-time.Hour),
					TimeEnd:   te.Add(time.Hour * 3),
				})

				require.NoError(t, err)
				require.Len(t, exprt, 3)
				assert.Equal(t, *id1, exprt[0].ID)
				assert.Equal(t, float64(123), exprt[0].Metrics["cals"])
				assert.Equal(t, *id2, exprt[1].ID)
				assert.Equal(t, float64(234), exprt[1].Metrics["cals"])
				assert.Equal(t, *id3, exprt[2].ID)
				assert.Equal(t, float64(678), exprt[2].Metrics["hap"])

				exprt, err = strg.QueryFoodLogEntries(ctx, types.QueryFoodLogEntriesParams{
					UserID:    userid,
					TimeStart: exprt[1].TimeStart.Add(-time.Minute),
					TimeEnd:   exprt[1].TimeEnd.Add(time.Minute),
				})
				require.NoError(t, err)
				require.Len(t, exprt, 1)
				assert.Equal(t, *id2, exprt[0].ID)
				assert.Equal(t, float64(234), exprt[0].Metrics["cals"])
			})
		})
	}

}

func mustParseTime(fmt, str string) time.Time {
	t, err := time.Parse(fmt, str)
	if err != nil {
		panic(err)
	}
	return t
}
