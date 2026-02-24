package controller

import (
	"strconv"

	"smartcalendar/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NotificationController 负责通知查询与已读状态更新。
type NotificationController struct{}

// ListNotifications 分页查询通知列表。
func (n NotificationController) ListNotifications(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	isReadQuery := c.Query("is_read")
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	offset := (page - 1) * pageSize

	query := model.DB.Model(&model.Notification{}).Where("user_id = ?", user.ID)
	if isReadQuery != "" {
		value, err := strconv.ParseBool(isReadQuery)
		if err != nil {
			Error(c, 40001, "参数校验失败：is_read 无效")
			return
		}
		query = query.Where("is_read = ?", value)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}

	var list []model.Notification
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}

	Success(c, gin.H{
		"list":      list,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// UnreadCount 返回未读通知数量。
func (n NotificationController) UnreadCount(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	var count int64
	if err := model.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", user.ID, false).Count(&count).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	Success(c, gin.H{"count": count})
}

// MarkRead 标记单条通知为已读。
func (n NotificationController) MarkRead(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	var notification model.Notification
	if err := model.DB.Where("user_id = ?", user.ID).First(&notification, c.Param("id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, 40401, "资源不存在")
			return
		}
		Error(c, 50000, "服务器内部错误")
		return
	}
	if notification.IsRead {
		Success(c, notification)
		return
	}
	if err := model.DB.Model(&model.Notification{}).Where("id = ?", notification.ID).Update("is_read", true).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	if err := model.DB.First(&notification, notification.ID).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	Success(c, notification)
}

// MarkAllRead 标记全部通知为已读。
func (n NotificationController) MarkAllRead(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	result := model.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", user.ID, false).Update("is_read", true)
	if result.Error != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	Success(c, gin.H{"updated": result.RowsAffected})
}
