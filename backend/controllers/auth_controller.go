package controllers

import (
	"net/http"

	"launay-dot-one/models"
	authsvc "launay-dot-one/services/auth"
	"launay-dot-one/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	authService authsvc.Service
	logger      *logrus.Logger
}

func NewAuthController(svc authsvc.Service, logger *logrus.Logger) *AuthController {
	return &AuthController{
		authService: svc,
		logger:      logger,
	}
}

func (ac *AuthController) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", ac.Register)
		auth.POST("/login", ac.Login)
	}
}

func (ac *AuthController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		ac.logger.Warn("Invalid register payload: ", err)
		utils.RespondError(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	if err := ac.authService.RegisterUser(c.Request.Context(), &user); err != nil {
		ac.logger.Error("Registration failed: ", err)
		utils.RespondError(c, http.StatusInternalServerError, "Registration failed", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "User registered successfully", nil)
}

func (ac *AuthController) Login(c *gin.Context) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&creds); err != nil {
		ac.logger.Warn("Invalid login payload: ", err)
		utils.RespondError(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	token, err := ac.authService.LoginUser(c.Request.Context(), creds.Email, creds.Password)
	if err != nil {
		ac.logger.Warn("Login failed: ", err)
		utils.RespondError(c, http.StatusUnauthorized, "Invalid credentials", err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Login successful", gin.H{"token": token})
}
