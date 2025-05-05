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

type PresenceController struct {
	presenceService realtime.PresenceService
	redisClient     *redis.Client
	logger          *logrus.Logger
	secret          []byte
	upgrader        websocket.Upgrader
}

func NewPresenceController(
	ps realtime.PresenceService,
	rc *redis.Client,
	l *logrus.Logger,
) *PresenceController {
	// buildUpgrader reads WS_ALLOWED_ORIGINS from env
	return &PresenceController{
		presenceService: ps,
		redisClient:     rc,
		logger:          l,
		secret:          []byte(utils.MustEnv("JWT_SECRET")),
		upgrader:        BuildUpgrader(),
	}
}

// HandleWebSocket authenticates, upgrades to WS, and tracks presence.
func (pc *PresenceController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 1. Extract token (Bearer or Sec-WebSocket-Protocol "jwt,<token>")
	auth := r.Header.Get("Authorization")
	var tokenStr string
	if strings.HasPrefix(auth, "Bearer ") {
		tokenStr = strings.TrimPrefix(auth, "Bearer ")
	} else if proto := r.Header.Get("Sec-WebSocket-Protocol"); strings.HasPrefix(proto, "jwt,") {
		tokenStr = strings.TrimPrefix(proto, "jwt,")
	}
	if tokenStr == "" {
		utils.RespondErrorRaw(w, http.StatusUnauthorized, "Unauthorized", "Missing Bearer token")
		return
	}

	// 2. Parse claims
	claims, err := ParseJWT(tokenStr, pc.secret)
	if err != nil {
		utils.RespondErrorRaw(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}
	userID := claims["user_id"].(string)

	// 3. Upgrade
	conn, err := pc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		pc.logger.Error("WS upgrade failed: ", err)
		return
	}
	defer conn.Close()

	// 4. Mark online and listen
	ctx := r.Context()
	pc.setInitialStatus(ctx, userID, conn)
	pc.listenForMessages(ctx, userID, conn)
	pc.setDisconnectedStatus(ctx, userID)
}

// setInitialStatus marks user online and notifies.
func (pc *PresenceController) setInitialStatus(ctx context.Context, userID string, c *websocket.Conn) {
	if err := pc.presenceService.SetStatus(ctx, userID, "online"); err != nil {
		pc.logger.Error("set online: ", err)
	}
	_ = c.WriteJSON(map[string]string{"status": "online", "message": "You are now online"})
}

// listenForMessages handles incoming status updates.
func (pc *PresenceController) listenForMessages(ctx context.Context, userID string, c *websocket.Conn) {
	for {
		_, payload, err := c.ReadMessage()
		if err != nil {
			pc.logger.Warn("WS read: ", err)
			break
		}
		pc.handleMessage(ctx, userID, c, payload)
	}
}

// handleMessage parses a status update and applies it.
func (pc *PresenceController) handleMessage(
	ctx context.Context,
	userID string,
	c *websocket.Conn,
	payload []byte,
) {
	var req struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		_ = c.WriteJSON(map[string]string{"error": "invalid json"})
		return
	}

	switch req.Status {
	case "online", "away", "dnd":
		if err := pc.presenceService.SetStatus(ctx, userID, req.Status); err != nil {
			pc.logger.Error("update status: ", err)
		}
		_ = c.WriteJSON(map[string]string{"status": req.Status, "message": "Status updated"})
	default:
		_ = c.WriteJSON(map[string]string{"error": "Invalid status"})
	}
}

// setDisconnectedStatus marks user offline on disconnect.
func (pc *PresenceController) setDisconnectedStatus(ctx context.Context, userID string) {
	if err := pc.presenceService.SetStatus(ctx, userID, "disconnected"); err != nil {
		pc.logger.Error("set disconnected: ", err)
	}
}

// GetAllPresence returns all presence keys & statuses.
func (pc *PresenceController) GetAllPresence(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	keys, err := pc.redisClient.Keys(ctx, "presence:*").Result()
	if err != nil {
		utils.RespondErrorRaw(w, http.StatusInternalServerError,
			"Failed to fetch presence keys", err.Error())
		return
	}

	out := make(map[string]string, len(keys))
	for _, k := range keys {
		status, err := pc.redisClient.Get(ctx, k).Result()
		if err != nil {
			pc.logger.Warnf("redis get %s: %v", k, err)
			continue
		}
		out[strings.TrimPrefix(k, "presence:")] = status
	}

	utils.RespondSuccessRaw(w, http.StatusOK, "Presence map fetched", out)
}
