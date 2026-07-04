// Package middleware 实现 HTTP 中间件。
// 包括统一错误处理、JWT 认证等中间件。
package middleware

import (
	"log/slog"
	"net/http"

	"github.com/example/demo-api/internal/response"

	"github.com/labstack/echo/v4"
)

// ErrorHandler 是 Echo 的统一错误处理函数。
// 它将不同类型的错误转换为统一的 JSON 错误响应格式。
func ErrorHandler(err error, c echo.Context) {
	// 记录错误日志
	slog.ErrorContext(c.Request().Context(), "请求处理错误",
		slog.Any("err", err),
		slog.String("path", c.Path()),
		slog.String("method", c.Request().Method),
	)

	// 默认内部错误
	code := 1006
	status := http.StatusInternalServerError
	message := "服务器内部错误"

	// 根据 HTTP 状态码映射业务错误码
	if he, ok := err.(*echo.HTTPError); ok {
		status = he.Code
		switch he.Code {
		case http.StatusBadRequest:
			code = 1001
			message = "请求参数错误"
		case http.StatusUnauthorized:
			code = 1004
			message = "认证失败"
		case http.StatusForbidden:
			code = 1005
			message = "权限不足"
		case http.StatusNotFound:
			code = 1002
			message = "资源不存在"
		case http.StatusConflict:
			code = 1003
			message = "资源已存在"
		}
		if msg, ok := he.Message.(string); ok && msg != "" {
			message = msg
		}
	}

	// 发送错误响应（若响应已提交则无法写入）
	if !c.Response().Committed {
		if err := c.JSON(status, response.ErrorWithDetails(code, message, nil)); err != nil {
			slog.Error("发送错误响应失败", slog.Any("err", err))
		}
	}
}
