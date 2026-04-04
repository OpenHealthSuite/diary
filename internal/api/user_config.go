package api

import (
	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/api/generated"
	"github.com/openhealthsuite/diary/internal/auth"
	"github.com/openhealthsuite/diary/internal/storage/types"
)

// GetUserConfig implements generated.ServerInterface.
func (g *ApiState) GetUserConfig(c *gin.Context, configId string) {
	uidptr, err := auth.GetUserId(c)
	if err != nil {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}
	userId := *uidptr

	cfg, err := g.Storage.GetUserConfig(c, types.GetUserConfigParams{
		UserID: userId,
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
func (g *ApiState) StoreUserConfig(c *gin.Context, configId string) {
	uidptr, err := auth.GetUserId(c)
	if err != nil {
		c.JSON(403, generated.Error{Code: 403, Message: "Missing user identification"})
		return
	}
	userId := *uidptr

	var val generated.ConfigurationValue
	if err := c.ShouldBindJSON(&val); err != nil {
		c.JSON(400, generated.Error{Code: 400, Message: "Invalid request body"})
		return
	}
	raw, _ := val.MarshalJSON()
	err = g.Storage.StoreUserConfig(c, types.StoreUserConfigParams{
		UserID:      userId,
		ID:          configId,
		ConfigValue: raw,
	})
	if err != nil {
		c.JSON(500, generated.Error{Code: 500, Message: err.Error()})
		return
	}
	c.Status(200)
}
