package controller

import (
	"path/filepath"
	"strings"

	"smartcalendar/config"
	"smartcalendar/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadController 负责头像上传接口。
type UploadController struct {
	Cfg config.AppConfig
}

// UploadAvatar 保存头像并返回可访问地址。
func (u UploadController) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Error(c, 40001, "参数校验失败：请上传文件")
		return
	}
	if file.Size > 10*1024*1024 {
		Error(c, 40001, "参数校验失败：文件过大")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".png"
	}
	objectKey := strings.Trim(u.Cfg.TOSAvatarPrefix, "/") + "/" + uuid.NewString() + "_" + sanitizeFilename(file.Filename)
	src, err := file.Open()
	if err != nil {
		Error(c, 50000, "文件上传失败")
		return
	}
	defer src.Close()
	url, err := service.UploadToTOS(c.Request.Context(), u.Cfg, objectKey, src)
	if err != nil {
		Error(c, 50000, "文件上传失败")
		return
	}
	Success(c, gin.H{
		"url": url,
	})
}

// sanitizeFilename 过滤文件名中的不安全字符。
func sanitizeFilename(name string) string {
	replaced := strings.ReplaceAll(name, " ", "_")
	replaced = strings.ReplaceAll(replaced, "..", "")
	return replaced
}
