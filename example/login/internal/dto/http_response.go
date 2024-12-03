// Package dto
// 模块名: 模块名
// 功能描述: 描述
// 作者:  yr  2024/11/16 0016 18:46
// 最后更新:  yr  2024/11/16 0016 18:46
package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/njtc406/chaosengine/engine/dto"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/pool"
	"github.com/njtc406/chaosengine/example/login/internal/def"
	msg "github.com/njtc406/chaosengine/example/msg/comm"
	"net/http"
)

var respPool = pool.NewPoolEx(make(chan pool.IPoolData, 2048), func() pool.IPoolData {
	return &HttpResponse{}
})

type HttpResponse struct {
	dto.DataRef `json:"-"`
	Status      int         `json:"-"`              // 返回状态
	Code        int32       `json:"code"`           // 返回码
	Msg         string      `json:"msg"`            // 返回消息
	Data        interface{} `json:"data,omitempty"` // 返回数据
}

func NewResponse() *HttpResponse {
	return respPool.Get().(*HttpResponse)
}

func Release(resp *HttpResponse) {
	respPool.Put(resp)
}

func (r *HttpResponse) Reset() {
	r.Status = http.StatusOK
	r.Code = def.DefaultHttpRespCode
	r.Msg = def.DefaultHttpRespMsg
	r.Data = nil
}

// SetStatus 设置返回状态码
func (r *HttpResponse) SetStatus(status int) *HttpResponse {
	r.Status = status
	return r
}

// SetCode 设置返回错误码
func (r *HttpResponse) SetCode(code msg.ErrCode) *HttpResponse {
	r.Code = int32(code)
	return r
}

// SetMessage 设置返回消息
func (r *HttpResponse) SetMessage(msg string) *HttpResponse {
	r.Msg = msg
	return r
}

// SetData 设置返回数据
func (r *HttpResponse) SetData(data interface{}) *HttpResponse {
	r.Data = data
	return r
}

// Success 判断返回错误码是否成功
func (r *HttpResponse) Success() bool {
	return r.Code == def.DefaultHttpRespCode
}

// SuccessStatus 判断返回状态码是否成功
func (r *HttpResponse) SuccessStatus() bool {
	return r.Status == http.StatusOK
}

// DoResponse 执行响应
func (r *HttpResponse) DoResponse(c *gin.Context) {
	//if !r.Success() {
	log.SysLogger.Warnf("res:%#v", r)
	//}

	//c.Header("Accept", "application/json")
	c.JSON(r.Status, r)
	Release(r)
}

func (r *HttpResponse) DoAbort(c *gin.Context) {
	if !r.Success() {
		log.SysLogger.Warnf("res:%#v", r)
	}
	c.AbortWithStatusJSON(r.Status, r)
	Release(r)
}
