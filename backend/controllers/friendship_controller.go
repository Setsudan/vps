package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"launay-dot-one/middlewares"
	frsvc "launay-dot-one/services/friendships"
	"launay-dot-one/utils"
)

type FriendshipController struct {
	svc    frsvc.Service
	logger *logrus.Logger
}

func NewFriendshipController(svc frsvc.Service, logger *logrus.Logger) *FriendshipController {
	return &FriendshipController{svc: svc, logger: logger}
}

func (fc *FriendshipController) RegisterRoutes(r *gin.Engine) {
	grp := r.Group("/friends", middlewares.AuthMiddleware())
	{
		grp.POST("/requests", fc.SendRequest)
		grp.POST("/requests/:request_id/respond", fc.RespondRequest)
		grp.GET("/requests", fc.ListRequests)
		grp.GET("", fc.ListFriends)
	}
}

// SendRequest handles POST /friends/requests
//
//	body: { "to_id": "<target user ID>" }
func (fc *FriendshipController) SendRequest(c *gin.Context) {
	var body struct {
		ToID string `json:"to_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	fromID := c.GetString("user_id")
	if err := fc.svc.SendRequest(c.Request.Context(), fromID, body.ToID); err != nil {
		fc.logger.Error("SendRequest error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to send friend request", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusCreated, "Friend request sent", nil)
}

// RespondRequest handles POST /friends/requests/:request_id/respond
//
//	body: { "accept": true }
func (fc *FriendshipController) RespondRequest(c *gin.Context) {
	requestID := c.Param("request_id")
	var body struct {
		Accept bool `json:"accept"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}
	if err := fc.svc.RespondRequest(c.Request.Context(), requestID, body.Accept); err != nil {
		fc.logger.Error("RespondRequest error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to respond to request", err.Error())
		return
	}
	msg := "Friend request rejected"
	if body.Accept {
		msg = "Friend request accepted"
	}
	utils.RespondSuccess(c, http.StatusOK, msg, nil)
}

// ListRequests handles GET /friends/requests
func (fc *FriendshipController) ListRequests(c *gin.Context) {
	userID := c.GetString("user_id")
	list, err := fc.svc.ListRequests(c.Request.Context(), userID)
	if err != nil {
		fc.logger.Error("ListRequests error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list requests", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Friend requests fetched", list)
}

// ListFriends handles GET /friends
func (fc *FriendshipController) ListFriends(c *gin.Context) {
	userID := c.GetString("user_id")
	friends, err := fc.svc.ListFriends(c.Request.Context(), userID)
	if err != nil {
		fc.logger.Error("ListFriends error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list friends", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Friends fetched", friends)
}
