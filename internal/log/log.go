// Package log 封装脚手架自身的结构化日志功能。
// 基于 Go 标准库 log/slog 实现，支持文本和 JSON 两种格式，
// 用于在脚手架 CLI 运行过程中输出调试、信息和错误日志。
package log

import (
	"log/slog"
	"os"
	"strings"
)

// Init 初始化脚手架的全局日志系统。
// level 参数支持 "debug"、"info"、"warn"、"error"，
// 不区分大小写；不识别时默认使用 info 级别。
// 日志输出到标准错误，采用文本格式以便终端阅读。
func Init(level string) {
	// 将字符串级别映射为 slog.Level
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn", "warning":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	// 创建文本格式 handler，输出到 stderr
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lv})
	slog.SetDefault(slog.New(handler))
}

// Debug 输出调试级别日志。
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

// Info 输出信息级别日志。
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Warn 输出警告级别日志。
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// Error 输出错误级别日志。
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}
