// Package model 定义数据库模型与初始化逻辑。
package model

import (
	"log"

	"smartcalendar/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB 为全局数据库连接实例。
var DB *gorm.DB

// InitDB 使用配置中的 SQLite 路径初始化数据库连接。
func InitDB(cfg config.AppConfig) {
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	DB = db
}
