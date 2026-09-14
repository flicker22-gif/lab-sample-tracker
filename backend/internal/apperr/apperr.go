// Package apperr 定义应用级错误：携带面向用户的可读消息与错误类别，
// 由 service 层产生，handler 层统一映射为 HTTP 状态码与统一响应体。
package apperr

import (
	"errors"
	"net/http"
)

// Kind 错误类别，决定 HTTP 状态码
type Kind int

const (
	KindInternal   Kind = iota // 服务器内部错误（不向客户端暴露细节）
	KindBadRequest             // 请求参数或业务前置条件不满足
	KindNotFound               // 资源不存在
	KindConflict               // 资源状态冲突（如机台被占用）
)

// Error 应用错误。Message 可安全返回给客户端；Err 为内部原因，仅记录日志。
type Error struct {
	Kind    Kind
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

// BadRequest 参数/前置条件错误（400）
func BadRequest(msg string) *Error { return &Error{Kind: KindBadRequest, Message: msg} }

// NotFound 资源不存在（404）
func NotFound(msg string) *Error { return &Error{Kind: KindNotFound, Message: msg} }

// Conflict 状态冲突（409）
func Conflict(msg string) *Error { return &Error{Kind: KindConflict, Message: msg} }

// Internal 内部错误（500），客户端只看到通用提示，真实原因进日志
func Internal(err error) *Error {
	return &Error{Kind: KindInternal, Message: "服务器内部错误", Err: err}
}

// Wrap 给内部错误补充上下文说明（仍然按 500 处理）
func Wrap(err error, ctx string) *Error {
	return &Error{Kind: KindInternal, Message: "服务器内部错误", Err: errors.New(ctx + ": " + err.Error())}
}

// KindOf 提取错误类别；非应用错误一律按内部错误处理
func KindOf(err error) Kind {
	var ae *Error
	if errors.As(err, &ae) {
		return ae.Kind
	}
	return KindInternal
}

// MessageOf 提取可返回给客户端的消息；内部错误返回通用提示
func MessageOf(err error) string {
	var ae *Error
	if errors.As(err, &ae) {
		return ae.Message
	}
	return "服务器内部错误"
}

// HTTPStatus 错误类别对应的 HTTP 状态码
func HTTPStatus(k Kind) int {
	switch k {
	case KindBadRequest:
		return http.StatusBadRequest
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
