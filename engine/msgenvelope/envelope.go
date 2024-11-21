// Package msgenvelope
// @Title  数据信封
// @Description  用于不同service之间的数据传递
// @Author  yr  2024/9/2 下午3:40
// @Update  yr  2024/9/2 下午3:40
package msgenvelope

import (
	"errors"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/dto"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/pool"
	"github.com/njtc406/chaosengine/engine/utils/serializer"
	"time"
)

type MsgEnvelope struct {
	dto.DataRef
	sender       *actor.PID           // 发送者
	receiver     *actor.PID           // 接收者
	senderClient inf.IRpcSender       // 发送者客户端(用于回调)
	method       string               // 调用方法
	reqID        uint64               // 请求ID(防止重复,目前还未做防重复逻辑)
	reply        bool                 // 是否是回复
	header       dto.Header           // 消息头
	timeout      time.Duration        // 请求超时时间
	request      interface{}          // 请求参数
	response     interface{}          // 回复数据
	needResp     bool                 // 是否需要回复
	err          error                // 错误
	callbacks    []dto.CompletionFunc // 完成回调
	done         chan struct{}        // 完成信号
}

func (e *MsgEnvelope) Reset() {
	e.sender = nil
	e.receiver = nil
	e.senderClient = nil
	e.method = ""
	e.reqID = 0
	e.reply = false
	e.header = nil
	e.timeout = 0
	e.request = nil
	e.response = nil
	e.needResp = false
	if e.done == nil {
		e.done = make(chan struct{}, 1)
	}
	if len(e.done) > 0 {
		<-e.done
	}
	e.err = nil
	e.callbacks = e.callbacks[:0]
}

func (e *MsgEnvelope) SetHeaders(header dto.Header) {
	for k, v := range header {
		e.SetHeader(k, v)
	}
}

func (e *MsgEnvelope) SetHeader(key string, value string) {
	e.header.Set(key, value)
}

func (e *MsgEnvelope) SetSender(sender *actor.PID) {
	e.sender = sender
}

func (e *MsgEnvelope) SetReceiver(receiver *actor.PID) {
	e.receiver = receiver
}

func (e *MsgEnvelope) SetSenderClient(client inf.IRpcSender) {
	e.senderClient = client
}

func (e *MsgEnvelope) SetMethod(method string) {
	e.method = method
}

func (e *MsgEnvelope) SetReqId(reqId uint64) {
	e.reqID = reqId
}

func (e *MsgEnvelope) SetReply() {
	e.reply = true
}

func (e *MsgEnvelope) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

func (e *MsgEnvelope) SetRequest(req interface{}) {
	e.request = req
}

func (e *MsgEnvelope) SetResponse(res interface{}) {
	e.response = res
}

func (e *MsgEnvelope) SetError(err error) {
	e.err = err
}

func (e *MsgEnvelope) SetErrStr(err string) {
	if err != "" {
		return
	}

	e.err = errors.New(err)
}

func (e *MsgEnvelope) SetNeedResponse(need bool) {
	e.needResp = need
}

func (e *MsgEnvelope) SetCallback(cbs []dto.CompletionFunc) {
	e.callbacks = append(e.callbacks, cbs...)
}

func (e *MsgEnvelope) GetHeader(key string) string {
	return e.header.Get(key)
}

func (e *MsgEnvelope) GetHeaders() dto.Header {
	return e.header
}

func (e *MsgEnvelope) GetSender() *actor.PID {
	return e.sender
}

func (e *MsgEnvelope) GetReceiver() *actor.PID {
	return e.receiver
}

func (e *MsgEnvelope) GetSenderClient() inf.IRpcSender {
	return e.senderClient
}

func (e *MsgEnvelope) GetMethod() string {
	return e.method
}

func (e *MsgEnvelope) GetReqId() uint64 {
	return e.reqID
}

func (e *MsgEnvelope) GetRequest() interface{} {
	return e.request
}

func (e *MsgEnvelope) GetResponse() interface{} {
	return e.response
}

func (e *MsgEnvelope) GetError() error {
	return e.err
}

func (e *MsgEnvelope) GetErrStr() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *MsgEnvelope) GetTimeout() time.Duration {
	return e.timeout
}

func (e *MsgEnvelope) NeedCallback() bool {
	return len(e.callbacks) > 0
}

func (e *MsgEnvelope) IsReply() bool {
	return e.reply
}

func (e *MsgEnvelope) NeedResponse() bool {
	return e.needResp
}

func (e *MsgEnvelope) Done() {
	if e.done != nil {
		e.done <- struct{}{}
	} else {
		log.SysLogger.Warn("=================envelope done is nil===================")
	}
}

func (e *MsgEnvelope) RunCompletions() {
	for _, cb := range e.callbacks {
		cb(e.response, e.err)
	}
}

func (e *MsgEnvelope) Wait() {
	<-e.done
}

func (e *MsgEnvelope) ToProtoMsg() *actor.Message {
	msg := &actor.Message{
		TypeId:        0, // 默认使用protobuf(后面有其他需求再修改这里)
		TypeName:      "",
		Sender:        e.sender,
		Receiver:      e.receiver,
		Method:        e.method,
		Request:       nil,
		Response:      nil,
		Err:           e.GetErrStr(),
		MessageHeader: e.header,
		Reply:         e.reply,
		ReqId:         e.reqID,
		NeedResp:      e.needResp,
	}

	var byteData []byte
	var typeName string
	var err error

	if e.request != nil {
		byteData, typeName, err = serializer.Serialize(e.request, msg.TypeId)
		if err != nil {
			log.SysLogger.Errorf("serialize message[%+v] is error: %s", e, err)
			return nil
		}
		msg.Request = byteData
	} else if e.response != nil {
		byteData, typeName, err = serializer.Serialize(e.response, msg.TypeId)
		if err != nil {
			log.SysLogger.Errorf("serialize message[%+v] is error: %s", e, err)
			return nil
		}
		msg.Response = byteData
	}

	msg.TypeName = typeName

	return msg
}

//======================================================

var msgEnvelopePool = pool.NewPoolEx(make(chan pool.IPoolData, 10240), func() pool.IPoolData {
	return &MsgEnvelope{}
})

// TODO 记得测试资源释放

func NewMsgEnvelope() *MsgEnvelope {
	return msgEnvelopePool.Get().(*MsgEnvelope)
}

func ReleaseMsgEnvelope(envelope inf.IEnvelope) {
	if envelope != nil {
		msgEnvelopePool.Put(envelope.(*MsgEnvelope))
	}
}
