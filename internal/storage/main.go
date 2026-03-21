package storage

import (
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/storage/postgres"
	sqlite "github.com/openhealthsuite/diary/internal/storage/sqlite3"
	"github.com/openhealthsuite/diary/internal/storage/types"
)

func NewStorage(cfg *config.ServerConfiguration) (types.Storage, error) {
	if cfg.PostgresConnectionString != "" {

		return postgres.SetupPostgres(cfg)
	}

	return sqlite.SetupSqlite(cfg)
}
