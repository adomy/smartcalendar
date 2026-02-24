// Package model 定义数据库模型与初始化逻辑。
package model

import "time"

// Event 表示日程实体。
type Event struct {
	ID           uint               `gorm:"primaryKey" json:"id"`
	UserID       uint               `gorm:"index;not null" json:"user_id"`
	Title        string             `gorm:"size:100;not null" json:"title"`
	Type         string             `gorm:"size:20;not null" json:"type"`
	StartTime    time.Time          `gorm:"not null" json:"start_time"`
	EndTime      time.Time          `gorm:"not null" json:"end_time"`
	Location     string             `gorm:"size:200" json:"location"`
	Description  string             `gorm:"size:500" json:"description"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Creator      User               `gorm:"foreignKey:UserID" json:"creator,omitempty"`
	Participants []EventParticipant `gorm:"foreignKey:EventID" json:"participants,omitempty"`
}

// EventParticipant 表示日程参与人关系。
type EventParticipant struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	EventID uint `gorm:"index;not null" json:"event_id"`
	UserID  uint `gorm:"index;not null" json:"user_id"`
	User    User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// OperationLog 表示用户操作记录。
type OperationLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	Action      string    `gorm:"size:20;not null" json:"action"`
	TargetTitle string    `gorm:"size:100" json:"target_title"`
	Detail      string    `gorm:"type:text" json:"detail"`
	CreatedAt   time.Time `json:"created_at"`
}

// Notification 表示用户通知记录。
type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Type      string    `gorm:"size:30;not null" json:"type"`
	Content   string    `gorm:"size:500;not null" json:"content"`
	EventID   uint      `gorm:"index" json:"event_id"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
