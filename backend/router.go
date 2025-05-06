package main

import (
	"net/http"

	"launay-dot-one/controllers"
	"launay-dot-one/middlewares"
	"launay-dot-one/utils"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authController *controllers.AuthController,
	presenceController *controllers.PresenceController,
	userController *controllers.UserController,
	messagingController *controllers.MessagingController,
	groupController *controllers.GroupController,
	resumeController *controllers.ResumeController,
	friendshipController *controllers.FriendshipController,
	guildController *controllers.GuildController,
	permsController *controllers.PermissionsController,
	categoryController *controllers.CategoriesController,
	channelController *controllers.ChannelsController,
	guildRolesController *controllers.GuildRolesController,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), middlewares.Logger())

	// Health & Liveness
	router.GET("/health", func(c *gin.Context) {
		utils.RespondSuccess(c, http.StatusOK, "Healthy", nil)
	})
	router.GET("/liveness", func(c *gin.Context) {
		utils.RespondSuccess(c, http.StatusOK, "Alive", nil)
	})

	// Core routes
	authController.RegisterRoutes(router)
	userController.RegisterRoutes(router)
	groupController.RegisterRoutes(router)
	messagingController.RegisterRoutes(router)
	resumeController.RegisterRoutes(router)
	friendshipController.RegisterRoutes(router)
	guildController.RegisterRoutes(router)
	permsController.RegisterRoutes(router)
	categoryController.RegisterRoutes(router)
	channelController.RegisterRoutes(router)
	guildRolesController.RegisterRoutes(router)

	// Presence WS & helper
	router.GET("/presence", gin.WrapF(presenceController.GetAllPresence))
	router.GET("/ws/presence", gin.WrapF(presenceController.HandleWebSocket))

	// Catch‐all
	router.NoRoute(func(c *gin.Context) {
		utils.RespondError(c, http.StatusNotFound, "Not Found", "The requested resource was not found")
	})

	// Root info
	router.GET("/", func(c *gin.Context) {
		utils.RespondSuccess(c, http.StatusOK, "launay-dot-one/Ogma's Golang API", nil)
	})
	return router
}
