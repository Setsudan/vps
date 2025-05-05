package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"launay-dot-one/middlewares"
	usersvc "launay-dot-one/services/users"
	"launay-dot-one/utils"
)

type UserController struct {
	userSvc usersvc.Service
	logger  *logrus.Logger
}

func NewUserController(
	logger *logrus.Logger,
	userSvc usersvc.Service,
) *UserController {
	return &UserController{
		userSvc: userSvc,
		logger:  logger,
	}
}

func (uc *UserController) RegisterRoutes(r *gin.Engine) {
	// /users endpoints
	users := r.Group("/users", middlewares.AuthMiddleware())
	{
		users.GET("/all", uc.GetAllUsers)
		users.POST("/avatar", uc.ChangeAvatar)
		users.GET("/:user_id", uc.GetUserByID)
	}

	// /user endpoints
	user := r.Group("/user", middlewares.AuthMiddleware())
	{
		user.PUT("/me", uc.UpdateProfile)
		user.GET("/me", uc.GetCurrent)
	}
}

// GetAllUsers returns all users in public form.
func (uc *UserController) GetAllUsers(c *gin.Context) {
	list, err := uc.userSvc.List(c.Request.Context())
	if err != nil {
		uc.logger.Error("GetAllUsers error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch users", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Users fetched", list)
}

// ChangeAvatar handles avatar file upload.
func (uc *UserController) ChangeAvatar(c *gin.Context) {
	userID := c.GetString("user_id")

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Error reading file", err.Error())
		return
	}
	defer file.Close()

	uc.logger.Infof("Avatar upload: user=%s, file=%s, size=%d", userID, header.Filename, header.Size)

	url, err := uc.userSvc.ChangeAvatar(c.Request.Context(), file, header, userID)
	if err != nil {
		uc.logger.Error("ChangeAvatar error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to upload avatar", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Avatar uploaded", gin.H{"avatarUrl": url})
}

// GetUserByID returns a public profile for the given user_id.
func (uc *UserController) GetUserByID(c *gin.Context) {
	userID := c.Param("user_id")
	pu, err := uc.userSvc.GetByID(c.Request.Context(), userID)
	if err != nil {
		uc.logger.Error("GetUserByID error: ", err)
		utils.RespondError(c, http.StatusNotFound, "User not found", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "User fetched", pu)
}

// GetCurrent returns the profile of the authenticated user.
func (uc *UserController) GetCurrent(c *gin.Context) {
	userID := c.GetString("user_id")
	pu, err := uc.userSvc.GetCurrent(c.Request.Context(), userID)
	if err != nil {
		uc.logger.Error("GetCurrent error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch user", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "User fetched", pu)
}

// UpdateProfile applies partial updates to the authenticated user's profile.
func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var input struct {
		Username *string `json:"username"`
		Bio      *string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	updates := make(map[string]interface{})
	if input.Username != nil {
		updates["username"] = *input.Username
	}
	if input.Bio != nil {
		updates["bio"] = *input.Bio
	}

	if err := uc.userSvc.UpdateProfile(c.Request.Context(), userID, updates); err != nil {
		uc.logger.Error("UpdateProfile error: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Profile updated", nil)
}
