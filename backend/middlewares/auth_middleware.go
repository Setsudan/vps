package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"launay-dot-one/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	// capture the secret once
	secret := []byte(utils.MustEnv("JWT_SECRET"))

	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing or invalid Authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return secret, nil
		})
		if err != nil || !tok.Valid {
			utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", err.Error())
			c.Abort()
			return
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Invalid token claims")
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			utils.RespondError(c, http.StatusUnauthorized, "Unauthorized", "Missing user_id claim")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
