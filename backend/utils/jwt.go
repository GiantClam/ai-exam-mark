package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 从环境变量获取JWT密钥，如果不存在则使用默认值
// 在生产环境中应始终设置环境变量JWT_SECRET
func getJWTKey() []byte {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// 默认密钥仅用于开发和测试环境
		// 实际生产环境中，如果未设置环境变量应该抛出错误
		jwtSecret = "homework_marking_default_dev_key"
		fmt.Println("警告: 未设置JWT_SECRET环境变量，使用默认开发密钥")
	}
	return []byte(jwtSecret)
}

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
	tokenString, err := token.SignedString(getJWTKey())

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
		return getJWTKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("令牌无效")
	}

	return claims, nil
}
