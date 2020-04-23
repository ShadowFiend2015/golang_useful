package aicode

import (
	"fmt"
)

var (
	ComInnerError     = errorPair(90001, "内部错误")
	ComNotExist       = errorPair(90002, "接口不存在")
	ComUnAuthorized   = errorPair(90003, "未鉴权")
	ComAuthFailed     = errorPair(90004, "鉴权失败")
	ComBadParam       = errorPair(90005, "请求参数错误")
	ComSupportScheme  = errorPair(90006, "不支持的流协议")
	ComLimit          = errorPair(90007, "超过限制")
	ComDuplicate      = errorPair(90008, "重复操作")
	ComEntityTooLarge = errorPair(90009, "请求体超过最大限制")
	ComMissSid        = errorPair(90010, "sid缺失")
	ComAuthExpired    = errorPair(90011, "鉴权已过期")
	ComDataInvalid    = errorPair(90012, "数据非法")
)

type (
	// HTTPError define
	HTTPError interface {
		Code() int
		Msg() string
		SetMsg(string) HTTPError
		Error() string
	}

	// BaseError http error
	BaseError struct {
		ErrCode int    `json:"code"`
		ErrMsg  string `json:"msg"`
	}
)

func NewHTTPError(code int, msg string) HTTPError {
	he := &BaseError{ErrCode: code, ErrMsg: msg}
	return he
}

func (e *BaseError) Code() int {
	return e.ErrCode
}

func (e *BaseError) Msg() string {
	return e.ErrMsg
}

func (e *BaseError) SetMsg(m string) HTTPError {
	nr := e
	nr.ErrMsg = m
	return nr
}

func (e *BaseError) Error() string {
	return fmt.Sprintf("Code %d, Msg %s", e.ErrCode, e.ErrMsg)
}

var errorMap = make(map[int]string)

func errorPair(code int, desc string) HTTPError {
	if v, ok := errorMap[code]; ok {
		panic("error code exit, desc : " + v)
	} else {
		errorMap[code] = desc
		return &BaseError{code, desc}
	}
}

func CodeToError(code int) HTTPError {
	if v, ok := errorMap[code]; ok {
		return &BaseError{code, v}
	} else {
		panic("error code not exit, desc : " + string(code))
	}
}
