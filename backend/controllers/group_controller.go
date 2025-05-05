package controllers

import (
	"net/http"

	"launay-dot-one/middlewares"
	"launay-dot-one/utils"

	mGroups "launay-dot-one/models/groups"
	groupService "launay-dot-one/services/groups"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type GroupController struct {
	svc    groupService.Service
	logger *logrus.Logger
}

func NewGroupController(svc groupService.Service, logger *logrus.Logger) *GroupController {
	return &GroupController{svc: svc, logger: logger}
}

func (gc *GroupController) RegisterRoutes(r *gin.Engine) {
	grp := r.Group("/groups", middlewares.AuthMiddleware())
	{
		grp.POST("", gc.CreateGroup)
		grp.GET("/:group_id", gc.GetGroup)
		grp.GET("", gc.ListGroups)
		grp.PUT("/:group_id", gc.UpdateGroup)
		grp.DELETE("/:group_id", gc.DeleteGroup)

		grp.POST("/:group_id/members", gc.AddMember)
		grp.PUT("/:group_id/members/:user_id", gc.UpdateMemberRole)
		grp.DELETE("/:group_id/members/:user_id", gc.RemoveMember)
		grp.GET("/:group_id/members", gc.ListMembers)
	}
}

func (gc *GroupController) CreateGroup(c *gin.Context) {
	var group mGroups.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	creatorID := c.GetString("user_id")
	if creatorID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id")
		return
	}

	if err := gc.svc.CreateGroup(c.Request.Context(), &group, creatorID); err != nil {
		gc.logger.Error("CreateGroup error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create group", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "Group created", group)
}

func (gc *GroupController) GetGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	group, err := gc.svc.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		gc.logger.Error("GetGroup error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to get group", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Group fetched", group)
}

func (gc *GroupController) ListGroups(c *gin.Context) {
	groups, err := gc.svc.ListGroups(c.Request.Context())
	if err != nil {
		gc.logger.Error("ListGroups error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list groups", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Groups fetched", groups)
}

func (gc *GroupController) UpdateGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	var upd mGroups.Group
	if err := c.ShouldBindJSON(&upd); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	requesterID := c.GetString("user_id")
	if requesterID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id")
		return
	}

	err := gc.svc.UpdateGroup(c.Request.Context(), groupID, &upd, requesterID)
	if err != nil {
		gc.logger.Error("UpdateGroup error: ", err)
		if err == groupService.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
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

	err := gc.svc.DeleteGroup(c.Request.Context(), groupID, requesterID)
	if err != nil {
		gc.logger.Error("DeleteGroup error: ", err)
		if err == groupService.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
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

	err := gc.svc.AddMember(c.Request.Context(), groupID, payload.UserID, payload.Role, requesterID)
	if err != nil {
		gc.logger.Error("AddMember error: ", err)
		if err == groupService.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
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

	err := gc.svc.UpdateMemberRole(c.Request.Context(), groupID, userID, payload.Role, requesterID)
	if err != nil {
		gc.logger.Error("UpdateMemberRole error: ", err)
		if err == groupService.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
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

	err := gc.svc.RemoveMember(c.Request.Context(), groupID, userID, requesterID)
	if err != nil {
		gc.logger.Error("RemoveMember error: ", err)
		if err == groupService.ErrUnauthorized {
			utils.RespondError(c, http.StatusForbidden, "Forbidden", err.Error())
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to remove member", err.Error())
		}
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Member removed", nil)
}

func (gc *GroupController) ListMembers(c *gin.Context) {
	groupID := c.Param("group_id")
	members, err := gc.svc.ListMembers(c.Request.Context(), groupID)
	if err != nil {
		gc.logger.Error("ListMembers error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list members", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Members fetched", members)
}
