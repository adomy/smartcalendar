// Package model 定义数据库模型与初始化逻辑。
package model

import "time"

// User 表示系统用户实体。
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Nickname  string    `gorm:"size:50;not null" json:"nickname"`
	Email     string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Avatar    string    `gorm:"size:500" json:"avatar"`
	Role      string    `gorm:"size:20;default:user" json:"role"`
	Status    string    `gorm:"size:20;default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
