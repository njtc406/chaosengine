// Package messagebus
// @Title  消息总线
// @Description  所有的消息都通过该模块进行发送和接收
// @Author  yr  2024/11/12
// @Update  yr  2024/11/12
package messagebus

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine1/synclib"
	"reflect"
	"sync/atomic"
	"time"
)

var requestID uint64

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

func genReqId() uint64 {
	return atomic.AddUint64(&requestID, 1)
}

// Call 同步调用服务
func (mb *MessageBus) Call(method string, in, out interface{}) error {
	defer ReleaseMessageBus(mb)
	if mb.sender == nil || mb.receiverClient == nil {
		return fmt.Errorf("sender or receiver is nil")
	}
	// 封装消息
	envelope := actor.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = genReqId()
	envelope.Request = in

	future := mb.receiverClient.SendMessageWithFuture(envelope, 0)
	defer actor.ReleaseFuture(future)

	resp, err := future.Result()
	if err != nil {
		return err
	}

	if reflect.TypeOf(out) != reflect.TypeOf(resp) {
		return fmt.Errorf("type not match, you give %v but got resp %v", reflect.TypeOf(out), reflect.TypeOf(resp))
	}

	reflect.ValueOf(out).Elem().Set(reflect.ValueOf(resp))

	return nil
}
func (mb *MessageBus) CallWithTimeout(method string, timeout time.Duration, in, out interface{}) error {
	defer ReleaseMessageBus(mb)
	if mb.sender == nil || mb.receiverClient == nil {
		return fmt.Errorf("sender or receiver is nil")
	}
	// 封装消息
	envelope := actor.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = genReqId()
	envelope.Request = in

	future := mb.receiverClient.SendMessageWithFuture(envelope, timeout)
	defer actor.ReleaseFuture(future)

	resp, err := future.Result()
	if err != nil {
		return err
	}

	if reflect.TypeOf(out) != reflect.TypeOf(resp) {
		return fmt.Errorf("type not match, you give %v but got resp %v", reflect.TypeOf(out), reflect.TypeOf(resp))
	}

	reflect.ValueOf(out).Elem().Set(reflect.ValueOf(resp))

	return nil
}

// AsyncCall 异步调用服务
func (mb *MessageBus) AsyncCall(method string, timeout time.Duration, callbacks []actor.CompletionFunc, in interface{}) (inf.CancelRpc, error) {
	defer ReleaseMessageBus(mb)
	if mb.sender == nil || mb.receiverClient == nil {
		return inf.EmptyCancelRpc, fmt.Errorf("sender or receiver is nil")
	}
	// 封装消息
	envelope := actor.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = genReqId()
	envelope.Request = in

	return mb.receiverClient.AsyncSendMessage(envelope, timeout, callbacks)
}

// Send 无返回调用
func (mb *MessageBus) Send(method string, in interface{}) error {
	defer ReleaseMessageBus(mb)
	if mb.receiverClient == nil {
		return fmt.Errorf("receiver is nil")
	}
	envelope := actor.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = genReqId()
	envelope.Request = in

	return mb.receiverClient.SendMessage(envelope)
}

// Cast 广播调用(实际和send是一样的)
func (mb *MessageBus) Cast(method string, in interface{}) error {
	defer ReleaseMessageBus(mb)
	if mb.receiverClient == nil {
		return fmt.Errorf("receiver is nil")
	}
	envelope := actor.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = mb.sender
	envelope.Receiver = mb.receiverClient.GetPID()
	envelope.ReqID = genReqId()
	envelope.Request = in

	return mb.receiverClient.SendMessage(envelope)
}

// TODO 这个还需要修改
