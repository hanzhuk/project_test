// Package errors 定义脚手架内部统一错误类型。
// 所有模块在遇到可分类的错误时，应返回 ScaffoldError，
// 以便上层根据错误码进行统一处理和友好的中文提示。
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode 是脚手架内部错误码的类型别名。
// 错误码与接口设计说明书 3.9.2 节定义的错误分类对应。
type ErrorCode int

// 预定义错误码常量，对应接口文档中的错误分类表。
const (
	CodeValidation          ErrorCode = 1001 // 参数校验失败
	CodeNotFound            ErrorCode = 1002 // 资源不存在
	CodeAlreadyExists       ErrorCode = 1003 // 资源已存在
	CodeAuth                ErrorCode = 1004 // 认证失败
	CodePermission          ErrorCode = 1005 // 权限不足
	CodeInternal            ErrorCode = 1006 // 服务器内部错误
	CodeDatabase            ErrorCode = 1007 // 数据库操作失败
	CodeOllamaConnection    ErrorCode = 1008 // Ollama 连接失败
	CodeModelUnavailable    ErrorCode = 1009 // 模型不可用
	CodeInvalidOutput       ErrorCode = 1010 // 模型输出格式错误
)

// ScaffoldError 为脚手架内部统一错误类型。
// Code 标识错误分类，Message 为面向用户的错误描述，
// Details 存放详细错误信息列表，Cause 保留原始错误便于排查。
type ScaffoldError struct {
	Code    ErrorCode // 错误码，与接口文档错误分类对应
	Message string    // 错误描述
	Details []string  // 详细错误信息
	Cause   error     // 原始错误
}

// Error 实现 error 接口，返回格式化的错误字符串。
// 当存在原始错误时，会附加原始错误信息便于调试。
func (e *ScaffoldError) Error() string {
	// 拼接错误码与描述
	msg := fmt.Sprintf("[%d] %s", e.Code, e.Message)
	// 追加详细错误信息
	for _, d := range e.Details {
		msg += "\n  - " + d
	}
	// 附加原始错误
	if e.Cause != nil {
		msg += fmt.Sprintf("\n  原因: %v", e.Cause)
	}
	return msg
}

// Unwrap 返回原始错误，支持 errors.Is / errors.As 链式判断。
func (e *ScaffoldError) Unwrap() error {
	return e.Cause
}

// WithDetails 向错误追加详细错误信息，返回错误自身便于链式调用。
func (e *ScaffoldError) WithDetails(details ...string) *ScaffoldError {
	e.Details = append(e.Details, details...)
	return e
}

// NewScaffoldError 创建统一错误实例。
// code 为错误码，message 为错误描述，cause 为可选的原始错误。
func NewScaffoldError(code ErrorCode, message string, cause error) *ScaffoldError {
	return &ScaffoldError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// IsScaffoldError 判断给定错误是否为 ScaffoldError。
// 若是则返回错误实例与 true，否则返回 nil 与 false。
func IsScaffoldError(err error) (*ScaffoldError, bool) {
	// 使用 errors.As 递归查找 ScaffoldError
	var se *ScaffoldError
	if errors.As(err, &se) {
		return se, true
	}
	return nil, false
}
