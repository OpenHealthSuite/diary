package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/server"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

func main() {
	var config config.ServerConfiguration
	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatal().Err(err).Msg("An error was encountered parsing environment variables")
	}
	srv, err := server.NewServer(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("Error setting up server")
	}
	quit, err := srv.RunServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Error running server")
	}
	signal.Notify(*quit, os.Interrupt)
	for {
	}
}
