package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/storage"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_MigrationsCompleteSuccessfully(t *testing.T) {
	ctx := context.Background()

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
	// Clean up the container
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed to terminate container")
		}
	}()

	cfg := config.ServerConfiguration{
		PostgresConnectionString: postgresContainer.MustConnectionString(ctx),
	}

	strg, err := storage.NewStorage(&cfg)

	if err != nil {
		log.Error().Err(err).Msg("error setting up storage")
		t.FailNow()
	}

	res, err := strg.GetQuerier().GetTestData(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error querying storage")
		t.FailNow()
	}
	assert.Equal(t, "Database is Live", *res)
}
