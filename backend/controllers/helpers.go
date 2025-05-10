package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"launay-dot-one/utils"
)

// parseJWT validates a Bearer token and returns its claims.
func ParseJWT(tokenStr string, secret []byte) (jwt.MapClaims, error) {
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

// buildUpgrader returns a websocket.Upgrader with origins from WS_ALLOWED_ORIGINS.
func BuildUpgrader() websocket.Upgrader {
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
