package config

import (
	"os"
	"strconv"
)

// AppConfig 统一管理服务启动所需配置。
type AppConfig struct {
	JWTSecret          string
	TokenExpireHours   int
	DBPath             string
	UploadAvatarDir    string
	UploadAvatarPrefix string
	CorsAllowOrigin    string
	ArkAPIKey          string
	ArkModelID         string
	ArkBaseURL         string
	ArkRegion          string
	ArkAccessKey       string
	ArkSecretKey       string
}

// Load 从环境变量读取配置并提供默认值。
func Load() AppConfig {
	return AppConfig{
		JWTSecret:          getEnv("JWT_SECRET", "smartcalendar-secret"),
		TokenExpireHours:   getEnvInt("TOKEN_EXPIRE_HOURS", 168),
		DBPath:             getEnv("DB_PATH", "data/smartcalendar.db"),
		UploadAvatarDir:    getEnv("UPLOAD_AVATAR_DIR", "upload/avatars"),
		UploadAvatarPrefix: getEnv("UPLOAD_AVATAR_PREFIX", "/upload/avatars"),
		CorsAllowOrigin:    getEnv("CORS_ALLOW_ORIGIN", "http://localhost:5173"),
		ArkAPIKey:          getEnv("ARK_API_KEY", ""),
		ArkModelID:         getEnv("ARK_MODEL_ID", ""),
		ArkBaseURL:         getEnv("ARK_BASE_URL", ""),
		ArkRegion:          getEnv("ARK_REGION", ""),
		ArkAccessKey:       getEnv("ARK_ACCESS_KEY", ""),
		ArkSecretKey:       getEnv("ARK_SECRET_KEY", ""),
	}
}

// getEnv 读取字符串环境变量。
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 读取整数环境变量。
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
