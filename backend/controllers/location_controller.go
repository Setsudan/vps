package controllers

import (
	"fmt"
	"net/http"
	"os"

	"launay-dot-one/middlewares"
	"launay-dot-one/services"
	"launay-dot-one/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

// LocationController manages location updates and retrieval.
type LocationController struct {
	locationService services.LocationService
	logger          *logrus.Logger
}

// NewLocationController returns a new LocationController.
func NewLocationController(ls services.LocationService, logger *logrus.Logger) *LocationController {
	return &LocationController{
		locationService: ls,
		logger:          logger,
	}
}

func (lc *LocationController) RegisterRoutes(r *gin.Engine) {
	// WebSocket endpoint for real-time location updates using query token.
	wsRoutes := r.Group("/ws")
	{
		wsRoutes.GET("/location", lc.HandleWebSocket)
	}

	// REST endpoint to retrieve current locations (still protected by auth middleware).
	restRoutes := r.Group("/locations", middlewares.AuthMiddleware())
	{
		restRoutes.GET("/", lc.GetLocations)
	}
}

// HandleWebSocket upgrades the connection and listens for location updates.
func (lc *LocationController) HandleWebSocket(c *gin.Context) {
	// Extract the JWT token from query parameters.
	tokenStr := c.Query("token")
	if tokenStr == "" {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing token in query parameters")
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		utils.RespondError(c, http.StatusInternalServerError, "Server error", "JWT_SECRET not set on server")
		return
	}

	// Validate the token similarly as in the auth middleware.
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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id in token claims")
		return
	}
	userID := claims["user_id"].(string)

	// Upgrade the connection to a websocket.
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		lc.logger.Error("WebSocket upgrade error: ", err)
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()
	for {
		var msg struct {
			Type      string  `json:"type"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}

		if err := conn.ReadJSON(&msg); err != nil {
			lc.logger.Warn("WebSocket read error: ", err)
			break
		}

		if msg.Type == "location_update" {
			// Update the location in Redis.
			if err := lc.locationService.UpdateLocation(ctx, userID, msg.Latitude, msg.Longitude); err != nil {
				lc.logger.Error("Failed updating location: ", err)
				_ = conn.WriteJSON(utils.APIResponse{
					Code:    http.StatusInternalServerError,
					Message: "Failed updating location",
					Error:   err.Error(),
				})
				continue
			}

			// Broadcast the location update to other connected clients.
			broadcastMsg := struct {
				UserID    string  `json:"user_id"`
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			}{
				UserID:    userID,
				Latitude:  msg.Latitude,
				Longitude: msg.Longitude,
			}

			locationChan := make(chan services.Location, 1)
			locationChan <- services.Location{
				UserID:    broadcastMsg.UserID,
				Latitude:  broadcastMsg.Latitude,
				Longitude: broadcastMsg.Longitude,
			}
			close(locationChan)

			lc.locationService.BroadcastLocationUpdate(ctx, locationChan)

			_ = conn.WriteJSON(utils.APIResponse{
				Code:    http.StatusOK,
				Message: "Location updated and broadcasted",
			})
		}
	}
}

// GetLocations retrieves all current locations from Redis.
func (lc *LocationController) GetLocations(c *gin.Context) {
	ctx := c.Request.Context()
	locations, err := lc.locationService.GetAllLocations(ctx)
	if err != nil {
		lc.logger.Error("Failed to get locations: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to get locations", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Locations fetched", locations)
}
