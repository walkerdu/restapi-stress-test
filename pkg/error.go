package ecode

import (
	"fmt"

	"github.com/flosch/pongo2"
	// nolint
)

type APIError struct {
	ErrCode    int64             // 错误码
	ErrMsg     string            // 错误提示, 它可能是模版(format格式 hello {name}）
	ErrMsgArgs map[string]string // 如果ErrMsg模版，这是它需要的一些参数. map[name]=red
}

// Error: 构造一个APIError
// code: 业务错误码
// message: 业务内部错误描述文字
func Error(code int, message string) *APIError {
	return &APIError{
		ErrCode:    int64(code),
		ErrMsg:     message,
		ErrMsgArgs: nil,
	}

}

// Errorf 构造一个APIError
// code: 业务错误码
// format,args 用于构建message
func Errorf(code int, format string, args ...interface{}) *APIError {
	return Error(code, fmt.Sprintf(format, args...))
}

// Code: 返回业务错误码
func (a *APIError) Code() int64 {
	if a == nil {
		return 0
	}
	return a.ErrCode
}

// TipsMessage: 返回给客户端的错误描述
func (a *APIError) Message() string {
	if a == nil {
		return ""
	}

	if len(a.ErrMsgArgs) == 0 {
		return a.ErrMsg
	}

	if len(a.ErrMsg) == 0 {
		return ""
	}

	tpl, err := pongo2.FromString(a.ErrMsg)
	if err != nil {
		return a.ErrMsg
	}
	args := make(map[string]interface{}, len(a.ErrMsgArgs))
	for i, arg := range a.ErrMsgArgs {
		args[i] = arg
	}
	out, err := tpl.Execute(pongo2.Context(args))
	if err != nil {
		return a.ErrMsg
	}

	return out
}

func (a *APIError) Error() string {
	if a == nil {
		return fmt.Sprintf("no error")
	}

	return fmt.Sprintf("{code:%d, message:[%s]}", a.Code(), a.Message())
}
