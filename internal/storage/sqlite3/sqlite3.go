package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/storage/sqlite3/sqlc"
	"github.com/openhealthsuite/diary/internal/storage/sqlite3/strggen"
	"github.com/openhealthsuite/diary/internal/storage/types"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

//go:generate sqlc generate

type Sqlite3Storage struct {
	Db      *sql.DB
	Querier strggen.Querier
}

// GetTestData implements [types.Storage].
func (p *Sqlite3Storage) GetTestData(ctx context.Context) (*string, error) {
	return p.Querier.GetTestData(ctx)
}

func rollbacker(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil && err != sql.ErrTxDone {
		log.Error().Err(err).Msg("error rolling back transaction")
	}
}

// CreateFoodLogEntry implements [types.Storage].
func (p *Sqlite3Storage) CreateFoodLogEntry(ctx context.Context, arg types.CreateFoodLogEntryParams) (*uuid.UUID, error) {
	tx, err := p.Db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbacker(tx)

	q := strggen.New(tx)

	lbls, err := json.Marshal(arg.Labels)

	if err != nil {
		return nil, err
	}
	id, err := q.CreateFoodLogEntry(ctx, strggen.CreateFoodLogEntryParams{
		ID:        uuid.NewString(),
		UserID:    arg.UserID,
		Name:      arg.Name,
		Labels:    lbls,
		TimeStart: arg.TimeStart,
		TimeEnd:   arg.TimeEnd,
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	uuid := uuid.MustParse(id)
	return &uuid, nil
}

// DeleteFoodLogEntry implements [types.Storage].
func (p *Sqlite3Storage) DeleteFoodLogEntry(ctx context.Context, arg types.DeleteFoodLogEntryParams) error {
	return p.Querier.DeleteFoodLogEntry(ctx, strggen.DeleteFoodLogEntryParams{
		UserID: arg.UserID,
		ID:     arg.ID.String(),
	})
}

// ExportFoodLogEntries implements [types.Storage].
func (p *Sqlite3Storage) ExportFoodLogEntries(ctx context.Context, userID string) ([]types.UserFoodlogentry, error) {
	entrs, err := p.Querier.ExportFoodLogEntries(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := []types.UserFoodlogentry{}

	for _, rw := range entrs {
		uuid, err := uuid.Parse(rw.ID)
		if err != nil {
			return nil, err
		}

		vl := types.UserFoodlogentry{
			ID:        uuid,
			UserID:    rw.UserID,
			Name:      rw.Name,
			Labels:    []string{},
			Metrics:   map[string]float64{},
			TimeStart: rw.TimeStart,
			TimeEnd:   rw.TimeEnd,
		}
		rwm := []rawQueryMetric{}
		err = json.Unmarshal([]byte(rw.Metrics.(string)), &rwm)
		if err != nil {
			return nil, err
		}
		for _, rm := range rwm {
			vl.Metrics[rm.Key] = rm.Value
		}
		lbls := []string{}

		err = json.Unmarshal(rw.Labels, &lbls)
		if err != nil {
			return nil, err
		}
		vl.Labels = lbls
		res = append(res, vl)
	}

	return res, nil
}

// GetFoodLogEntry implements [types.Storage].
func (p *Sqlite3Storage) GetFoodLogEntry(ctx context.Context, arg types.GetFoodLogEntryParams) (*types.UserFoodlogentry, error) {
	tx, err := p.Db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbacker(tx)
	q := strggen.New(tx)

	rfl, err := q.GetFoodLogEntry(ctx, strggen.GetFoodLogEntryParams{
		UserID: arg.UserID,
		ID:     arg.ID.String(),
	})
	if err != nil {
		return nil, err
	}
	mtrcs, err := q.GetFoodLogEntryMetrics(ctx, strggen.GetFoodLogEntryMetricsParams{
		UserID:         arg.UserID,
		FoodlogentryID: arg.ID.String(),
	})

	if err != nil {
		return nil, err
	}
	res := types.UserFoodlogentry{
		ID:        arg.ID,
		UserID:    rfl.UserID,
		Name:      rfl.Name,
		Labels:    []string{},
		Metrics:   map[string]float64{},
		TimeStart: rfl.TimeStart,
		TimeEnd:   rfl.TimeEnd,
	}

	for _, m := range mtrcs {
		res.Metrics[m.MetricKey] = m.MetricValue
	}
	lbls := []string{}

	err = json.Unmarshal(rfl.Labels, &lbls)
	if err != nil {
		return nil, err
	}
	res.Labels = lbls
	return &res, nil
}

// GetUserConfig implements [types.Storage].
func (p *Sqlite3Storage) GetUserConfig(ctx context.Context, arg types.GetUserConfigParams) (*types.UserConfig, error) {
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
func (p *Sqlite3Storage) StoreUserConfig(ctx context.Context, arg types.StoreUserConfigParams) error {
	return p.Querier.StoreUserConfig(ctx, strggen.StoreUserConfigParams{
		UserID:      arg.UserID,
		ID:          arg.ID,
		ConfigValue: arg.ConfigValue,
	})
}

// PurgeFoodLogEntries implements [types.Storage].
func (p *Sqlite3Storage) PurgeFoodLogEntries(ctx context.Context, userID string) error {
	return p.Querier.PurgeFoodLogEntries(ctx, userID)
}

type rawQueryMetric struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

// QueryFoodLogEntries implements [types.Storage].
func (p *Sqlite3Storage) QueryFoodLogEntries(ctx context.Context, arg types.QueryFoodLogEntriesParams) ([]types.UserFoodlogentry, error) {
	entrs, err := p.Querier.QueryFoodLogEntries(ctx, strggen.QueryFoodLogEntriesParams{
		UserID:    arg.UserID,
		TimeStart: arg.TimeStart,
		TimeEnd:   arg.TimeEnd,
	})
	if err != nil {
		return nil, err
	}

	res := []types.UserFoodlogentry{}

	for _, rw := range entrs {
		vl := types.UserFoodlogentry{
			ID:        uuid.MustParse(rw.ID),
			UserID:    rw.UserID,
			Name:      rw.Name,
			Labels:    []string{},
			Metrics:   map[string]float64{},
			TimeStart: rw.TimeStart,
			TimeEnd:   rw.TimeEnd,
		}
		rwm := []rawQueryMetric{}

		err := json.Unmarshal([]byte(rw.Metrics.(string)), &rwm)
		if err != nil {
			return nil, err
		}
		for _, rm := range rwm {
			vl.Metrics[rm.Key] = rm.Value
		}
		lbls := []string{}

		err = json.Unmarshal(rw.Labels, &lbls)
		if err != nil {
			return nil, err
		}
		vl.Labels = lbls
		res = append(res, vl)
	}

	return res, nil
}

// UpdateFoodLogEntry implements [types.Storage].
func (p *Sqlite3Storage) UpdateFoodLogEntry(ctx context.Context, arg types.UpdateFoodLogEntryParams) error {
	tx, err := p.Db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbacker(tx)
	q := strggen.New(tx)

	_, err = q.GetFoodLogEntry(ctx, strggen.GetFoodLogEntryParams{
		UserID: arg.UserID,
		ID:     arg.ID.String(),
	})
	if err != nil {
		return err
	}

	lbls, err := json.Marshal(arg.Labels)

	if err != nil {
		return err
	}

	err = q.UpdateFoodLogEntry(ctx, strggen.UpdateFoodLogEntryParams{
		ID:        arg.ID.String(),
		UserID:    arg.UserID,
		Name:      arg.Name,
		Labels:    lbls,
		TimeStart: arg.TimeStart,
		TimeEnd:   arg.TimeEnd,
	})
	if err != nil {
		return err
	}

	q.DeleteFoodLogEntryMetrics(ctx, strggen.DeleteFoodLogEntryMetricsParams{
		UserID:         arg.UserID,
		FoodlogentryID: arg.ID.String(),
	})

	for k, v := range arg.Metrics {
		err := q.CreateFoodLogEntryMetric(ctx, strggen.CreateFoodLogEntryMetricParams{
			UserID:         arg.UserID,
			FoodlogentryID: arg.ID.String(),
			MetricKey:      k,
			MetricValue:    v,
		})
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func SetupSqlite(cfg *config.ServerConfiguration) (types.Storage, error) {

	db, err := sql.Open("sqlite", cfg.SqliteFile)
	if err != nil {
		return nil, err
	}
	// Create migrations table
	_, err = db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS migrations (id TEXT);")
	if err != nil {
		return nil, err
	}

	for _, migration := range getPgMigrations() {
		selectSQL := `
		SELECT id
		FROM migrations
		WHERE id = ?;
`
		ctx, cncl := context.WithTimeout(context.Background(), time.Second*5)
		defer cncl()
		rows := db.QueryRowContext(ctx, selectSQL, migration[0])
		if err != nil {
			return nil, err
		}

		if rows.Scan() == nil {
			// already done, just continue
			log.Debug().Any("migrationindex", migration[0]).Msg("migration already done")
			continue
		}
		log.Debug().Any("migrationindex", migration[0]).Msg("migration being processed")
		_, err = db.ExecContext(context.Background(), migration[1])
		if err != nil {
			return nil, err
		}
		_, err = db.ExecContext(context.Background(), "INSERT INTO migrations (id) VALUES (?);", migration[0])
		if err != nil {
			return nil, err
		}
		log.Debug().Any("migrationindex", migration[0]).Msg("migration successful")
	}

	log.Info().Msg("migrations complete")

	stg := Sqlite3Storage{
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
