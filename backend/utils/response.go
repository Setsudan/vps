package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// For Gin-based controllers
func RespondSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Code:    statusCode,
		Message: message,
		Data:    data,
	})
}

func RespondError(c *gin.Context, statusCode int, message string, err interface{}) {
	c.JSON(statusCode, APIResponse{
		Code:    statusCode,
		Message: message,
		Error:   err,
	})
}

// For net/http (raw handlers)
func RespondSuccessRaw(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Code:    statusCode,
		Message: message,
		Data:    data,
	})
}

func RespondErrorRaw(w http.ResponseWriter, statusCode int, message string, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Code:    statusCode,
		Message: message,
		Error:   err,
	})
}
