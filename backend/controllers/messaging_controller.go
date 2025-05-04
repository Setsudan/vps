package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	connectionmanager "launay-dot-one/manager"
	"launay-dot-one/middlewares"
	"launay-dot-one/models"
	"launay-dot-one/services"
	"launay-dot-one/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// ----------------------------------------------------------------------------
// helpers
// ----------------------------------------------------------------------------

func parseJWT(tokenStr string, secret []byte) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil || !tok.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return nil, fmt.Errorf("missing user_id claim")
	}
	return claims, nil
}

func buildUpgrader() websocket.Upgrader {
	raw := utils.GetEnv("WS_ALLOWED_ORIGINS", "https://app.launay.one")
	allowed := strings.Split(raw, ",")

	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			for _, a := range allowed {
				if strings.TrimSpace(a) == origin {
					return true
				}
			}
			return false
		},
	}
}

// ----------------------------------------------------------------------------
// controller
// ----------------------------------------------------------------------------

type MessagingController struct {
	messagingService services.MessagingService
	groupService     services.GroupService
	logger           *logrus.Logger
	secret           []byte
	upgrader         websocket.Upgrader
}

func NewMessagingController(ms services.MessagingService, gs services.GroupService, l *logrus.Logger) *MessagingController {
	sec := utils.MustEnv("JWT_SECRET")
	return &MessagingController{
		messagingService: ms,
		groupService:     gs,
		logger:           l,
		secret:           []byte(sec),
		upgrader:         buildUpgrader(),
	}
}

func (mc *MessagingController) RegisterRoutes(r *gin.Engine) {
	msg := r.Group("/messages")
	{
		msg.GET("/history", middlewares.AuthMiddleware(), mc.GetChatHistory)
		msg.POST("/reaction", middlewares.AuthMiddleware(), mc.HandleAddReaction)
		msg.GET("/ws", mc.HandleWebSocket) // auth inside
		msg.GET("/conversations", middlewares.AuthMiddleware(), mc.GetAllUserConversations)
	}
}

// ----------------------------------------------------------------------------
// WebSocket handler
// ----------------------------------------------------------------------------

func (mc *MessagingController) HandleWebSocket(c *gin.Context) {
	// 1. Extract token from the Authorization header like the usual middleware.
	auth := c.GetHeader("Authorization")
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing Bearer token")
		return
	}
	claims, err := parseJWT(parts[1], mc.secret)
	if err != nil {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}
	senderID := claims["user_id"].(string)

	// 2. Upgrade the connection.
	conn, err := mc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		mc.logger.Error("WebSocket upgrade error: ", err)
		return
	}
	connectionmanager.ConnManager.Add(senderID, conn)
	defer connectionmanager.ConnManager.Remove(senderID)

	// 3. Handle messages.
	ctx := c.Request.Context()
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			mc.logger.Warn("WebSocket read error: ", err)
			break
		}
		var msg models.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			mc.logger.Warn("Invalid message JSON: ", err)
			continue
		}
		msg.SenderID = senderID
		msg.Timestamp = time.Now()

		if err := mc.messagingService.SendMessage(ctx, msg); err != nil {
			_ = conn.WriteJSON(utils.APIResponse{
				Code: http.StatusInternalServerError, Message: "Failed to send message", Error: err.Error(),
			})
			continue
		}

		switch msg.TargetType {
		case "user":
			if rc, ok := connectionmanager.ConnManager.Get(msg.TargetID); ok {
				_ = rc.WriteJSON(utils.APIResponse{Code: http.StatusOK, Message: "New message", Data: msg})
			}
		case "group", "channel":
			members, err := mc.groupService.ListMembers(ctx, msg.TargetID)
			if err != nil {
				mc.logger.Error("list members: ", err)
			}
			for _, m := range members {
				if m.UserID == senderID {
					continue
				}
				if rc, ok := connectionmanager.ConnManager.Get(m.UserID); ok {
					_ = rc.WriteJSON(utils.APIResponse{Code: http.StatusOK, Message: "New message", Data: msg})
				}
			}
		}

		_ = conn.WriteJSON(utils.APIResponse{Code: http.StatusOK, Message: "Message sent", Data: msg})
	}
}

func (mc *MessagingController) GetChatHistory(c *gin.Context) {
	targetID := c.Query("target_id")
	targetType := c.Query("target_type")
	if targetID == "" || targetType == "" {
		utils.RespondError(c, http.StatusBadRequest, "target_id and target_type are required", nil)
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user ID")
		return
	}

	messages, err := mc.messagingService.GetChatHistory(c.Request.Context(), userID, targetID)
	if err != nil {
		mc.logger.Error("Failed to retrieve chat history: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to retrieve chat history", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Chat history fetched", map[string]interface{}{"messages": messages})
}

func (mc *MessagingController) HandleAddReaction(c *gin.Context) {
	var payload struct {
		MessageID string `json:"message_id"`
		Reaction  string `json:"reaction"`
	}
	if err := c.BindJSON(&payload); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id")
		return
	}

	if err := mc.messagingService.AddReaction(c.Request.Context(), payload.MessageID, payload.Reaction, userID); err != nil {
		mc.logger.Error("Failed to add reaction: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to add reaction", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Reaction added", nil)
}
func (mc *MessagingController) GetAllUserConversations(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user ID")
		return
	}

	convos, err := mc.messagingService.GetUserConversations(c.Request.Context(), userID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch conversations", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Conversations retrieved", gin.H{
		"conversations": convos,
	})
}
