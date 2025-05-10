package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"launay-dot-one/middlewares"
	mg "launay-dot-one/models/guilds"
	guildsvc "launay-dot-one/services/guilds"
	"launay-dot-one/utils"
)

type GuildController struct {
	svc    guildsvc.Service
	logger *logrus.Logger
}

func NewGuildController(svc guildsvc.Service, logger *logrus.Logger) *GuildController {
	return &GuildController{svc: svc, logger: logger}
}

func (gc *GuildController) RegisterRoutes(r *gin.Engine) {
	grp := r.Group("/guilds", middlewares.AuthMiddleware())
	{
		grp.POST("", gc.CreateGuild)
		grp.GET("", gc.ListGuilds)
		grp.GET("/:guild_id", gc.GetGuild)
		grp.PUT("/:guild_id", gc.UpdateGuild)
		grp.DELETE("/:guild_id", gc.DeleteGuild)

		grp.POST("/:guild_id/members", gc.AddMember)
		grp.PUT("/:guild_id/members/:user_id", gc.UpdateMemberRoles)
		grp.DELETE("/:guild_id/members/:user_id", gc.RemoveMember)
		grp.GET("/:guild_id/members", gc.ListMembers)
	}
}

func (gc *GuildController) CreateGuild(c *gin.Context) {
	var guild mg.Guild
	if err := c.ShouldBindJSON(&guild); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	ownerID := c.GetString("user_id")
	if err := gc.svc.CreateGuild(c.Request.Context(), &guild, ownerID); err != nil {
		gc.logger.Error("CreateGuild error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create guild", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusCreated, "Guild created", guild)
}

func (gc *GuildController) ListGuilds(c *gin.Context) {
	list, err := gc.svc.ListGuilds(c.Request.Context())
	if err != nil {
		gc.logger.Error("ListGuilds error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list guilds", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Guilds fetched", list)
}

func (gc *GuildController) GetGuild(c *gin.Context) {
	id := c.Param("guild_id")
	guild, err := gc.svc.GetGuild(c.Request.Context(), id)
	if err != nil {
		gc.logger.Error("GetGuild error: ", err)
		utils.RespondError(c, http.StatusNotFound, "Guild not found", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Guild fetched", guild)
}

func (gc *GuildController) UpdateGuild(c *gin.Context) {
	id := c.Param("guild_id")
	var upd mg.Guild
	if err := c.ShouldBindJSON(&upd); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	requester := c.GetString("user_id")
	if err := gc.svc.UpdateGuild(c.Request.Context(), id, &upd, requester); err != nil {
		gc.logger.Error("UpdateGuild error: ", err)
		if err == guildsvc.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to update guild", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Guild updated", nil)
}

func (gc *GuildController) DeleteGuild(c *gin.Context) {
	id := c.Param("guild_id")
	requester := c.GetString("user_id")
	if err := gc.svc.DeleteGuild(c.Request.Context(), id, requester); err != nil {
		gc.logger.Error("DeleteGuild error: ", err)
		if err == guildsvc.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to delete guild", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Guild deleted", nil)
}

func (gc *GuildController) AddMember(c *gin.Context) {
	guildID := c.Param("guild_id")
	var payload struct {
		UserID  string   `json:"user_id" binding:"required"`
		RoleIDs []string `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	requester := c.GetString("user_id")
	if err := gc.svc.AddMember(c.Request.Context(), guildID, payload.UserID, payload.RoleIDs, requester); err != nil {
		gc.logger.Error("AddMember error: ", err)
		if err == guildsvc.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to add member", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Member added", nil)
}

func (gc *GuildController) UpdateMemberRoles(c *gin.Context) {
	guildID := c.Param("guild_id")
	userID := c.Param("user_id")
	var payload struct {
		RoleIDs []string `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	requester := c.GetString("user_id")
	if err := gc.svc.UpdateMemberRoles(c.Request.Context(), guildID, userID, payload.RoleIDs, requester); err != nil {
		gc.logger.Error("UpdateMemberRoles error: ", err)
		if err == guildsvc.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to update member roles", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Member roles updated", nil)
}

func (gc *GuildController) RemoveMember(c *gin.Context) {
	guildID := c.Param("guild_id")
	userID := c.Param("user_id")
	requester := c.GetString("user_id")
	if err := gc.svc.RemoveMember(c.Request.Context(), guildID, userID, requester); err != nil {
		gc.logger.Error("RemoveMember error: ", err)
		if err == guildsvc.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to remove member", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Member removed", nil)
}

func (gc *GuildController) ListMembers(c *gin.Context) {
	guildID := c.Param("guild_id")
	list, err := gc.svc.ListMembers(c.Request.Context(), guildID)
	if err != nil {
		gc.logger.Error("ListMembers error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list members", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Members fetched", list)
}
