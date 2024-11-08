// Package actor
// @Title  数据信封
// @Description  用于不同service之间的数据传递
// @Author  yr  2024/9/2 下午3:40
// @Update  yr  2024/9/2 下午3:40
package actor

import (
	"github.com/njtc406/chaosengine/engine1/synclib"
	"github.com/njtc406/chaosutil/chaoserrors"
)

type CompletionFunc func(res []interface{}, err chaoserrors.CError)

type messageHeader map[string]string

func (header messageHeader) Get(key string) string {
	return header[key]
}

func (header messageHeader) Set(key string, value string) {
	header[key] = value
}

func (header messageHeader) Keys() []string {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	return keys
}

func (header messageHeader) Length() int {
	return len(header)
}

func (header messageHeader) ToMap() map[string]string {
	mp := make(map[string]string)
	for k, v := range header {
		mp[k] = v
	}
	return mp
}

type ReadonlyMessageHeader interface {
	Get(key string) string
	Keys() []string
	Length() int
	ToMap() map[string]string
}

type MsgEnvelope struct {
	synclib.DataRef

	Sender   *PID          // 发送者
	Receiver *PID          // 接收者
	Method   string        // 调用方法
	Request  interface{}   // 方法参数
	Header   messageHeader // 消息头(额外信息)
	Reply    bool          // 是否是回复
	ReqID    uint64        // 请求ID
	IsRpc    bool          // 是否是RPC
	Err      string        // 错误

	done        chan struct{}      // 完成信号
	err         chaoserrors.CError // 错误
	completions []CompletionFunc   // 完成回调
}

func (envelope *MsgEnvelope) Reset() {
	envelope.Sender = nil
	envelope.Receiver = nil
	envelope.Method = ""
	envelope.Request = nil
	envelope.Header = nil
	envelope.Reply = false
	envelope.ReqID = 0
	envelope.IsRpc = false
	envelope.Err = ""

	if envelope.done == nil {
		envelope.done = make(chan struct{}, 1)
	}
	if len(envelope.done) > 0 {
		<-envelope.done
	}
	envelope.err = nil
	envelope.completions = nil
}

func (envelope *MsgEnvelope) GetArgs() interface{} {
	return envelope.Request
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

func (envelope *MsgEnvelope) GetMethod() string {
	return envelope.Method
}

func (envelope *MsgEnvelope) SetHeader(key string, value string) {
	if envelope.Header == nil {
		envelope.Header = make(map[string]string)
	}
	envelope.Header.Set(key, value)
}

func (envelope *MsgEnvelope) SetSender(sender *PID) {
	envelope.Sender = sender
}

func (envelope *MsgEnvelope) SetReceiver(receiver *PID) {
	envelope.Receiver = receiver
}

func (envelope *MsgEnvelope) SetArgs(args interface{}) {
	envelope.Request = args
}

func (envelope *MsgEnvelope) SetMethod(method string) {
	envelope.Method = method
}

func (envelope *MsgEnvelope) SetReq(reqID uint64) {
	envelope.ReqID = reqID
}

func (envelope *MsgEnvelope) SetReply() {
	envelope.Reply = true
}

func (envelope *MsgEnvelope) SetError(err chaoserrors.CError) {
	envelope.err = err
}

func (envelope *MsgEnvelope) AddCompletion(cbs ...CompletionFunc) {
	envelope.completions = append(envelope.completions, cbs...)
}

func (envelope *MsgEnvelope) NeedCallback() bool {
	return len(envelope.completions) > 0
}

func (envelope *MsgEnvelope) RunCompletions() {
	var resp []interface{}
	switch envelope.Request.(type) {
	case []interface{}:
		resp = envelope.Request.([]interface{})
	default:
		resp = []interface{}{envelope.Request}
	}
	for _, cb := range envelope.completions {
		cb(resp, envelope.err)
	}
}

func (envelope *MsgEnvelope) GetError() chaoserrors.CError {
	return envelope.err
}

func (envelope *MsgEnvelope) IsReply() bool {
	return envelope.Reply
}

func (envelope *MsgEnvelope) SetHeaders(header map[string]string) {
	if envelope == nil || header == nil {
		return
	}
	if envelope.Header == nil {
		envelope.Header = make(messageHeader, len(header))
	}
	for k, v := range header {
		envelope.SetHeader(k, v)
	}
}

func (envelope *MsgEnvelope) CreateDone() {
	envelope.done = make(chan struct{}, 1)
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

var EmptyMessageHeader = make(messageHeader)

var envelopeCount int64

func NewMsgEnvelope() *MsgEnvelope {
	//log.SysLogger.Debugf(">>>>>>>>>>>>>>msg envelope count add: %d", atomic.AddInt64(&envelopeCount, 1))

	return msgEnvelopePool.Get().(*MsgEnvelope)
}

func ReleaseMsgEnvelope(envelope *MsgEnvelope) {
	if envelope != nil {
		//log.SysLogger.Debugf(">>>>>>>>>>>>>>msg envelope count sub: %d", atomic.AddInt64(&envelopeCount, -1))
		msgEnvelopePool.Put(envelope)
	}
}
