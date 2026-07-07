// Package response 定义统一的 HTTP 响应格式。
// 所有接口返回统一的 JSON 结构，包含 code、message 和 data 字段。
package response

import "github.com/labstack/echo/v4"

// Response 是统一的响应结构。
type Response struct {
	Code    int    `json:"code"`           // 业务错误码，0 表示成功
	Message string `json:"message"`        // 响应消息
	Data    any    `json:"data,omitempty"` // 响应数据
}

// ErrorResponse 是错误响应结构。
type ErrorResponse struct {
	Code    int      `json:"code"`              // 业务错误码
	Message string   `json:"message"`           // 错误描述
	Details []string `json:"details,omitempty"` // 详细错误信息
}

// Success 返回成功响应。
func Success(data any) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// SuccessMessage 返回带自定义消息的成功响应。
func SuccessMessage(message string, data any) Response {
	return Response{
		Code:    0,
		Message: message,
		Data:    data,
	}
}

// Error 返回错误响应。
func Error(code int, message string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// ErrorWithDetails 返回带详细信息的错误响应。
func ErrorWithDetails(code int, message string, details []string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// JSON 发送统一格式的 JSON 响应。
func JSON(c echo.Context, status int, data any) error {
	return c.JSON(status, Success(data))
}
