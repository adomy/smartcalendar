package router

import (
	"smartcalendar/ai"
	"smartcalendar/config"
	"smartcalendar/controller"
	"smartcalendar/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 注册路由与中间件。
func SetupRouter(cfg config.AppConfig) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CorsAllowOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	authController := controller.AuthController{Cfg: cfg}
	aiController := controller.AIController{Service: ai.NewAIService(cfg)}
	adminController := controller.AdminController{}
	eventController := controller.EventController{}
	userController := controller.UserController{}
	logController := controller.OperationLogController{}
	notificationController := controller.NotificationController{}
	uploadController := controller.UploadController{Cfg: cfg}

	api := r.Group("/api")
	{
		api.POST("/auth/register", authController.Register)
		api.POST("/auth/login", authController.Login)

		authed := api.Group("")
		authed.Use(middleware.AuthRequired(cfg))
		{
			authed.POST("/events", eventController.CreateEvent)
			authed.GET("/events", eventController.ListEvents)
			authed.GET("/events/:id", eventController.GetEventDetail)
			authed.PUT("/events/:id", eventController.UpdateEvent)
			authed.DELETE("/events/:id", eventController.DeleteEvent)

			authed.GET("/operation-logs", logController.ListLogs)

			authed.GET("/notifications", notificationController.ListNotifications)
			authed.GET("/notifications/unread-count", notificationController.UnreadCount)
			authed.PUT("/notifications/:id/read", notificationController.MarkRead)
			authed.PUT("/notifications/read-all", notificationController.MarkAllRead)

			authed.POST("/ai/chat", aiController.Chat)

			authed.GET("/user/profile", userController.GetProfile)
			authed.PUT("/user/profile", userController.UpdateProfile)
			authed.POST("/upload/avatar", uploadController.UploadAvatar)
			authed.GET("/users/search", userController.SearchUsers)

			admin := authed.Group("/admin")
			admin.Use(middleware.AdminRequired())
			{
				admin.GET("/users", adminController.ListUsers)
				admin.PUT("/users/:id/status", adminController.UpdateUserStatus)
				admin.PUT("/users/:id/reset-password", adminController.ResetPassword)
			}
		}
	}

	return r
}
