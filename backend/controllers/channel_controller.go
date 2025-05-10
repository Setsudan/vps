package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"launay-dot-one/middlewares"
	"launay-dot-one/models/guilds"
	chsvc "launay-dot-one/services/channels"
	"launay-dot-one/utils"
)

const channelIDParam = "/:channel_id"

type ChannelsController struct {
	svc    chsvc.Service
	logger *logrus.Logger
}

func NewChannelsController(svc chsvc.Service, logger *logrus.Logger) *ChannelsController {
	return &ChannelsController{svc, logger}
}

func (cc *ChannelsController) RegisterRoutes(r *gin.Engine) {
	// Under a guild
	g := r.Group("/guilds/:guild_id/channels", middlewares.AuthMiddleware())
	{
		g.GET("", cc.ListByGuild)
		g.POST("", cc.Create)
	}
	// Under a category
	c := r.Group("/categories/:category_id/channels", middlewares.AuthMiddleware())
	{
		c.GET("", cc.ListByCategory)
	}
	// Single-channel operations
	single := r.Group("/channels", middlewares.AuthMiddleware())
	{
		single.GET(channelIDParam, cc.Get)
		single.PUT(channelIDParam, cc.Update)
		single.DELETE(channelIDParam, cc.Delete)
	}
}

func (cc *ChannelsController) Create(c *gin.Context) {
	guildID := c.Param("guild_id")

	var ch guilds.Channel
	if err := c.ShouldBindJSON(&ch); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	ch.GuildID = guildID

	// Pass nil for categoryID → top-level channel
	if err := cc.svc.Create(c.Request.Context(), &ch, nil); err != nil {
		cc.logger.Error("Create channel error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create channel", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusCreated, "Channel created", ch)
}

func (cc *ChannelsController) ListByGuild(c *gin.Context) {
	gid := c.Param("guild_id")
	out, err := cc.svc.ListByGuild(c.Request.Context(), gid)
	if err != nil {
		cc.logger.Error("ListByGuild error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list channels", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Channels fetched", out)
}

func (cc *ChannelsController) ListByCategory(c *gin.Context) {
	cid := c.Param("category_id")
	out, err := cc.svc.ListByCategory(c.Request.Context(), cid)
	if err != nil {
		cc.logger.Error("ListByCategory error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list channels", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Channels fetched", out)
}

func (cc *ChannelsController) Get(c *gin.Context) {
	id := c.Param("channel_id")
	out, err := cc.svc.Get(c.Request.Context(), id)
	if err != nil {
		cc.logger.Error("Get channel error: ", err)
		utils.RespondError(c, http.StatusNotFound, "Channel not found", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Channel fetched", out)
}

func (cc *ChannelsController) Update(c *gin.Context) {
	id := c.Param("channel_id")
	var ch guilds.Channel
	if err := c.ShouldBindJSON(&ch); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	ch.ID = id
	if err := cc.svc.Update(c.Request.Context(), &ch); err != nil {
		cc.logger.Error("Update channel error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update channel", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Channel updated", ch)
}

func (cc *ChannelsController) Delete(c *gin.Context) {
	id := c.Param("channel_id")
	if err := cc.svc.Delete(c.Request.Context(), id); err != nil {
		cc.logger.Error("Delete channel error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to delete channel", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Channel deleted", nil)
}
