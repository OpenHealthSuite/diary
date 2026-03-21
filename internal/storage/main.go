package storage

import (
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/storage/postgres"
	"github.com/openhealthsuite/diary/internal/storage/types"
)

func NewStorage(cfg *config.ServerConfiguration) (types.Storage, error) {
	stg, err := postgres.SetupPostgres(cfg)
	if err != nil {
		return nil, err
	}

	return stg, nil
}
