package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"launay-dot-one/models"
	"launay-dot-one/services"
	"launay-dot-one/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

type UserController struct {
	userService *services.UserService
	logger      *logrus.Logger
	jwtSecret   string
}

func NewUserController(logger *logrus.Logger, userService *services.UserService, jwtSecret string) *UserController {
	return &UserController{
		userService: userService,
		logger:      logger,
		jwtSecret:   jwtSecret,
	}
}

func (uc *UserController) RegisterRoutes(r *gin.Engine) {
	users := r.Group("/users")
	{
		users.GET("/all", uc.AuthMiddleware(), uc.GetAllUsersHandler)
		users.POST("/avatar", uc.AuthMiddleware(), uc.ChangeAvatar)
		users.GET("/:user_id", uc.AuthMiddleware(), uc.GetUserByIDHandler)
	}

	user := r.Group("/user")
	{
		user.PUT("/me", uc.AuthMiddleware(), uc.UpdateProfile)
		user.GET("/me", uc.AuthMiddleware(), uc.GetCurrentAuthenticatedUserHandler)
	}

}

func (uc *UserController) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondError(c, http.StatusUnauthorized, "Missing or invalid Authorization header", nil)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(uc.jwtSecret), nil
		})
		if err != nil || !token.Valid {
			uc.logger.Warnf("Token validation failed: %v", err)
			utils.RespondError(c, http.StatusUnauthorized, "Invalid token", err)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.RespondError(c, http.StatusUnauthorized, "Invalid token claims", nil)
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			utils.RespondError(c, http.StatusUnauthorized, "User ID not found in token", nil)
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func (uc *UserController) ChangeAvatar(c *gin.Context) {
	userID := c.GetString("user_id")

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Error retrieving file", err.Error())
		return
	}
	defer file.Close()

	uc.logger.Infof("Avatar upload received: filename=%s, size=%d, header=%v", header.Filename, header.Size, header.Header)

	avatarURL, err := uc.userService.ChangeAvatar(c.Request.Context(), file, header, userID)
	if err != nil {
		uc.logger.Error("Error uploading avatar: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Error uploading avatar", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Avatar uploaded successfully", gin.H{"avatarUrl": avatarURL})
}

func (uc *UserController) GetUserByIDHandler(c *gin.Context) {
	userID := c.Param("user_id")

	user, err := uc.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		uc.logger.Error("Error retrieving user by ID: ", err)
		utils.RespondError(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "User retrieved successfully", user)
}

func (uc *UserController) GetCurrentAuthenticatedUserHandler(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := uc.userService.GetCurrentAuthenticatedUser(c.Request.Context(), userID)
	if err != nil {
		uc.logger.Error("Error retrieving authenticated user: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to retrieve user", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Authenticated user retrieved", user)
}

func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var input struct {
		Username *string `json:"username"`
		Bio      *string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid input", nil)
		return
	}

	updates := map[string]interface{}{}
	if input.Username != nil {
		updates["username"] = *input.Username
	}
	if input.Bio != nil {
		updates["bio"] = *input.Bio
	}

	if err := uc.userService.GetDB().
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Profile updated", nil)
}

func (uc *UserController) GetAllUsersHandler(c *gin.Context) {
	users, err := uc.userService.GetAllUsers(c.Request.Context())
	if err != nil {
		uc.logger.Error("Failed to get all users: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to get users", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "All users fetched", users)
}
