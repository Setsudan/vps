package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"launay-dot-one/realtime"
	"launay-dot-one/utils"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://app.launay.one" ||
			r.Header.Get("Origin") == "http://localhost"
	},
}

// PresenceController handles WebSocket connections for user presence.
type PresenceController struct {
	presenceService realtime.PresenceService
	redisClient     *redis.Client
	logger          *logrus.Logger
}

// NewPresenceController returns a new PresenceController.
func NewPresenceController(ps realtime.PresenceService, redisClient *redis.Client, logger *logrus.Logger) *PresenceController {
	return &PresenceController{
		presenceService: ps,
		redisClient:     redisClient,
		logger:          logger,
	}
}

// HandleWebSocket upgrades the HTTP connection to WebSocket, sets initial status, and listens for status updates.
func (pc *PresenceController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		utils.RespondErrorRaw(w, http.StatusBadRequest, "Missing user_id", "user_id is required in query parameters")
		return
	}

	conn, err := pc.upgradeConnection(w, r)
	if err != nil {
		pc.logger.Error("WebSocket upgrade failed: ", err)
		return
	}
	defer conn.Close()

	ctx := r.Context()
	pc.setInitialStatus(ctx, userID, conn)

	pc.listenForMessages(ctx, userID, conn)
	pc.setDisconnectedStatus(ctx, userID)
}

func (pc *PresenceController) upgradeConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return upgrader.Upgrade(w, r, nil)
}

func (pc *PresenceController) setInitialStatus(ctx context.Context, userID string, conn *websocket.Conn) {
	if err := pc.presenceService.SetStatus(ctx, userID, "online"); err != nil {
		pc.logger.Error("Failed to set online status: ", err)
	}
	_ = conn.WriteJSON(map[string]string{"status": "online", "message": "You are now online"})
}

func (pc *PresenceController) listenForMessages(ctx context.Context, userID string, conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			pc.logger.Warn("WebSocket read error: ", err)
			break
		}

		pc.handleMessage(ctx, userID, conn, message)
	}
}

func (pc *PresenceController) handleMessage(ctx context.Context, userID string, conn *websocket.Conn, message []byte) {
	var msg struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		pc.logger.Warn("Invalid message format: ", err)
		return
	}

	if msg.Status == "online" || msg.Status == "away" || msg.Status == "dnd" {
		if err := pc.presenceService.SetStatus(ctx, userID, msg.Status); err != nil {
			pc.logger.Error("Failed to update status: ", err)
		} else {
			_ = conn.WriteJSON(map[string]string{"status": msg.Status, "message": "Status updated"})
		}
	} else {
		pc.logger.Warn("Received invalid status: ", msg.Status)
		_ = conn.WriteJSON(map[string]string{"error": "Invalid status"})
	}
}

func (pc *PresenceController) setDisconnectedStatus(ctx context.Context, userID string) {
	if err := pc.presenceService.SetStatus(ctx, userID, "disconnected"); err != nil {
		pc.logger.Error("Failed to set disconnected status: ", err)
	}
}

// GetAllPresence retrieves all keys with the "presence:" prefix from Redis
func (pc *PresenceController) GetAllPresence(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	keys, err := pc.redisClient.Keys(ctx, "presence:*").Result()
	if err != nil {
		utils.RespondErrorRaw(w, http.StatusInternalServerError, "Failed to fetch presence keys", err.Error())
		return
	}

	result := make(map[string]string)
	for _, key := range keys {
		status, err := pc.redisClient.Get(ctx, key).Result()
		if err != nil {
			pc.logger.Warn("Error fetching key ", key, ": ", err)
			continue
		}
		userID := strings.TrimPrefix(key, "presence:")
		result[userID] = status
	}

	utils.RespondSuccessRaw(w, http.StatusOK, "Presence map fetched", result)
}
