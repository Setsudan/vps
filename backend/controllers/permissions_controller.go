package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"launay-dot-one/middlewares"
	"launay-dot-one/models/guilds"
	permsvc "launay-dot-one/services/permissions"
	"launay-dot-one/utils"
)

type PermissionsController struct {
	svc    permsvc.Service
	logger *logrus.Logger
}

func NewPermissionsController(svc permsvc.Service, logger *logrus.Logger) *PermissionsController {
	return &PermissionsController{svc, logger}
}

func (pc *PermissionsController) RegisterRoutes(r *gin.Engine) {
	grp := r.Group("/guilds/:guild_id/permissions", middlewares.AuthMiddleware())
	{
		grp.GET("", pc.List)
		grp.POST("", pc.Create)
		grp.PUT("/:perm_id", pc.Update)
		grp.DELETE("/:perm_id", pc.Delete)
	}
}

func (pc *PermissionsController) List(c *gin.Context) {
	guildID := c.Param("guild_id")
	var (
		cat = c.Query("category_id")
		ch  = c.Query("channel_id")
	)
	var catPtr, chPtr *string
	if cat != "" {
		catPtr = &cat
	}
	if ch != "" {
		chPtr = &ch
	}
	out, err := pc.svc.List(c.Request.Context(), guildID, catPtr, chPtr)
	if err != nil {
		pc.logger.Error("List permissions error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list permissions", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Permissions fetched", out)
}

func (pc *PermissionsController) Create(c *gin.Context) {
	guildID := c.Param("guild_id")
	var o guilds.PermissionOverwrite
	if err := c.ShouldBindJSON(&o); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	o.GuildID = guildID
	if err := pc.svc.Create(c.Request.Context(), &o); err != nil {
		pc.logger.Error("Create permission error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create permission", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusCreated, "Permission created", o)
}

func (pc *PermissionsController) Update(c *gin.Context) {
	guildID := c.Param("guild_id")
	permID := c.Param("perm_id")
	var o guilds.PermissionOverwrite
	if err := c.ShouldBindJSON(&o); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	o.GuildID = guildID
	o.ID = permID
	if err := pc.svc.Update(c.Request.Context(), &o); err != nil {
		pc.logger.Error("Update permission error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update permission", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Permission updated", o)
}

func (pc *PermissionsController) Delete(c *gin.Context) {
	permID := c.Param("perm_id")
	if err := pc.svc.Delete(c.Request.Context(), permID); err != nil {
		pc.logger.Error("Delete permission error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to delete permission", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Permission deleted", nil)
}
