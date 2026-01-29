package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/storage/strggen"
	"github.com/rs/zerolog/log"
)

type Storage interface {
	GetQuerier() strggen.Querier
}

//go:generate sqlc generate

type PostgresStorage struct {
	Db      *pgxpool.Pool
	Querier strggen.Querier
}

// GetQuerier implements Storage.
func (p *PostgresStorage) GetQuerier() strggen.Querier {
	return p.Querier
}

func NewStorage(cfg *config.ServerConfiguration) (Storage, error) {
	db, err := pgxpool.New(context.Background(), cfg.PostgresConnectionString)
	if err != nil {
		return nil, err
	}
	err = db.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	err = setupPostgres(db)
	if err != nil {
		return nil, err
	}

	stg := PostgresStorage{
		Db:      db,
		Querier: strggen.New(db),
	}
	return &stg, nil
}

func setupPostgres(db *pgxpool.Pool) error {
	// Create migrations table
	_, err := db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS migrations (id varchar(255));")
	if err != nil {
		return err
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
			return err
		}

		if rows.Scan() != pgx.ErrNoRows {
			// already done, just continue
			log.Debug().Any("migrationindex", migration[0]).Msg("migration already done")
			continue
		}
		log.Debug().Any("migrationindex", migration[0]).Msg("migration being processed")
		_, err = db.Exec(context.Background(), migration[1])
		if err != nil {
			return err
		}
		_, err = db.Exec(context.Background(), "INSERT INTO migrations (id) VALUES ($1);", migration[0])
		if err != nil {
			return err
		}
		log.Debug().Any("migrationindex", migration[0]).Msg("migration successful")
	}

	log.Info().Msg("migrations complete")

	return nil
}
