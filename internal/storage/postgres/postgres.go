package postgres

import (
	"context"
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/storage/postgres/sqlc"
	"github.com/openhealthsuite/diary/internal/storage/postgres/strggen"
	"github.com/openhealthsuite/diary/internal/storage/types"
	"github.com/rs/zerolog/log"
)

//go:generate sqlc generate

type PostgresStorage struct {
	Db      *pgxpool.Pool
	Querier strggen.Querier
}

// GetTestData implements [types.Storage].
func (p *PostgresStorage) GetTestData(ctx context.Context) (*string, error) {
	return p.Querier.GetTestData(ctx)
}

func rollbacker(tx pgx.Tx) {
	err := tx.Rollback(context.Background())
	if err != nil && err != pgx.ErrTxClosed {
		log.Error().Err(err).Msg("error rolling back transaction")
	}
}

// CreateFoodLogEntry implements [types.Storage].
func (p *PostgresStorage) CreateFoodLogEntry(ctx context.Context, arg types.CreateFoodLogEntryParams) (*uuid.UUID, error) {
	tx, err := p.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbacker(tx)

	q := strggen.New(tx)
	id, err := q.CreateFoodLogEntry(ctx, strggen.CreateFoodLogEntryParams{
		UserID:    arg.UserID,
		Name:      arg.Name,
		Labels:    arg.Labels,
		TimeStart: pgtype.Timestamp{Time: arg.TimeStart, Valid: true},
		TimeEnd:   pgtype.Timestamp{Time: arg.TimeEnd, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	for k, v := range arg.Metrics {
		err := q.CreateFoodLogEntryMetric(ctx, strggen.CreateFoodLogEntryMetricParams{
			UserID:         arg.UserID,
			FoodlogentryID: id,
			MetricKey:      k,
			MetricValue:    v,
		})
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// DeleteFoodLogEntry implements [types.Storage].
func (p *PostgresStorage) DeleteFoodLogEntry(ctx context.Context, arg types.DeleteFoodLogEntryParams) error {
	return p.Querier.DeleteFoodLogEntry(ctx, strggen.DeleteFoodLogEntryParams{
		UserID: arg.UserID,
		ID:     arg.ID,
	})
}

// ExportFoodLogEntries implements [types.Storage].
func (p *PostgresStorage) ExportFoodLogEntries(ctx context.Context, userID string) ([]types.UserFoodlogentry, error) {
	entrs, err := p.Querier.ExportFoodLogEntries(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := []types.UserFoodlogentry{}

	for _, rw := range entrs {
		vl := types.UserFoodlogentry{
			ID:        rw.ID,
			UserID:    rw.UserID,
			Name:      rw.Name,
			Labels:    rw.Labels,
			Metrics:   map[string]float64{},
			TimeStart: rw.TimeStart.Time,
			TimeEnd:   rw.TimeEnd.Time,
		}
		rwm := []rawQueryMetric{}

		err := json.Unmarshal(rw.Metrics, &rwm)
		if err != nil {
			return nil, err
		}
		for _, rm := range rwm {
			vl.Metrics[rm.Key] = rm.Value
		}
		res = append(res, vl)
	}

	return res, nil
}

// GetFoodLogEntry implements [types.Storage].
func (p *PostgresStorage) GetFoodLogEntry(ctx context.Context, arg types.GetFoodLogEntryParams) (*types.UserFoodlogentry, error) {
	tx, err := p.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbacker(tx)
	q := strggen.New(tx)

	rfl, err := q.GetFoodLogEntry(ctx, strggen.GetFoodLogEntryParams{
		UserID: arg.UserID,
		ID:     arg.ID,
	})
	if err == pgx.ErrNoRows {
		return nil, types.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	mtrcs, err := q.GetFoodLogEntryMetrics(ctx, strggen.GetFoodLogEntryMetricsParams{
		UserID:         arg.UserID,
		FoodlogentryID: arg.ID,
	})

	if err != nil {
		return nil, err
	}
	res := types.UserFoodlogentry{
		ID:        rfl.ID,
		UserID:    rfl.UserID,
		Name:      rfl.Name,
		Labels:    rfl.Labels,
		Metrics:   map[string]float64{},
		TimeStart: rfl.TimeStart.Time,
		TimeEnd:   rfl.TimeEnd.Time,
	}

	for _, m := range mtrcs {
		res.Metrics[m.MetricKey] = m.MetricValue
	}
	return &res, nil
}

// GetUserConfig implements [types.Storage].
func (p *PostgresStorage) GetUserConfig(ctx context.Context, arg types.GetUserConfigParams) (*types.UserConfig, error) {
	rw, err := p.Querier.GetUserConfig(ctx, strggen.GetUserConfigParams{
		UserID: arg.UserID,
		ID:     arg.ID,
	})
	if err != nil {
		return nil, err
	}
	return &types.UserConfig{
		UserID:      rw.UserID,
		ID:          rw.ID,
		ConfigValue: rw.ConfigValue,
	}, nil
}

// StoreUserConfig implements [types.Storage].
func (p *PostgresStorage) StoreUserConfig(ctx context.Context, arg types.StoreUserConfigParams) error {
	return p.Querier.StoreUserConfig(ctx, strggen.StoreUserConfigParams{
		UserID:      arg.UserID,
		ID:          arg.ID,
		ConfigValue: arg.ConfigValue,
	})
}

// PurgeFoodLogEntries implements [types.Storage].
func (p *PostgresStorage) PurgeFoodLogEntries(ctx context.Context, userID string) error {
	return p.Querier.PurgeFoodLogEntries(ctx, userID)
}

type rawQueryMetric struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

// QueryFoodLogEntries implements [types.Storage].
func (p *PostgresStorage) QueryFoodLogEntries(ctx context.Context, arg types.QueryFoodLogEntriesParams) ([]types.UserFoodlogentry, error) {
	entrs, err := p.Querier.QueryFoodLogEntries(ctx, strggen.QueryFoodLogEntriesParams{
		UserID:    arg.UserID,
		TimeStart: pgtype.Timestamp{Time: arg.TimeStart, Valid: true},
		TimeEnd:   pgtype.Timestamp{Time: arg.TimeEnd, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	res := []types.UserFoodlogentry{}

	for _, rw := range entrs {
		vl := types.UserFoodlogentry{
			ID:        rw.ID,
			UserID:    rw.UserID,
			Name:      rw.Name,
			Labels:    rw.Labels,
			Metrics:   map[string]float64{},
			TimeStart: rw.TimeStart.Time,
			TimeEnd:   rw.TimeEnd.Time,
		}
		rwm := []rawQueryMetric{}

		err := json.Unmarshal(rw.Metrics, &rwm)
		if err != nil {
			return nil, err
		}
		for _, rm := range rwm {
			vl.Metrics[rm.Key] = rm.Value
		}
		res = append(res, vl)
	}

	return res, nil
}

// UpdateFoodLogEntry implements [types.Storage].
func (p *PostgresStorage) UpdateFoodLogEntry(ctx context.Context, arg types.UpdateFoodLogEntryParams) error {
	tx, err := p.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbacker(tx)
	q := strggen.New(tx)

	_, err = q.UpdateFoodLogEntry(ctx, strggen.UpdateFoodLogEntryParams{
		ID:        arg.ID,
		UserID:    arg.UserID,
		Name:      arg.Name,
		Labels:    arg.Labels,
		TimeStart: pgtype.Timestamp{Time: arg.TimeStart, Valid: true},
		TimeEnd:   pgtype.Timestamp{Time: arg.TimeEnd, Valid: true},
	})
	if err == pgx.ErrNoRows {
		return types.ErrNotFound
	}
	if err != nil {
		return err
	}

	q.DeleteFoodLogEntryMetrics(ctx, strggen.DeleteFoodLogEntryMetricsParams{
		UserID:         arg.UserID,
		FoodlogentryID: arg.ID,
	})

	for k, v := range arg.Metrics {
		err := q.CreateFoodLogEntryMetric(ctx, strggen.CreateFoodLogEntryMetricParams{
			UserID:         arg.UserID,
			FoodlogentryID: arg.ID,
			MetricKey:      k,
			MetricValue:    v,
		})
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func SetupPostgres(cfg *config.ServerConfiguration) (types.Storage, error) {

	db, err := pgxpool.New(context.Background(), cfg.PostgresConnectionString)
	if err != nil {
		return nil, err
	}
	err = db.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	// Create migrations table
	_, err = db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS migrations (id varchar(255));")
	if err != nil {
		return nil, err
	}

	for _, migration := range getPgMigrations() {
		selectSQL := `
		SELECT id
		FROM migrations
		WHERE id = $1;
`
		ctx, cncl := context.WithTimeout(context.Background(), time.Second*5)
		defer cncl()
		rows := db.QueryRow(ctx, selectSQL, migration[0])
		if err != nil {
			return nil, err
		}

		if rows.Scan() != pgx.ErrNoRows {
			// already done, just continue
			log.Debug().Any("migrationindex", migration[0]).Msg("migration already done")
			continue
		}
		log.Debug().Any("migrationindex", migration[0]).Msg("migration being processed")
		_, err = db.Exec(context.Background(), migration[1])
		if err != nil {
			return nil, err
		}
		_, err = db.Exec(context.Background(), "INSERT INTO migrations (id) VALUES ($1);", migration[0])
		if err != nil {
			return nil, err
		}
		log.Debug().Any("migrationindex", migration[0]).Msg("migration successful")
	}

	log.Info().Msg("migrations complete")

	stg := PostgresStorage{
		Db:      db,
		Querier: strggen.New(db),
	}
	return &stg, nil
}

func getPgMigrations() [][]string {
	migrationfiles, err := sqlc.SqlFiles.ReadDir("schema")
	if err != nil {
		log.Fatal().Err(err).Msg("Fatal error reading migrations dir")
	}
	data := [][]string{}
	re := regexp.MustCompile(`^\d{5}$`)
	for _, fl := range migrationfiles {
		if !fl.IsDir() {
			fnm_prts := strings.Split(fl.Name(), "_")
			key := fnm_prts[0]
			if !re.MatchString(key) {
				log.Warn().Msgf("Malformed filename in migrations: %s", fl.Name())
				continue
			}
			migration, err := sqlc.SqlFiles.ReadFile("schema/" + fl.Name())
			if err != nil {
				log.Fatal().Err(err).Msgf("Fatal error reading file %s", fl.Name())
			}
			data = append(data, []string{key, string(migration)})
		}
	}
	slices.SortFunc(data, func(a []string, b []string) int {
		return strings.Compare(a[0], b[0])
	})
	return data
}
