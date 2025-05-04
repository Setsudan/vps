package utils

import (
	"os"
	"strings"
)

// IsOriginAllowed checks the Origin header against ALLOWED_WS_ORIGINS.
//
// ALLOWED_WS_ORIGINS = "https://app.example.com,https://other.example.com"
func IsOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	allowed := strings.Split(os.Getenv("ALLOWED_WS_ORIGINS"), ",")
	for _, o := range allowed {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}
