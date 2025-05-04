package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for now; adjust for production.
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://app.launay.one" ||
			r.Header.Get("Origin") == "http://localhost"
	},
}

type MessagingController struct {
	messagingService services.MessagingService
	groupService     services.GroupService
	logger           *logrus.Logger
}

func NewMessagingController(ms services.MessagingService, gs services.GroupService, logger *logrus.Logger) *MessagingController {
	return &MessagingController{
		messagingService: ms,
		groupService:     gs,
		logger:           logger,
	}
}

// RegisterRoutes registers messaging endpoints with auth middleware.
func (mc *MessagingController) RegisterRoutes(r *gin.Engine) {
	messageRoutes := r.Group("/messages")
	{
		messageRoutes.GET("/history", middlewares.AuthMiddleware(), mc.GetChatHistory)
		messageRoutes.POST("/reaction", middlewares.AuthMiddleware(), mc.HandleAddReaction)
		messageRoutes.GET("/ws", mc.HandleWebSocket)
		messageRoutes.GET("/conversations", middlewares.AuthMiddleware(), mc.GetAllUserConversations)

	}
}

// HandleWebSocket upgrades the connection, stores it, processes incoming messages,
// and delivers them to the appropriate recipient(s) in real time.
func (mc *MessagingController) HandleWebSocket(c *gin.Context) {
	// Extract the token from the query string.
	tokenStr := c.Query("token")
	if tokenStr == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing token in query parameters")
		return
	}

	// Retrieve the JWT secret from the environment.
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		utils.RespondError(c, http.StatusInternalServerError, "Server error", "JWT_SECRET not set on server")
		return
	}

	// Validate the token using the same logic as the auth middleware.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Invalid or expired token")
		return
	}

	// Extract the user_id from the token claims.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id in token claims")
		return
	}
	senderID := claims["user_id"].(string)

	// Upgrade the connection to a websocket.
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		mc.logger.Error("WebSocket upgrade error: ", err)
		return
	}
	// Add the connection to the connection manager.
	connectionmanager.ConnManager.Add(senderID, conn)
	defer connectionmanager.ConnManager.Remove(senderID)

	ctx := c.Request.Context()
	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			mc.logger.Warn("WebSocket read error: ", err)
			break
		}

		var msg models.Message
		if err := json.Unmarshal(msgData, &msg); err != nil {
			mc.logger.Warn("Invalid message format: ", err)
			continue
		}
		// Ensure the sender's identity is correct and set the timestamp.
		msg.SenderID = senderID
		msg.Timestamp = time.Now()

		// Save the message temporarily in Redis.
		if err := mc.messagingService.SendMessage(ctx, msg); err != nil {
			mc.logger.Error("Failed to send message: ", err)
			_ = conn.WriteJSON(utils.APIResponse{
				Code:    http.StatusInternalServerError,
				Message: "Failed to send message",
				Error:   err.Error(),
			})
			continue
		}

		// Real-time delivery logic.
		switch msg.TargetType {
		case "user":
			// Direct message: deliver to recipient if connected.
			if recipientConn, ok := connectionmanager.ConnManager.Get(msg.TargetID); ok {
				if err := recipientConn.WriteJSON(utils.APIResponse{
					Code:    http.StatusOK,
					Message: "New message received",
					Data:    msg,
				}); err != nil {
					mc.logger.Error("Failed to deliver message to recipient: ", err)
				}
			} else {
				mc.logger.Infof("Recipient %s not connected", msg.TargetID)
			}

		case "group", "channel":
			// For groups/channels: fetch group members via groupService.
			members, err := mc.groupService.ListMembers(ctx, msg.TargetID)
			if err != nil {
				mc.logger.Error("Failed to list group members: ", err)
			} else {
				for _, membership := range members {
					// Skip the sender.
					if membership.UserID == senderID {
						continue
					}
					if recipientConn, ok := connectionmanager.ConnManager.Get(membership.UserID); ok {
						if err := recipientConn.WriteJSON(utils.APIResponse{
							Code:    http.StatusOK,
							Message: "New message received",
							Data:    msg,
						}); err != nil {
							mc.logger.Errorf("Failed to deliver message to member %s: %v", membership.UserID, err)
						}
					}
				}
			}
		}

		// Echo confirmation back to the sender.
		if err := conn.WriteJSON(utils.APIResponse{
			Code:    http.StatusOK,
			Message: "Message sent",
			Data:    msg,
		}); err != nil {
			mc.logger.Error("Failed to send confirmation to sender: ", err)
		}
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
