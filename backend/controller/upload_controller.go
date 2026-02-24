package controller

import (
	"path/filepath"
	"strings"
	"time"

	"smartcalendar/config"

	"github.com/gin-gonic/gin"
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
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".png"
	}
	fileName := time.Now().Format("20060102_150405") + "_" + sanitizeFilename(file.Filename)
	savePath := filepath.Join(u.Cfg.UploadAvatarDir, fileName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		Error(c, 50000, "文件上传失败")
		return
	}
	Success(c, gin.H{
		"url": u.Cfg.UploadAvatarPrefix + "/" + fileName,
	})
}

// sanitizeFilename 过滤文件名中的不安全字符。
func sanitizeFilename(name string) string {
	replaced := strings.ReplaceAll(name, " ", "_")
	replaced = strings.ReplaceAll(replaced, "..", "")
	return replaced
}
