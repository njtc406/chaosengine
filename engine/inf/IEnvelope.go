// Package inf
// @Title  信封接口
// @Description  desc
// @Author  yr  2024/11/14
// @Update  yr  2024/11/14
package inf

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/dto"
	"time"
)

type IEnvelope interface {
	IDataDef
	// Set

	SetHeaders(header dto.Header)
	SetHeader(key string, value string)
	SetSender(sender *actor.PID)
	SetReceiver(receiver *actor.PID)
	SetSenderClient(client IRpcSender)
	SetMethod(method string)
	SetReqId(reqId uint64)
	SetReply()
	SetTimeout(timeout time.Duration)
	SetRequest(req interface{})
	SetResponse(res interface{})
	SetError(err error)
	SetErrStr(err string)
	SetNeedResponse(need bool)
	SetCallback(cbs []dto.CompletionFunc)

	// Get

	GetHeader(key string) string
	GetHeaders() dto.Header
	GetSender() *actor.PID
	GetReceiver() *actor.PID
	GetSenderClient() IRpcSender
	GetMethod() string
	GetReqId() uint64
	GetRequest() interface{}
	GetResponse() interface{}
	GetError() error
	GetErrStr() string
	GetTimeout() time.Duration

	// Check

	NeedCallback() bool
	IsReply() bool
	NeedResponse() bool

	// Option

	Done()
	RunCompletions()
	Wait()
	ToProtoMsg() *actor.Message
}
