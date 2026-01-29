package server

import (
	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/server/generated"
	"github.com/openhealthsuite/diary/internal/storage/strggen"
)

// GetUserConfig implements generated.ServerInterface.
func (g *ServerState) GetUserConfig(c *gin.Context, configId string) {
	userId, ok := c.Get("userId")
	if !ok {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}

	cfg, err := g.storage.GetQuerier().GetUserConfig(c, strggen.GetUserConfigParams{
		UserID: userId.(string),
		ID:     configId,
	})
	if err != nil {
		c.JSON(404, generated.Error{Code: 404, Message: "Config not found"})
		return
	}

	// Unmarshal config_value (JSONB) to ConfigurationValue
	var val generated.ConfigurationValue
	_ = val.UnmarshalJSON(cfg.ConfigValue)
	resp := generated.Configuration{
		Id:    cfg.ID,
		Value: val,
	}
	c.JSON(200, resp)
}

// StoreUserConfig implements generated.ServerInterface.
func (g *ServerState) StoreUserConfig(c *gin.Context, configId string) {
	userId, ok := c.Get("userId")
	if !ok {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}

	var val generated.ConfigurationValue
	if err := c.ShouldBindJSON(&val); err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid request body"})
		return
	}
	raw, _ := val.MarshalJSON()
	err := g.storage.GetQuerier().StoreUserConfig(c, strggen.StoreUserConfigParams{
		UserID:      userId.(string),
		ID:          configId,
		ConfigValue: raw,
	})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}
	c.Status(200)
}
