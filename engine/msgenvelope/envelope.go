// Package msgenvelope
// @Title  数据信封
// @Description  用于不同service之间的数据传递
// @Author  yr  2024/9/2 下午3:40
// @Update  yr  2024/9/2 下午3:40
package msgenvelope

import (
	"errors"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/serializer"
	"github.com/njtc406/chaosengine/engine1/synclib"
	"time"
)

type CompletionFunc func(resp interface{}, err error) // 异步回调函数

type Header map[string]string

func (header Header) Get(key string) string {
	return header[key]
}

func (header Header) Set(key string, value string) {
	header[key] = value
}

func (header Header) Keys() []string {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	return keys
}

func (header Header) Length() int {
	return len(header)
}

func (header Header) ToMap() map[string]string {
	mp := make(map[string]string)
	for k, v := range header {
		mp[k] = v
	}
	return mp
}

// MetaData 消息元数据
type MetaData struct {
	Sender       *actor.PID    // 发送者
	Receiver     *actor.PID    // 接收者
	SenderClient inf.IClient   // 发送者客户端(用于回调)
	Method       string        // 调用方法
	ReqID        uint64        // 请求ID(防止重复,目前还未做防重复逻辑)
	Reply        bool          // 是否是回复
	Header       Header        // 消息头
	Timeout      time.Duration // 请求超时时间
}

func (m *MetaData) reset() {
	m.Sender = nil
	m.Receiver = nil
	m.SenderClient = nil
	m.Method = ""
	m.ReqID = 0
	m.Reply = false
	m.Header = nil
	m.Timeout = 0
}

type MessageData struct {
	Request  interface{} // 请求参数
	Response interface{} // 回复数据
	NeedResp bool        // 是否需要回复
}

func (m *MessageData) reset() {
	m.Request = nil
	m.Response = nil
	m.NeedResp = false
}

type Completion struct {
	done      chan struct{}    // 完成信号
	Err       error            // 错误
	Callbacks []CompletionFunc // 完成回调
}

func (c *Completion) reset() {
	if c.done == nil {
		c.done = make(chan struct{}, 1)
	}
	if len(c.done) > 0 {
		<-c.done
	}
	c.Err = nil
	c.Callbacks = c.Callbacks[:0]
}

type MsgEnvelope struct {
	synclib.DataRef
	MessageData
	MetaData
	Completion
}

func (envelope *MsgEnvelope) Reset() {
	envelope.MetaData.reset()
	envelope.MessageData.reset()
	envelope.Completion.reset()
}

func (envelope *MsgEnvelope) SetHeaders(header Header) {
	for k, v := range header {
		envelope.Header.Set(k, v)
	}
}

func (envelope *MsgEnvelope) SetHeader(key string, value string) {
	if envelope.Header == nil {
		envelope.Header = Header{}
	}
	envelope.Header.Set(key, value)
}

func (envelope *MsgEnvelope) GetSender() *actor.PID {
	return envelope.Sender
}

func (envelope *MsgEnvelope) GetReceiver() *actor.PID {
	return envelope.Receiver
}

func (envelope *MsgEnvelope) GetReq() uint64 {
	return envelope.ReqID
}

func (envelope *MsgEnvelope) GetHeaders() map[string]string {
	if envelope.Header == nil {
		return EmptyMessageHeader
	}
	return envelope.Header.ToMap()
}

func (envelope *MsgEnvelope) GetHeader(key string) string {
	if envelope.Header == nil {
		return ""
	}
	return envelope.Header.Get(key)
}

func (envelope *MsgEnvelope) SetReply() {
	envelope.Reply = true
}

func (envelope *MsgEnvelope) IsReply() bool {
	return envelope.Reply
}

func (envelope *MsgEnvelope) SetError(err error) {
	envelope.Err = err
}

func (envelope *MsgEnvelope) SetErrStr(err string) {
	if err != "" {
		envelope.Err = errors.New(err)
	}
}

func (envelope *MsgEnvelope) GetError() error {
	return envelope.Err
}

func (envelope *MsgEnvelope) GetErrStr() string {
	if envelope.Err != nil {
		return envelope.Err.Error()
	}
	return ""
}

func (envelope *MsgEnvelope) GetMethod() string {
	return envelope.Method
}

func (envelope *MsgEnvelope) AddCompletion(cbs ...CompletionFunc) {
	envelope.Callbacks = append(envelope.Callbacks, cbs...)
}

func (envelope *MsgEnvelope) NeedCallback() bool {
	return len(envelope.Callbacks) > 0
}

func (envelope *MsgEnvelope) RunCompletions() {
	for _, cb := range envelope.Callbacks {
		cb(envelope.Response, envelope.Err)
	}
}

func (envelope *MsgEnvelope) Done() {
	if envelope.done != nil {
		envelope.done <- struct{}{}
	}
}

func (envelope *MsgEnvelope) wait() {
	<-envelope.done
}

func (envelope *MsgEnvelope) Wait() {
	envelope.wait()
}

func (envelope *MsgEnvelope) Result() (interface{}, error) {
	envelope.wait()
	return envelope.Response, envelope.Err
}

func (envelope *MsgEnvelope) GetReqID() uint64 {
	return envelope.ReqID
}

func (envelope *MsgEnvelope) GetTimeout() time.Duration {
	return envelope.Timeout
}

func (envelope *MsgEnvelope) ToProtoMsg() *actor.Message {
	msg := &actor.Message{
		TypeId:        0, // 默认使用protobuf(后面有其他需求再修改这里)
		TypeName:      "",
		Sender:        envelope.Sender,
		Receiver:      envelope.Receiver,
		Method:        envelope.Method,
		Request:       nil,
		Response:      nil,
		Err:           envelope.GetErrStr(),
		MessageHeader: envelope.Header,
		Reply:         envelope.Reply,
		ReqId:         envelope.ReqID,
		NeedResp:      envelope.NeedResp,
	}

	var byteData []byte
	var typeName string
	var err error

	if envelope.Request != nil {
		byteData, typeName, err = serializer.Serialize(envelope.Request, msg.TypeId)
		if err != nil {
			log.SysLogger.Errorf("serialize message[%+v] is error: %s", envelope, err)
			return nil
		}
		msg.Request = byteData
	} else if envelope.Response != nil {
		byteData, typeName, err = serializer.Serialize(envelope.Response, msg.TypeId)
		if err != nil {
			log.SysLogger.Errorf("serialize message[%+v] is error: %s", envelope, err)
			return nil
		}
		msg.Response = byteData
	}

	msg.TypeName = typeName

	return msg
}

//======================================================

var msgEnvelopePool = synclib.NewPoolEx(make(chan synclib.IPoolData, 10240), func() synclib.IPoolData {
	return &MsgEnvelope{}
})

var EmptyMessageHeader = make(Header)

func NewMsgEnvelopeWithHeader(header Header) *MsgEnvelope {
	envelope := NewMsgEnvelope()
	envelope.SetHeaders(header)
	return envelope
}

// TODO 记得测试资源释放

func NewMsgEnvelope() *MsgEnvelope {
	//log.SysLogger.Debugf(">>>>>>>>>>>>>>msg envelope count add: %d", atomic.AddInt64(&envelopeCount, 1))

	return msgEnvelopePool.Get().(*MsgEnvelope)
}

func ReleaseMsgEnvelope(envelope *MsgEnvelope) {
	if envelope != nil {
		msgEnvelopePool.Put(envelope)
	}
}
