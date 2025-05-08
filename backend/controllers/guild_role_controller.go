package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"launay-dot-one/middlewares"
	"launay-dot-one/models/guilds"
	grsvc "launay-dot-one/services/guildroles"
	"launay-dot-one/utils"
)

type GuildRolesController struct {
	svc    grsvc.Service
	logger *logrus.Logger
}

type rolePayload struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []uint64 `json:"permissions"`
	Color       int      `json:"color"`
	Hoist       bool     `json:"hoist"`
	Position    int      `json:"position"`
}

func NewGuildRolesController(svc grsvc.Service, logger *logrus.Logger) *GuildRolesController {
	return &GuildRolesController{svc, logger}
}

func (rc *GuildRolesController) RegisterRoutes(r *gin.Engine) {
	grp := r.Group("/guilds/:guild_id/roles", middlewares.AuthMiddleware())
	{
		grp.GET("", rc.List)
		grp.POST("", rc.Create)
		grp.GET("/:role_id", rc.Get)
		grp.PUT("/:role_id", rc.Update)
		grp.DELETE("/:role_id", rc.Delete)
	}
}

func (rc *GuildRolesController) List(c *gin.Context) {
	guildID := c.Param("guild_id")
	roles, err := rc.svc.List(c.Request.Context(), guildID)
	if err != nil {
		rc.logger.Error("List roles error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list roles", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Roles fetched", roles)
}

func (rc *GuildRolesController) Create(c *gin.Context) {
	guildID := c.Param("guild_id")

	var in rolePayload
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	// OR-fold the slice into a single uint64 bitmask
	var mask uint64
	for _, p := range in.Permissions {
		mask |= p
	}

	role := guilds.GuildRole{
		GuildID:     guildID,
		Name:        in.Name,
		Permissions: mask, // ← packed mask
		Color:       in.Color,
		Hoist:       in.Hoist,
		Position:    in.Position,
	}

	if err := rc.svc.Create(c.Request.Context(), &role); err != nil {
		rc.logger.Error("Create role error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create role", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusCreated, "Role created", role)
}

func (rc *GuildRolesController) Update(c *gin.Context) {
	guildID := c.Param("guild_id")
	roleID := c.Param("role_id")

	var in rolePayload
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	var mask uint64
	for _, p := range in.Permissions {
		mask |= p
	}

	role := guilds.GuildRole{
		ID:          roleID,
		GuildID:     guildID,
		Name:        in.Name,
		Permissions: mask,
		Color:       in.Color,
		Hoist:       in.Hoist,
		Position:    in.Position,
	}

	if err := rc.svc.Update(c.Request.Context(), &role); err != nil {
		rc.logger.Error("Update role error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update role", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Role updated", role)
}

func (rc *GuildRolesController) Get(c *gin.Context) {
	roleID := c.Param("role_id")
	role, err := rc.svc.Get(c.Request.Context(), roleID)
	if err != nil {
		rc.logger.Error("Get role error: ", err)
		utils.RespondError(c, http.StatusNotFound, "Role not found", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Role fetched", role)
}

func (rc *GuildRolesController) Delete(c *gin.Context) {
	roleID := c.Param("role_id")
	if err := rc.svc.Delete(c.Request.Context(), roleID); err != nil {
		rc.logger.Error("Delete role error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to delete role", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Role deleted", nil)
}
