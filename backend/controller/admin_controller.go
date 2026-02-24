package controller

import (
	"strings"

	"smartcalendar/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminController 负责管理员接口。
type AdminController struct{}

// UpdateUserStatusRequest 表示用户状态更新请求。
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// ListUsers 分页查询用户列表。
func (a AdminController) ListUsers(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	offset := (page - 1) * pageSize

	var total int64
	if err := model.DB.Model(&model.User{}).Count(&total).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	var users []model.User
	if err := model.DB.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	Success(c, gin.H{
		"list":      users,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// UpdateUserStatus 更新用户启用/禁用状态。
func (a AdminController) UpdateUserStatus(c *gin.Context) {
	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败："+err.Error())
		return
	}
	status := strings.TrimSpace(req.Status)
	if status != "active" && status != "disabled" {
		Error(c, 40001, "参数校验失败：status 无效")
		return
	}
	var user model.User
	if err := model.DB.First(&user, c.Param("id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, 40401, "资源不存在")
			return
		}
		Error(c, 50000, "服务器内部错误")
		return
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", user.ID).Update("status", status).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	if err := model.DB.First(&user, user.ID).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	Success(c, user)
}

// ResetPassword 重置用户密码为默认值。
func (a AdminController) ResetPassword(c *gin.Context) {
	var user model.User
	if err := model.DB.First(&user, c.Param("id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, 40401, "资源不存在")
			return
		}
		Error(c, 50000, "服务器内部错误")
		return
	}
	newPassword := "Smart@123"
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", user.ID).Update("password", string(hashed)).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	Success(c, gin.H{
		"user_id":      user.ID,
		"new_password": newPassword,
	})
}
