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

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

func allowedOriginsFromEnv() []string {
	raw := utils.GetEnv("WS_ALLOWED_ORIGINS", "https://app.launay.one")
	parts := strings.Split(raw, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// -----------------------------------------------------------------------------
// PresenceController
// -----------------------------------------------------------------------------

type PresenceController struct {
	presenceService realtime.PresenceService
	redisClient     *redis.Client
	logger          *logrus.Logger
	secret          []byte
	upgrader        websocket.Upgrader
}

func NewPresenceController(ps realtime.PresenceService, rc *redis.Client, l *logrus.Logger) *PresenceController {
	allowed := allowedOriginsFromEnv()
	return &PresenceController{
		presenceService: ps,
		redisClient:     rc,
		logger:          l,
		secret:          []byte(utils.MustEnv("JWT_SECRET")),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, o := range allowed {
					if o == origin {
						return true
					}
				}
				return false
			},
		},
	}
}

// HandleWebSocket authenticates, upgrades to WS, and tracks presence.
func (pc *PresenceController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 1. Extract JWT (prefer header, fallback to sub‑protocol "jwt,<token>").
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

	claims, err := parseJWT(tokenStr, pc.secret)
	if err != nil {
		utils.RespondErrorRaw(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}
	userID := claims["user_id"].(string)

	// 2. Upgrade connection.
	conn, err := pc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		pc.logger.Error("WS upgrade failed: ", err)
		return
	}
	defer conn.Close()

	ctx := r.Context()
	pc.setInitialStatus(ctx, userID, conn)
	pc.listenForMessages(ctx, userID, conn)
	pc.setDisconnectedStatus(ctx, userID)
}

// -----------------------------------------------------------------------------

func (pc *PresenceController) setInitialStatus(ctx context.Context, userID string, c *websocket.Conn) {
	if err := pc.presenceService.SetStatus(ctx, userID, "online"); err != nil {
		pc.logger.Error("set online: ", err)
	}
	_ = c.WriteJSON(map[string]string{"status": "online", "message": "You are now online"})
}

func (pc *PresenceController) listenForMessages(ctx context.Context, userID string, c *websocket.Conn) {
	for {
		_, payload, err := c.ReadMessage()
		if err != nil {
			pc.logger.Warn("ws read: ", err)
			break
		}
		pc.handleMessage(ctx, userID, c, payload)
	}
}

func (pc *PresenceController) handleMessage(ctx context.Context, userID string, c *websocket.Conn, payload []byte) {
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

func (pc *PresenceController) setDisconnectedStatus(ctx context.Context, userID string) {
	if err := pc.presenceService.SetStatus(ctx, userID, "disconnected"); err != nil {
		pc.logger.Error("set disconnected: ", err)
	}
}

// -----------------------------------------------------------------------------
// Helper endpoint
// -----------------------------------------------------------------------------

func (pc *PresenceController) GetAllPresence(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	keys, err := pc.redisClient.Keys(ctx, "presence:*").Result()
	if err != nil {
		utils.RespondErrorRaw(w, http.StatusInternalServerError, "Failed to fetch presence keys", err.Error())
		return
	}

	out := make(map[string]string, len(keys))
	for _, k := range keys {
		status, err := pc.redisClient.Get(ctx, k).Result()
		if err != nil {
			pc.logger.Warn("redis get ", k, ": ", err)
			continue
		}
		out[strings.TrimPrefix(k, "presence:")] = status
	}

	utils.RespondSuccessRaw(w, http.StatusOK, "Presence map fetched", out)
}
