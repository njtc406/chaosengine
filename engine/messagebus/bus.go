// Package messagebus
// @Title  消息总线
// @Description  所有的消息都通过该模块进行发送和接收
// @Author  yr  2024/11/12
// @Update  yr  2024/11/12
package messagebus

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/msgenvelope"
	"github.com/njtc406/chaosengine/engine/utils/errorlib"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine1/synclib"
	"reflect"
	"time"
)

type MessageBus struct {
	synclib.DataRef
	sender         *actor.PID
	receiverClient inf.IClient
}

func (mb *MessageBus) Reset() {
	mb.sender = nil
	mb.receiverClient = nil
}

var pool = synclib.NewPoolEx(make(chan synclib.IPoolData, 2048), func() synclib.IPoolData {
	return &MessageBus{}
})

func NewMessageBus(sender *actor.PID, receiver inf.IClient) *MessageBus {
	mb := pool.Get().(*MessageBus)
	mb.sender = sender
	mb.receiverClient = receiver
	return mb
}

func ReleaseMessageBus(mb *MessageBus) {
	pool.Put(mb)
}

// Call 同步调用服务
func (mb *MessageBus) Call(method string, in, out interface{}) error {
	if mb.sender == nil || mb.receiverClient == nil {
		return fmt.Errorf("sender or receiver is nil")
	}

	// 封装消息
	envelope := msgenvelope.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = mb.receiverClient.GenSeq()
	envelope.Request = in
	envelope.NeedResp = true
	envelope.Timeout = def.DefaultRpcTimeout // TODO 暂时不支持没有超时时间的调用,风险太高,容易造成服务阻塞

	defer msgenvelope.ReleaseMsgEnvelope(envelope) // 释放资源

	// 加入等待队列
	mb.receiverClient.Add(envelope)
	// 发送消息
	mb.receiverClient.SendMessageWithFuture(envelope)

	resp, err := envelope.Result()
	if err != nil {
		return err
	}

	if reflect.TypeOf(out) != reflect.TypeOf(resp) {
		return fmt.Errorf("call: type not match, expected %v but got %v", reflect.TypeOf(out), reflect.TypeOf(resp))
	}

	reflect.ValueOf(out).Elem().Set(reflect.ValueOf(resp))

	return nil
}
func (mb *MessageBus) CallWithTimeout(method string, timeout time.Duration, in, out interface{}) error {
	if mb.sender == nil || mb.receiverClient == nil {
		return fmt.Errorf("sender or receiver is nil")
	}

	if timeout == 0 {
		timeout = def.DefaultRpcTimeout
	}
	// 封装消息
	envelope := msgenvelope.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = mb.receiverClient.GenSeq()
	envelope.Request = in
	envelope.NeedResp = true
	envelope.Timeout = timeout

	mb.receiverClient.SendMessageWithFuture(envelope)

	resp, err := envelope.Result()
	if err != nil {
		return err
	}

	if reflect.TypeOf(out) != reflect.TypeOf(resp) {
		return fmt.Errorf("callWithTimeout: type not match, expected %v but got %v", reflect.TypeOf(out), reflect.TypeOf(resp))
	}

	reflect.ValueOf(out).Elem().Set(reflect.ValueOf(resp))

	return nil
}

// AsyncCall 异步调用服务
func (mb *MessageBus) AsyncCall(method string, timeout time.Duration, callbacks []msgenvelope.CompletionFunc, in interface{}) (inf.CancelRpc, error) {
	if mb.sender == nil || mb.receiverClient == nil {
		return inf.EmptyCancelRpc, fmt.Errorf("sender or receiver is nil")
	}

	if timeout == 0 {
		timeout = def.DefaultRpcTimeout
	}

	// 封装消息
	envelope := msgenvelope.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = mb.receiverClient.GenSeq()
	envelope.Request = in
	envelope.NeedResp = true

	return mb.receiverClient.AsyncSendMessage(envelope, timeout, callbacks)
}

// Send 无返回调用
func (mb *MessageBus) Send(method string, in interface{}) error {
	if mb.receiverClient == nil {
		return fmt.Errorf("receiver is nil")
	}
	envelope := msgenvelope.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = mb.receiverClient.GenSeq()
	envelope.Request = in
	envelope.NeedResp = false

	return mb.receiverClient.SendMessage(envelope)
}

// Cast 广播调用(实际和send是一样的)
func (mb *MessageBus) Cast(method string, in interface{}) error {
	if mb.receiverClient == nil {
		return fmt.Errorf("receiver is nil")
	}
	envelope := msgenvelope.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = mb.receiverClient.GenSeq()
	envelope.Request = in
	envelope.NeedResp = false

	return mb.receiverClient.SendMessage(envelope)
}

func (mb *MessageBus) Response(envelope *msgenvelope.MsgEnvelope) error {
	return mb.receiverClient.SendMessage(envelope)
}

// TODO 这个还需要修改

// MultiBus 多节点调用
type MultiBus []inf.IBus

func (m MultiBus) Call(serviceMethod string, in, out interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to call %s", serviceMethod)
		return nil
	}

	if len(m) > 1 {
		// 释放所有节点
		for _, bus := range m {
			ReleaseMessageBus(bus.(*MessageBus))
		}
		return fmt.Errorf("only one node can be called at a time, now got %v", len(m))
	}

	// call只允许调用一个节点
	return m[0].Call(serviceMethod, in, out)
}

func (m MultiBus) CallWithTimeout(serviceMethod string, timeout time.Duration, in, out interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to call timeout %s", serviceMethod)
		return nil
	}

	if len(m) > 1 {
		// 释放所有节点
		for _, bus := range m {
			ReleaseMessageBus(bus.(*MessageBus))
		}
		return fmt.Errorf("only one node can be called at a time, now got %v", len(m))
	}

	// call只允许调用一个节点
	return m[0].CallWithTimeout(serviceMethod, timeout, in, out)
}

func (m MultiBus) AsyncCall(serviceMethod string, timeout time.Duration, callbacks []msgenvelope.CompletionFunc, in interface{}) (inf.CancelRpc, error) {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to async call %s", serviceMethod)
		return inf.EmptyCancelRpc, nil
	}
	if len(m) > 1 {
		// 释放所有节点
		for _, bus := range m {
			ReleaseMessageBus(bus.(*MessageBus))
		}
		return inf.EmptyCancelRpc, fmt.Errorf("only one node can be called at a time, now got %v", len(m))
	}
	// call只允许调用一个节点
	return m[0].AsyncCall(serviceMethod, timeout, callbacks, in)
}

func (m MultiBus) Send(serviceMethod string, in interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to send %s", serviceMethod)
		return nil
	}
	var errs []error
	for _, bus := range m {
		if err := bus.Send(serviceMethod, in); err != nil {
			errs = append(errs, err)
		}
	}

	return errorlib.CombineErr(errs...)
}

func (m MultiBus) Cast(serviceMethod string, in interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to cast %s", serviceMethod)
		return nil
	}
	var errs []error
	for _, bus := range m {
		if err := bus.Cast(serviceMethod, in); err != nil {
			errs = append(errs, err)
		}
	}

	return errorlib.CombineErr(errs...)
}

func (m MultiBus) Response(envelope *msgenvelope.MsgEnvelope) error {
	if len(m) != 1 {
		return fmt.Errorf("response bus must be one, but got %v", len(m))
	}

	return m[0].Response(envelope)
}
