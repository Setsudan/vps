package controllers

import (
	"net/http"

	"launay-dot-one/middlewares"
	"launay-dot-one/models"
	"launay-dot-one/services"
	"launay-dot-one/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type GroupController struct {
	groupService services.GroupService
	logger       *logrus.Logger
}

func NewGroupController(gs services.GroupService, logger *logrus.Logger) *GroupController {
	return &GroupController{
		groupService: gs,
		logger:       logger,
	}
}

// RegisterRoutes registers all group-related endpoints with auth middleware.
func (gc *GroupController) RegisterRoutes(r *gin.Engine) {
	// Wrap group routes with auth middleware.
	groupRoutes := r.Group("/groups", middlewares.AuthMiddleware())
	{
		groupRoutes.POST("", gc.CreateGroup)
		groupRoutes.GET("/:group_id", gc.GetGroup)
		groupRoutes.GET("", gc.ListGroups)
		groupRoutes.PUT("/:group_id", gc.UpdateGroup)
		groupRoutes.DELETE("/:group_id", gc.DeleteGroup)

		// Membership endpoints.
		groupRoutes.POST("/:group_id/members", gc.AddMember)
		groupRoutes.PUT("/:group_id/members/:user_id", gc.UpdateMemberRole)
		groupRoutes.DELETE("/:group_id/members/:user_id", gc.RemoveMember)
		groupRoutes.GET("/:group_id/members", gc.ListMembers)
	}
}

func (gc *GroupController) CreateGroup(c *gin.Context) {
	var group models.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	creatorID := c.GetString("user_id")
	if creatorID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id")
		return
	}

	if err := gc.groupService.CreateGroup(c.Request.Context(), &group, creatorID); err != nil {
		gc.logger.Error("Failed to create group: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create group", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "Group created", group)
}

func (gc *GroupController) GetGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	group, err := gc.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		gc.logger.Error("Failed to get group: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to get group", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Group fetched", group)
}

func (gc *GroupController) ListGroups(c *gin.Context) {
	groups, err := gc.groupService.ListGroups(c.Request.Context())
	if err != nil {
		gc.logger.Error("Failed to list groups: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list groups", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Groups fetched", groups)
}

func (gc *GroupController) UpdateGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	var update models.Group
	if err := c.ShouldBindJSON(&update); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	requesterID := c.GetString("user_id")
	if requesterID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id")
		return
	}

	if err := gc.groupService.UpdateGroup(c.Request.Context(), groupID, &update, requesterID); err != nil {
		gc.logger.Error("Failed to update group: ", err)
		if err == services.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Unauthorized", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to update group", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Group updated", nil)
}

func (gc *GroupController) DeleteGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	requesterID := c.GetString("user_id")
	if requesterID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id")
		return
	}
	if err := gc.groupService.DeleteGroup(c.Request.Context(), groupID, requesterID); err != nil {
		gc.logger.Error("Failed to delete group: ", err)
		if err == services.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Unauthorized", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to delete group", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Group deleted", nil)
}

func (gc *GroupController) AddMember(c *gin.Context) {
	groupID := c.Param("group_id")
	var payload struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	requesterID := c.GetString("user_id")
	if requesterID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id")
		return
	}
	if err := gc.groupService.AddMember(c.Request.Context(), groupID, payload.UserID, payload.Role, requesterID); err != nil {
		gc.logger.Error("Failed to add member: ", err)
		if err == services.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Unauthorized", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to add member", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Member added", nil)
}

func (gc *GroupController) UpdateMemberRole(c *gin.Context) {
	groupID := c.Param("group_id")
	userID := c.Param("user_id")
	var payload struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	requesterID := c.GetString("user_id")
	if requesterID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id")
		return
	}
	if err := gc.groupService.UpdateMemberRole(c.Request.Context(), groupID, userID, payload.Role, requesterID); err != nil {
		gc.logger.Error("Failed to update member role: ", err)
		if err == services.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Unauthorized", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to update member role", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Member role updated", nil)
}

func (gc *GroupController) RemoveMember(c *gin.Context) {
	groupID := c.Param("group_id")
	userID := c.Param("user_id")
	requesterID := c.GetString("user_id")
	if requesterID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id")
		return
	}
	if err := gc.groupService.RemoveMember(c.Request.Context(), groupID, userID, requesterID); err != nil {
		gc.logger.Error("Failed to remove member: ", err)
		if err == services.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Unauthorized", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to remove member", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Member removed", nil)
}

func (gc *GroupController) ListMembers(c *gin.Context) {
	groupID := c.Param("group_id")
	members, err := gc.groupService.ListMembers(c.Request.Context(), groupID)
	if err != nil {
		gc.logger.Error("Failed to list members: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list members", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Members fetched", members)
}
