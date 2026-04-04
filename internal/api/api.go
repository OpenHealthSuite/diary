package api

import (
	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/api/generated"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/foodlogs"
	"github.com/openhealthsuite/diary/internal/metrics"
	"github.com/rs/zerolog/log"

	strgtyp "github.com/openhealthsuite/diary/internal/storage/types"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../tools/oapi_codegen/server.cfg.yaml ../../api/swagger.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../tools/oapi_codegen/types.cfg.yaml ../../api/swagger.yaml

type ServerState struct {
	Config   *config.ServerConfiguration
	Storage  strgtyp.Storage
	Metrics  metrics.MetricsProvider
	FoodLogs foodlogs.FoodLogService
}

func (srvst *ServerState) AsGeneratedInterface() generated.ServerInterface {
	return srvst
}

func Setup(sts *ServerState, r *gin.Engine) error {
	generated.RegisterHandlers(r, sts.AsGeneratedInterface())
	return nil
}

// TestEndpoint implements generated.ServerInterface.
func (g *ServerState) TestEndpoint(c *gin.Context) {
	_, err := g.Storage.GetTestData(c)
	if err != nil {
		log.Error().Err(err).Msg("error pinging database")
		c.String(500, "KO")
		return
	}
	c.String(200, "OK")
}
