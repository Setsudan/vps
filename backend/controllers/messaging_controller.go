package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	connectionmanager "launay-dot-one/manager"
	"launay-dot-one/middlewares"
	"launay-dot-one/models"
	"launay-dot-one/services/groups"
	"launay-dot-one/services/messaging"
	"launay-dot-one/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
)

type MessagingController struct {
	msgSvc   messaging.Service
	grpSvc   groups.Service
	logger   *logrus.Logger
	secret   []byte
	upgrader websocket.Upgrader
}

func NewMessagingController(
	ms messaging.Service,
	gs groups.Service,
	l *logrus.Logger,
) *MessagingController {
	sec := utils.MustEnv("JWT_SECRET")
	return &MessagingController{
		msgSvc:   ms,
		grpSvc:   gs,
		logger:   l,
		secret:   []byte(sec),
		upgrader: BuildUpgrader(),
	}
}
func (mc *MessagingController) RegisterRoutes(r *gin.Engine) {
	msg := r.Group("/messages")
	{
		msg.GET("/history", middlewares.AuthMiddleware(), mc.GetChatHistory)
		msg.POST("/reaction", middlewares.AuthMiddleware(), mc.HandleAddReaction)
		msg.GET("/ws", mc.HandleWebSocket) // internal auth
		msg.GET("/conversations", middlewares.AuthMiddleware(), mc.GetAllUserConversations)
	}
}

// GetChatHistory now dispatches to DM vs. channel‐based history.
func (mc *MessagingController) GetChatHistory(c *gin.Context) {
	targetID := c.Query("target_id")
	targetType := c.Query("target_type")
	if targetID == "" || targetType == "" {
		utils.RespondError(c, http.StatusBadRequest,
			"target_id and target_type are required", nil)
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		utils.RespondError(c, http.StatusUnauthorized,
			"Unauthorized", "Missing user_id")
		return
	}

	var (
		msgs []models.Message
		err  error
	)
	ctx := c.Request.Context()
	if targetType == "user" {
		msgs, err = mc.msgSvc.GetMessagesBetweenUsers(ctx, userID, targetID)
	} else {
		// group or channel → treat targetID as channel ID
		msgs, err = mc.msgSvc.GetChannelHistory(ctx, targetID)
	}
	if err != nil {
		mc.logger.Error("Failed to fetch chat history: ", err)
		utils.RespondError(c, http.StatusInternalServerError,
			"Failed to fetch chat history", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK,
		"Chat history fetched", gin.H{"messages": msgs})
}

// ----------------------------------------------------------------------------
// HandleWebSocket
// ----------------------------------------------------------------------------

func (mc *MessagingController) HandleWebSocket(c *gin.Context) {
	// 1) Auth
	auth := c.GetHeader("Authorization")
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.RespondError(c, http.StatusUnauthorized,
			"Unauthorized", "Missing Bearer token")
		return
	}
	claims, err := ParseJWT(parts[1], mc.secret)
	if err != nil {
		utils.RespondError(c, http.StatusUnauthorized,
			"Unauthorized", err.Error())
		return
	}
	senderID := claims["user_id"].(string)

	// 2) Upgrade
	conn, err := mc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		mc.logger.Error("WebSocket upgrade error: ", err)
		return
	}
	connectionmanager.ConnManager.Add(senderID, conn)
	defer connectionmanager.ConnManager.Remove(senderID)

	// 3) Read & broadcast loop
	ctx := c.Request.Context()
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			mc.logger.Warn("WS read error: ", err)
			break
		}

		// unmarshal only the old payload fields
		var p struct {
			TargetID    string          `json:"target_id"`
			TargetType  string          `json:"target_type"`
			Content     string          `json:"content"`
			Attachments json.RawMessage `json:"attachments,omitempty"`
		}
		if err := json.Unmarshal(data, &p); err != nil {
			mc.logger.Warn("Invalid WS JSON: ", err)
			continue
		}

		// build new Message (cast attachments)
		msg := models.Message{
			ChannelID:   p.TargetID,
			AuthorID:    senderID,
			Content:     p.Content,
			Attachments: datatypes.JSON(p.Attachments),
		}
		if err := mc.msgSvc.SendMessage(ctx, &msg); err != nil {
			_ = conn.WriteJSON(utils.APIResponse{
				Code:    http.StatusInternalServerError,
				Message: "Failed to send message",
				Error:   err.Error(),
			})
			continue
		}

		// dispatch
		switch p.TargetType {
		case "user":
			if rc, ok := connectionmanager.ConnManager.Get(p.TargetID); ok {
				_ = rc.WriteJSON(utils.APIResponse{
					Code:    http.StatusOK,
					Message: "New message",
					Data:    msg,
				})
			}

		case "group":
			members, err := mc.grpSvc.ListMembers(ctx, p.TargetID)
			if err != nil {
				mc.logger.Error("ListMembers error: ", err)
			}
			for _, m := range members {
				if m.UserID == senderID {
					continue
				}
				if rc, ok := connectionmanager.ConnManager.Get(m.UserID); ok {
					_ = rc.WriteJSON(utils.APIResponse{
						Code:    http.StatusOK,
						Message: "New message",
						Data:    msg,
					})
				}
			}
		}

		// ack back
		_ = conn.WriteJSON(utils.APIResponse{
			Code:    http.StatusOK,
			Message: "Message sent",
			Data:    msg,
		})
	}
}

// HandleAddReaction unchanged
func (mc *MessagingController) HandleAddReaction(c *gin.Context) {
	var p struct {
		MessageID string `json:"message_id"`
		Reaction  string `json:"reaction"`
	}
	if err := c.BindJSON(&p); err != nil {
		utils.RespondError(c, http.StatusBadRequest,
			"Invalid payload", err.Error())
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		utils.RespondError(c, http.StatusUnauthorized,
			"Unauthorized", "Missing user_id")
		return
	}
	if err := mc.msgSvc.AddReaction(
		c.Request.Context(), p.MessageID, p.Reaction, userID,
	); err != nil {
		mc.logger.Error("AddReaction error: ", err)
		utils.RespondError(c, http.StatusInternalServerError,
			"Failed to add reaction", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Reaction added", nil)
}

// GetAllUserConversations proxies to msgSvc.GetUserConversations
func (mc *MessagingController) GetAllUserConversations(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.RespondError(c, http.StatusUnauthorized,
			"Unauthorized", "Missing user_id")
		return
	}
	convos, err := mc.msgSvc.GetUserConversations(c.Request.Context(), userID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError,
			"Failed to fetch conversations", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK,
		"Conversations retrieved", gin.H{"conversations": convos})
}
