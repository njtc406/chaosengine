// Package actor
// @Title  数据信封
// @Description  用于不同service之间的数据传递
// @Author  yr  2024/9/2 下午3:40
// @Update  yr  2024/9/2 下午3:40
package actor

import (
	"errors"
	"github.com/njtc406/chaosengine/engine1/synclib"
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
	Sender   *PID   // 发送者
	Receiver *PID   // 接收者
	Method   string // 调用方法
	ReqID    uint64 // 请求ID(防止重复)
	Reply    bool   // 是否是回复
	Header   Header // 消息头
}

func (m *MetaData) reset() {
	m.Sender = nil
	m.Receiver = nil
	m.Method = ""
	m.ReqID = 0
	m.Reply = false
	m.Header = nil
}

type MessageData struct {
	Request interface{} // 请求参数
	IsRpc   bool        // 是否是rpc调用
}

func (m *MessageData) reset() {
	m.Request = nil
	m.IsRpc = false
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

func (envelope *MsgEnvelope) GetSender() *PID {
	return envelope.Sender
}

func (envelope *MsgEnvelope) GetReceiver() *PID {
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
		cb(envelope.Request, envelope.Err)
	}
}

func (envelope *MsgEnvelope) Done() {
	if envelope.done != nil {
		envelope.done <- struct{}{}
	}
}

func (envelope *MsgEnvelope) Wait() {
	<-envelope.done
}

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
