package controller

import (
	"smartcalendar/model"

	"github.com/gin-gonic/gin"
)

// OperationLogController 负责操作记录查询接口。
type OperationLogController struct{}

// ListLogs 分页查询操作记录。
func (o OperationLogController) ListLogs(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	action := c.Query("action")
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	offset := (page - 1) * pageSize

	query := model.DB.Model(&model.OperationLog{}).Where("user_id = ?", user.ID)
	if action != "" {
		query = query.Where("action = ?", action)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}

	var logs []model.OperationLog
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}

	Success(c, gin.H{
		"list":      logs,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}
