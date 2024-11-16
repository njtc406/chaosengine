package httplib

import (
	"github.com/gin-gonic/gin"
	"github.com/njtc406/chaosengine/engine/errdef/errcode"
	"github.com/njtc406/chaosengine/engine/utils/log"

	"net/http"
)

type IResponse interface {
	SetStatus(int) IResponse
	SetCode(interface{}) IResponse
	SetMessage(string) IResponse
	SetData(interface{}) IResponse
	Success() bool
	DoResponse(*gin.Context)
	DoAbort(*gin.Context)
}

type Response struct {
	Status int         `json:"-"`              // 返回状态
	Code   int32       `json:"code"`           // 返回码
	Msg    string      `json:"msg"`            // 返回消息
	Data   interface{} `json:"data,omitempty"` // 返回数据
}

func NewResponse() *Response {
	return &Response{
		Status: http.StatusOK,
		Code:   errcode.Ok,
		Msg:    "ok",
	}
}

// SetStatus 设置返回状态码
func (r *Response) SetStatus(status int) IResponse {
	r.Status = status
	return r
}

// SetCode 设置返回错误码
func (r *Response) SetCode(code interface{}) IResponse {
	r.Code = code.(int32)
	return r
}

// SetMessage 设置返回消息
func (r *Response) SetMessage(msg string) IResponse {
	r.Msg = msg
	return r
}

// SetData 设置返回数据
func (r *Response) SetData(data interface{}) IResponse {
	r.Data = data
	return r
}

// Success 判断返回错误码是否成功
func (r *Response) Success() bool {
	return r.Code == errcode.Ok
}

// SuccessStatus 判断返回状态码是否成功
func (r *Response) SuccessStatus() bool {
	return r.Status == http.StatusOK
}

// DoResponse 执行响应
func (r *Response) DoResponse(c *gin.Context) {
	if !r.Success() {
		log.SysLogger.Warnf("res:%#v", r)
	}

	//c.Header("Accept", "application/json")
	c.JSON(r.Status, r)
}

func (r *Response) DoAbort(c *gin.Context) {
	if !r.Success() {
		log.SysLogger.Warnf("res:%#v", r)
	}
	c.AbortWithStatusJSON(r.Status, r)
}
