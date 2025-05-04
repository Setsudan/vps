package main

import (
	"launay-dot-one/controllers"
	"launay-dot-one/middlewares"
	"launay-dot-one/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authController *controllers.AuthController,
	presenceController *controllers.PresenceController,
	userController *controllers.UserController,
	messagingController *controllers.MessagingController,
	groupController *controllers.GroupController,
	locationController *controllers.LocationController,
) *gin.Engine {
	router := gin.New()

	// Middlewares
	router.Use(gin.Recovery())
	router.Use(middlewares.Logger())

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		utils.RespondSuccess(c, http.StatusOK, "Healthy", nil)
	})

	// Liveness check route
	router.GET("/liveness", func(c *gin.Context) {
		utils.RespondSuccess(c, http.StatusOK, "Alive", nil)
	})

	// Routes registration.
	authController.RegisterRoutes(router)
	userController.RegisterRoutes(router)
	groupController.RegisterRoutes(router)

	locationController.RegisterRoutes(router)
	messagingController.RegisterRoutes(router)

	// Other websocket routes.
	router.GET("/presence", gin.WrapF(presenceController.GetAllPresence))
	router.GET("/ws/presence", gin.WrapF(presenceController.HandleWebSocket))

	// Route path
	router.NoRoute(func(c *gin.Context) {
		utils.RespondError(c, http.StatusNotFound, "Not Found", "The requested resource was not found")
	})

	router.GET("/", func(c *gin.Context) {
		utils.RespondSuccess(c, http.StatusOK, "launay-dot-one/Ogma's Golang API", nil)
	})

	return router
}
