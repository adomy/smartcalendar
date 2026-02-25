package main

import (
	"net/http"
	"os"
	"time"

	"smartcalendar/config"
	"smartcalendar/model"
	"smartcalendar/router"
	"smartcalendar/service"

	"github.com/gin-gonic/gin"
)

// main 初始化配置、数据库与路由，并启动提醒任务与 HTTP 服务。
func main() {
	cfg := config.Load()

	if err := os.MkdirAll("data", 0755); err != nil {
		panic(err)
	}

	model.InitDB(cfg)
	if err := model.DB.AutoMigrate(&model.User{}, &model.Event{}, &model.EventParticipant{}, &model.OperationLog{}, &model.Notification{}); err != nil {
		panic(err)
	}

	engine := router.SetupRouter(cfg)
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	go startReminderJob()

	_ = engine.Run(":8080")
}

// startReminderJob 每分钟生成 15 分钟内即将开始的日程提醒通知。
func startReminderJob() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		_ = service.GenerateReminderNotifications(time.Now())
	}
}
