package service

import (
	"time"

	"smartcalendar/config"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims 表示 JWT 中的用户信息与标准字段。
type UserClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成带过期时间的 JWT。
func GenerateToken(cfg config.AppConfig, userID uint, role string) (string, error) {
	expireAt := time.Now().Add(time.Duration(cfg.TokenExpireHours) * time.Hour)
	claims := UserClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ParseToken 解析并校验 JWT。
func ParseToken(cfg config.AppConfig, tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
