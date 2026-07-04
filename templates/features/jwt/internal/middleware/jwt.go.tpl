// Package middleware - JWT 认证中间件
// 该文件在启用 JWT 功能时生成，提供 JWT token 的签发与校验。
package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"{{.ModulePath}}/internal/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTClaims 定义 JWT 的声明结构。
type JWTClaims struct {
	UserID int    `json:"user_id"` // 用户 ID
	Email  string `json:"email"`   // 用户邮箱
	jwt.RegisteredClaims
}

// JWTConfig 保存 JWT 配置。
type JWTConfig struct {
	Secret      string        // 签名密钥
	ExpireHours time.Duration // 过期时间（小时）
}

// GenerateToken 生成 JWT token。
// userID 为用户 ID，email 为用户邮箱。
func GenerateToken(cfg *JWTConfig, userID int, email string) (string, error) {
	// 构建声明
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.ExpireHours * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	// 使用 HS256 签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken 解析并校验 JWT token。
func ParseToken(cfg *JWTConfig, tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// JWTAuth 是 JWT 认证中间件。
// 从 Authorization 头提取 Bearer token 并校验，通过后将用户信息存入上下文。
func JWTAuth(cfg *JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 提取 Authorization 头
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, response.Error(1004, "未提供认证信息"))
			}
			// 校验 Bearer 前缀
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, response.Error(1004, "认证格式错误"))
			}
			// 解析 token
			claims, err := ParseToken(cfg, parts[1])
			if err != nil {
				slog.WarnContext(c.Request().Context(), "JWT 校验失败", slog.Any("err", err))
				return c.JSON(http.StatusUnauthorized, response.Error(1004, "认证失败，token 无效或已过期"))
			}
			// 将用户信息存入上下文
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			return next(c)
		}
	}
}
