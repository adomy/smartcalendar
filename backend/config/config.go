package config

import (
	"os"
	"strconv"
)

// AppConfig 统一管理服务启动所需配置。
type AppConfig struct {
	// 基础配置
	JWTSecret        string // JWT_SECRET：JWT 签名密钥（必填）
	TokenExpireHours int    // TOKEN_EXPIRE_HOURS：Token 过期小时数，默认 168
	DBPath           string // DB_PATH：SQLite 文件路径，默认 data/smartcalendar.db
	CorsAllowOrigin  string // CORS_ALLOW_ORIGIN：允许的前端域名，可用逗号分隔多个

	// Ark 大模型配置（二选一鉴权：ARK_API_KEY 或 ARK_ACCESS_KEY/ARK_SECRET_KEY）
	ArkModelID   string // ARK_MODEL_ID：模型 Endpoint ID（必填）
	ArkAPIKey    string // ARK_API_KEY：鉴权密钥
	ArkAccessKey string // ARK_ACCESS_KEY：鉴权密钥
	ArkSecretKey string // ARK_SECRET_KEY：鉴权密钥
	ArkBaseURL   string // ARK_BASE_URL：自定义 BaseURL（可选）
	ArkRegion    string // ARK_REGION：区域（可选）

	// 豆包语音识别配置
	SpeechApiKey        string // SPEECH_APP_KEY：控制台 App ID（必填）
	SpeechResourceID    string // SPEECH_RESOURCE_ID：资源 ID（必填，如 volc.seedasr.auc）
	SpeechBaseURL       string // SPEECH_BASE_URL：提交/查询地址前缀，默认 https://openspeech.bytedance.com/api/v3/auc/bigmodel
	SpeechPublicBaseURL string // SPEECH_PUBLIC_BASE_URL：对外可访问的音频文件域名（可选，TOS 场景无需填写）
	SpeechModelName     string // SPEECH_MODEL_NAME：模型名称，默认 bigmodel
	SpeechModelVersion  string // SPEECH_MODEL_VERSION：模型版本（可选）
	SpeechLanguage      string // SPEECH_LANGUAGE：指定识别语言（可选，空表示自动/多语）
	SpeechFormat        string // SPEECH_FORMAT：音频容器格式，默认 webm
	SpeechCodec         string // SPEECH_CODEC：音频编码格式，默认 opus
	SpeechRate          int    // SPEECH_RATE：采样率，默认 16000
	SpeechBits          int    // SPEECH_BITS：位深，默认 16
	SpeechChannel       int    // SPEECH_CHANNEL：声道，默认 1

	// 火山引擎 TOS 配置（必填）
	TOSAccessKey     string // TOS_ACCESS_KEY：鉴权密钥
	TOSSecretKey     string // TOS_SECRET_KEY：鉴权密钥
	TOSEndpoint      string // TOS_ENDPOINT：如 https://tos-cn-beijing.volces.com
	TOSRegion        string // TOS_REGION：如 cn-beijing
	TOSBucket        string // TOS_BUCKET：桶名
	TOSPublicBaseURL string // TOS_PUBLIC_BASE_URL：自定义公网访问域名（可选）
	TOSAvatarPrefix  string // TOS_AVATAR_PREFIX：头像对象前缀，默认 avatars
	TOSAudioPrefix   string // TOS_AUDIO_PREFIX：音频对象前缀，默认 audio
}

// Load 从环境变量读取配置并提供默认值。
func Load() AppConfig {
	return AppConfig{
		JWTSecret:        getEnv("JWT_SECRET", "smartcalendar-secret"),
		TokenExpireHours: getEnvInt("TOKEN_EXPIRE_HOURS", 168),
		DBPath:           getEnv("DB_PATH", "data/smartcalendar.db"),
		CorsAllowOrigin:  getEnv("CORS_ALLOW_ORIGIN", "http://localhost:5173"),

		ArkAPIKey:    getEnv("ARK_API_KEY", ""),
		ArkModelID:   getEnv("ARK_MODEL_ID", ""),
		ArkBaseURL:   getEnv("ARK_BASE_URL", ""),
		ArkRegion:    getEnv("ARK_REGION", ""),
		ArkAccessKey: getEnv("ARK_ACCESS_KEY", ""),
		ArkSecretKey: getEnv("ARK_SECRET_KEY", ""),

		SpeechApiKey:        getEnv("SPEECH_API_KEY", ""),
		SpeechResourceID:    getEnv("SPEECH_RESOURCE_ID", ""),
		SpeechBaseURL:       getEnv("SPEECH_BASE_URL", "https://openspeech.bytedance.com/api/v3/auc/bigmodel"),
		SpeechPublicBaseURL: getEnv("SPEECH_PUBLIC_BASE_URL", ""),
		SpeechModelName:     getEnv("SPEECH_MODEL_NAME", "bigmodel"),
		SpeechModelVersion:  getEnv("SPEECH_MODEL_VERSION", ""),
		SpeechLanguage:      getEnv("SPEECH_LANGUAGE", ""),
		SpeechFormat:        getEnv("SPEECH_FORMAT", "raw"),
		SpeechCodec:         getEnv("SPEECH_CODEC", "raw"),
		SpeechRate:          getEnvInt("SPEECH_RATE", 16000),
		SpeechBits:          getEnvInt("SPEECH_BITS", 16),
		SpeechChannel:       getEnvInt("SPEECH_CHANNEL", 1),

		TOSAccessKey:     getEnv("TOS_ACCESS_KEY", ""),
		TOSSecretKey:     getEnv("TOS_SECRET_KEY", ""),
		TOSEndpoint:      getEnv("TOS_ENDPOINT", ""),
		TOSRegion:        getEnv("TOS_REGION", ""),
		TOSBucket:        getEnv("TOS_BUCKET", ""),
		TOSPublicBaseURL: getEnv("TOS_PUBLIC_BASE_URL", ""),
		TOSAvatarPrefix:  getEnv("TOS_AVATAR_PREFIX", "avatars"),
		TOSAudioPrefix:   getEnv("TOS_AUDIO_PREFIX", "audio"),
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
