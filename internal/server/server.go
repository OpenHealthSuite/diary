package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/api"
	"github.com/openhealthsuite/diary/internal/auth"
	"github.com/openhealthsuite/diary/internal/charts"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/foodlogs"
	"github.com/openhealthsuite/diary/internal/metrics"
	"github.com/openhealthsuite/diary/internal/pages"
	"github.com/openhealthsuite/diary/internal/storage"
	strgtyp "github.com/openhealthsuite/diary/internal/storage/types"
	"github.com/rs/zerolog/log"
)

type DiaryServer interface {
	RunServer() (*chan os.Signal, error)
}

func NewServer(cfg *config.ServerConfiguration) (DiaryServer, error) {
	strg, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, err
	}
	mtr, err := metrics.NewMetricsProvider(strg)
	if err != nil {
		return nil, err
	}
	lgs, err := foodlogs.NewFoodLogService(strg)
	if err != nil {
		return nil, err
	}
	sts := &ServerState{
		Config:   cfg,
		storage:  strg,
		metrics:  mtr,
		foodlogs: lgs,
	}
	err = sts.setupHttpServer()
	if err != nil {
		return nil, err
	}
	return sts, nil
}

func (sts *ServerState) setupHttpServer() error {
	r := gin.Default()

	// User identification middleware
	r.Use(auth.UserAuthenticationMiddleware(sts.Config))

	err := pages.Setup(sts.Config, sts.storage, sts.metrics, sts.foodlogs, r)

	if err != nil {
		return err
	}

	err = charts.Setup(sts.Config, sts.storage, sts.metrics, sts.foodlogs, r)
	if err != nil {
		return err
	}

	err = api.Setup(sts.Config, sts.storage, sts.metrics, sts.foodlogs, r)

	if err != nil {
		return err
	}

	sts.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", sts.Config.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return nil
}

type ServerState struct {
	Config   *config.ServerConfiguration
	storage  strgtyp.Storage
	metrics  metrics.MetricsProvider
	foodlogs foodlogs.FoodLogService

	httpServer *http.Server
}

func (sts *ServerState) RunServer() (*chan os.Signal, error) {

	quit := make(chan os.Signal, 1)

	go func() {
		<-quit
		log.Info().Msg("receive interrupt signal")
		if err := sts.httpServer.Close(); err != nil {
			log.Error().Err(err).Msg("Server Close error")
		}
	}()

	go func() {
		if err := sts.httpServer.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Info().Msg("Server closed under request")
			} else {
				log.Error().Err(err).Msg("Server closed unexpectedly")
			}
		}

	}()
	return &quit, nil
}
