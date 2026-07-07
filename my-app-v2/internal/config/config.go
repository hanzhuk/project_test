// Package config 负责生成项目的运行时配置管理。
// 通过环境变量加载配置，支持数据库连接、服务端口、JWT 密钥等配置项。
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 定义生成项目的运行时配置。
// 所有配置项通过环境变量注入，支持 .env 文件加载。
type Config struct {
	// 数据库配置
	DBHost     string // 数据库主机地址
	DBPort     string // 数据库端口
	DBUser     string // 数据库用户名
	DBPassword string // 数据库密码
	DBName     string // 数据库名称
	DBSSLMode  string // SSL 模式

	// 服务配置
	Port string // HTTP 服务端口

	// JWT 配置（启用 JWT 时使用）
	JWTSecret      string // JWT 签名密钥
	JWTExpireHours int    // JWT 过期时间（小时）
}

// Load 从环境变量加载配置。
// 未设置的环境变量使用默认值。
func Load() (*Config, error) {
	cfg := &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "secret"),
		DBName:         getEnv("DB_NAME", "my-app-v2_db"),
		DBSSLMode:      getEnv("DB_SSL_MODE", "disable"),
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
	}

	// 校验必要配置
	if cfg.JWTSecret == "" && false {
		// JWT 启用时密钥不能为空
		cfg.JWTSecret = "default-secret-please-change-in-production"
	}

	return cfg, nil
}

// DSN 返回 PostgreSQL 数据库连接字符串。
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

// getEnv 读取环境变量，不存在时返回默认值。
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt 读取环境变量并转为整数，不存在或转换失败时返回默认值。
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}
