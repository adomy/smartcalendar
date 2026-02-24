package controller

import (
	"net/http"
	"strings"

	"smartcalendar/config"
	"smartcalendar/model"
	"smartcalendar/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthController 负责注册与登录相关接口。
type AuthController struct {
	Cfg config.AppConfig
}

// RegisterRequest 表示注册请求参数。
type RegisterRequest struct {
	Nickname string `json:"nickname" binding:"required,min=1,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Avatar   string `json:"avatar" binding:"max=500"`
}

// LoginRequest 表示登录请求参数。
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=1,max=50"`
}

// Register 处理用户注册并返回 Token。
func (a AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败："+err.Error())
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" {
		Error(c, 40001, "参数校验失败：昵称不能为空")
		return
	}

	var count int64
	if err := model.DB.Model(&model.User{}).Count(&count).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	if count == 0 && req.Nickname != "admin" {
		Error(c, 40001, "系统首位用户须以 admin 身份注册")
		return
	}

	var existing model.User
	if err := model.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		Error(c, 40901, "邮箱已注册")
		return
	} else if err != nil && err != gorm.ErrRecordNotFound {
		Error(c, 50000, "服务器内部错误")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}

	role := "user"
	if count == 0 {
		role = "admin"
	}
	user := model.User{
		Nickname: req.Nickname,
		Email:    req.Email,
		Password: string(hashed),
		Avatar:   req.Avatar,
		Role:     role,
		Status:   "active",
	}
	if err := model.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			Error(c, 40901, "邮箱已注册")
			return
		}
		Error(c, 50000, "服务器内部错误")
		return
	}

	token, err := service.GenerateToken(a.Cfg, user.ID, user.Role)
	if err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}

	Success(c, gin.H{
		"token": token,
		"user":  user,
	})
}

// Login 处理用户登录并返回 Token。
func (a AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败："+err.Error())
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user model.User
	if err := model.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, 40001, "邮箱或密码错误")
			return
		}
		Error(c, 50000, "服务器内部错误")
		return
	}
	if user.Status == "disabled" {
		Error(c, 40301, "账号已被禁用")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		Error(c, 40001, "邮箱或密码错误")
		return
	}

	token, err := service.GenerateToken(a.Cfg, user.ID, user.Role)
	if err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"token": token,
			"user":  user,
		},
	})
}
