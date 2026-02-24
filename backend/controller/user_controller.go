package controller

import (
	"strings"

	"smartcalendar/model"

	"github.com/gin-gonic/gin"
)

// UserController 负责用户资料与检索接口。
type UserController struct{}

// UpdateProfileRequest 表示资料更新请求参数。
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,min=1,max=50"`
	Avatar   string `json:"avatar" binding:"omitempty,max=500"`
}

// GetProfile 返回当前登录用户资料。
func (u UserController) GetProfile(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		Error(c, 40102, "Token 无效或已过期")
		return
	}
	Success(c, userValue)
}

// UpdateProfile 更新昵称或头像信息。
func (u UserController) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败："+err.Error())
		return
	}
	userValue, exists := c.Get("user")
	if !exists {
		Error(c, 40102, "Token 无效或已过期")
		return
	}
	user := userValue.(model.User)
	updates := map[string]interface{}{}
	if strings.TrimSpace(req.Nickname) != "" {
		updates["nickname"] = strings.TrimSpace(req.Nickname)
	}
	if strings.TrimSpace(req.Avatar) != "" {
		updates["avatar"] = strings.TrimSpace(req.Avatar)
	}
	if len(updates) == 0 {
		Success(c, user)
		return
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	if err := model.DB.First(&user, user.ID).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	Success(c, user)
}

// SearchUsers 按关键词搜索用户。
func (u UserController) SearchUsers(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		Error(c, 40001, "参数校验失败：keyword 不能为空")
		return
	}
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	offset := (page - 1) * pageSize

	var users []model.User
	query := model.DB.Model(&model.User{}).Where("nickname LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	if err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
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
