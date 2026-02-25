package router

import (
	"smartcalendar/ai"
	"smartcalendar/config"
	"smartcalendar/controller"
	"smartcalendar/middleware"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 注册路由与中间件。
func SetupRouter(cfg config.AppConfig) *gin.Engine {
	r := gin.Default()
	allowOrigins := buildAllowOrigins(cfg.CorsAllowOrigin)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	authController := controller.AuthController{Cfg: cfg}
	aiController := controller.AIController{Cfg: cfg, Service: ai.NewAIService(cfg)}
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
			authed.POST("/ai/speech/submit", aiController.SpeechSubmit)
			authed.POST("/ai/speech/query", aiController.SpeechQuery)

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

func buildAllowOrigins(configValue string) []string {
	candidates := []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	for _, item := range strings.Split(configValue, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			candidates = append(candidates, trimmed)
		}
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(candidates))
	for _, item := range candidates {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
