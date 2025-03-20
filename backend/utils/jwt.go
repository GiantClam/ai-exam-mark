package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 密钥，实际应用中应存储在安全的环境变量或配置中
var jwtKey = []byte("homework_marking_secret_key")

// 用户声明结构体
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成JWT令牌
func GenerateJWT(userID, role string) (string, error) {
	// 设置token过期时间为24小时
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建JWT声明
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "homework_marking_system",
			Subject:   userID,
		},
	}

	// 创建并签名token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseJWT 解析JWT令牌
func ParseJWT(tokenString string) (*Claims, error) {
	// 解析JWT
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("令牌无效")
	}

	return claims, nil
}
