package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"launay-dot-one/middlewares"
	"launay-dot-one/models/guilds"
	catsvc "launay-dot-one/services/categories"
	"launay-dot-one/utils"
)

const categoryIDParam = "/:category_id"

type CategoriesController struct {
	svc    catsvc.Service
	logger *logrus.Logger
}

func NewCategoriesController(svc catsvc.Service, logger *logrus.Logger) *CategoriesController {
	return &CategoriesController{svc, logger}
}

func (cc *CategoriesController) RegisterRoutes(r *gin.Engine) {
	grp := r.Group("/guilds/:guild_id/categories", middlewares.AuthMiddleware())
	{
		grp.GET("", cc.List)
		grp.POST("", cc.Create)
		grp.GET(categoryIDParam, cc.Get)
		grp.PUT(categoryIDParam, cc.Update)
		grp.DELETE(categoryIDParam, cc.Delete)
	}
}

func (cc *CategoriesController) Create(c *gin.Context) {
	guildID := c.Param("guild_id")

	// Payload now includes optional Channels slice
	var payload struct {
		Name     string            `json:"name" binding:"required"`
		Position int               `json:"position" binding:"required"`
		Channels []*guilds.Channel `json:"channels,omitempty"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	// Build the Category
	category := guilds.Category{
		GuildID:  guildID,
		Name:     payload.Name,
		Position: payload.Position,
	}

	// Ensure each channel knows its guild
	for _, ch := range payload.Channels {
		ch.GuildID = guildID
	}

	// Pass channels slice into service
	if err := cc.svc.Create(c.Request.Context(), &category, payload.Channels); err != nil {
		cc.logger.Error("Create category error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create category", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusCreated, "Category created", category)
}

func (cc *CategoriesController) List(c *gin.Context) {
	guildID := c.Param("guild_id")
	out, err := cc.svc.List(c.Request.Context(), guildID)
	if err != nil {
		cc.logger.Error("List categories error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list categories", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Categories fetched", out)
}

func (cc *CategoriesController) Get(c *gin.Context) {
	id := c.Param("category_id")
	out, err := cc.svc.Get(c.Request.Context(), id)
	if err != nil {
		cc.logger.Error("Get category error: ", err)
		utils.RespondError(c, http.StatusNotFound, "Category not found", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Category fetched", out)
}

func (cc *CategoriesController) Update(c *gin.Context) {
	guildID := c.Param("guild_id")
	id := c.Param("category_id")
	var cat guilds.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	cat.ID = id
	cat.GuildID = guildID
	if err := cc.svc.Update(c.Request.Context(), &cat); err != nil {
		cc.logger.Error("Update category error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update category", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Category updated", cat)
}

func (cc *CategoriesController) Delete(c *gin.Context) {
	id := c.Param("category_id")
	if err := cc.svc.Delete(c.Request.Context(), id); err != nil {
		cc.logger.Error("Delete category error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to delete category", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Category deleted", nil)
}
