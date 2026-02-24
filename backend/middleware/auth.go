package middleware

import (
	"errors"
	"net/http"
	"strings"

	"smartcalendar/config"
	"smartcalendar/model"
	"smartcalendar/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthRequired 校验 JWT 并注入用户上下文。
func AuthRequired(cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearer(c.GetHeader("Authorization"))
		if tokenString == "" {
			c.JSON(http.StatusOK, gin.H{"code": 40101, "message": "未登录或 Token 缺失", "data": nil})
			c.Abort()
			return
		}
		claims, err := service.ParseToken(cfg, tokenString)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 40102, "message": "Token 无效或已过期", "data": nil})
			c.Abort()
			return
		}
		var user model.User
		if err := model.DB.First(&user, claims.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"code": 40102, "message": "Token 无效或已过期", "data": nil})
				c.Abort()
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 50000, "message": "服务器内部错误", "data": nil})
			c.Abort()
			return
		}
		if user.Status == "disabled" {
			c.JSON(http.StatusOK, gin.H{"code": 40301, "message": "账号已被禁用", "data": nil})
			c.Abort()
			return
		}
		c.Set("userID", user.ID)
		c.Set("role", user.Role)
		c.Set("user", user)
		c.Next()
	}
}

// AdminRequired 限制管理员访问。
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists || roleValue != "admin" {
			c.JSON(http.StatusOK, gin.H{"code": 40301, "message": "无权限", "data": nil})
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractBearer 提取 Authorization: Bearer <token>。
func extractBearer(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}
