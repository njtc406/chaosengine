// Package messagebus
// @Title  消息总线
// @Description  所有的消息都通过该模块进行发送和接收
// @Author  yr  2024/11/12
// @Update  yr  2024/11/12
package messagebus

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/dto"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/monitor"
	"github.com/njtc406/chaosengine/engine/msgenvelope"
	"github.com/njtc406/chaosengine/engine/utils/errorlib"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/pool"
	"reflect"
	"time"
)

type MessageBus struct {
	dto.DataRef
	senderClient   inf.IRpcSender
	receiverClient inf.IRpcSender
	err            error
}

func (mb *MessageBus) Reset() {
	mb.senderClient = nil
	mb.receiverClient = nil
}

var busPool = pool.NewPoolEx(make(chan pool.IPoolData, 2048), func() pool.IPoolData {
	return &MessageBus{}
})

func NewMessageBus(sender inf.IRpcSender, receiver inf.IRpcSender, err error) *MessageBus {
	mb := busPool.Get().(*MessageBus)
	mb.senderClient = sender
	mb.receiverClient = receiver
	mb.err = err
	return mb
}

func ReleaseMessageBus(mb *MessageBus) {
	busPool.Put(mb)
}

func (mb *MessageBus) call(method string, timeout time.Duration, in, out interface{}) error {
	defer ReleaseMessageBus(mb)
	if mb.senderClient == nil || mb.receiverClient == nil {
		return fmt.Errorf("senderClient or receiver is nil")
	}
	if mb.err != nil {
		return mb.err
	}
	if out != nil {
		switch out.(type) {
		case []interface{}:
			// 远程调用都是固定的proto消息,不会出现这个类型的参数
			// 本地调用,接收多参数返回值,那么所有的接收参数都必须是指针或者引用类型
			for i, v := range out.([]interface{}) {
				kd := reflect.TypeOf(v).Kind()
				if kd != reflect.Ptr && kd != reflect.Interface &&
					kd != reflect.Func && kd != reflect.Map &&
					kd != reflect.Slice && kd != reflect.Chan {
					return fmt.Errorf("multi out call: all out params must be pointer, but the %v one got %v", i, kd)
				}
			}
		default:
			kd := reflect.TypeOf(out).Kind()
			if kd != reflect.Ptr && kd != reflect.Interface &&
				kd != reflect.Func && kd != reflect.Map &&
				kd != reflect.Slice && kd != reflect.Chan {
				return fmt.Errorf("single out call: out param must be pointer, but got:%v", kd)
			}
		}
	}

	mt := monitor.GetRpcMonitor()

	// 创建请求
	envelope := msgenvelope.NewMsgEnvelope()
	envelope.SetMethod(method)
	envelope.SetSender(mb.senderClient.GetPID())
	envelope.SetReceiver(mb.receiverClient.GetPID())
	envelope.SetSenderClient(mb.senderClient)
	envelope.SetRequest(in)
	envelope.SetResponse(nil) // 容错
	envelope.SetReqId(mt.GenSeq())
	envelope.SetNeedResponse(true)
	envelope.SetTimeout(timeout)

	// 加入等待队列
	mt.Add(envelope)

	// 发送消息
	if err := mb.receiverClient.SendRequest(envelope); err != nil {
		// 发送失败,释放资源
		mt.Remove(envelope.GetReqId())
		msgenvelope.ReleaseMsgEnvelope(envelope)
		log.SysLogger.Errorf("service[%s] send message[%s] request to client failed, error: %v", mb.senderClient.GetName(), envelope.GetMethod(), err)
		return errdef.RPCCallFailed
	}

	// 等待回复
	envelope.Wait()

	mt.Remove(envelope.GetReqId()) // 容错,不管有没有释放,都释放一次(实际上在所有设置done之前都会释放)
	if err := envelope.GetError(); err != nil {
		return err
	}

	resp := envelope.GetResponse()

	// 获取到返回后直接释放
	msgenvelope.ReleaseMsgEnvelope(envelope)

	// 如果out为nil表示丢弃返回值
	if out != nil {
		switch out.(type) {
		case []interface{}:
			outList := out.([]interface{})
			respList, ok := resp.([]interface{})
			if !ok {
				return fmt.Errorf("call: type not match, expected %v but got %v", reflect.TypeOf(out), reflect.TypeOf(respList))
			}
			for idx, v := range outList {
				respType := reflect.TypeOf(respList[idx])
				respKd := respType.Kind()
				if respKd == reflect.Ptr {
					respType = respType.Elem()
				}
				outType := reflect.TypeOf(v)
				outKd := outType.Kind()
				if outKd == reflect.Ptr {
					outType = outType.Elem()
				}
				if outType != respType {
					return fmt.Errorf("call: type not match, expected %v but got %v", outType, respType)
				}
				respVal := reflect.ValueOf(respList[idx])
				if respVal.Kind() == reflect.Ptr {
					respVal = respVal.Elem()
				}

				reflect.ValueOf(v).Elem().Set(respVal)
			}
		default:
			respType := reflect.TypeOf(resp)
			respKd := respType.Kind()
			if respKd == reflect.Ptr {
				respType = respType.Elem()
			}
			outType := reflect.TypeOf(out)
			outKd := outType.Kind()
			if outKd == reflect.Ptr {
				outType = outType.Elem()
			}
			if outType != respType {
				return fmt.Errorf("call: type not match, expected %v but got %v", outType, respType)
			}
			respVal := reflect.ValueOf(resp)
			if respVal.Kind() == reflect.Ptr {
				respVal = respVal.Elem()
			}

			reflect.ValueOf(out).Elem().Set(respVal)
		}
	}

	return nil
}

// Call 同步调用服务
func (mb *MessageBus) Call(method string, in, out interface{}) error {
	return mb.call(method, def.DefaultRpcTimeout, in, out)
}
func (mb *MessageBus) CallWithTimeout(method string, timeout time.Duration, in, out interface{}) error {
	return mb.call(method, timeout, in, out)
}

// AsyncCall 异步调用服务
func (mb *MessageBus) AsyncCall(method string, timeout time.Duration, callbacks []dto.CompletionFunc, in interface{}) (dto.CancelRpc, error) {
	defer ReleaseMessageBus(mb)
	if mb.senderClient == nil || mb.receiverClient == nil {
		return dto.EmptyCancelRpc, fmt.Errorf("senderClient or receiver is nil")
	}

	if mb.err != nil {
		return dto.EmptyCancelRpc, mb.err
	}

	mt := monitor.GetRpcMonitor()

	// 创建请求
	envelope := msgenvelope.NewMsgEnvelope()
	envelope.SetMethod(method)
	envelope.SetSender(mb.senderClient.GetPID())
	envelope.SetReceiver(mb.receiverClient.GetPID())
	envelope.SetSenderClient(mb.senderClient)
	envelope.SetRequest(in)
	envelope.SetResponse(nil) // 容错
	envelope.SetReqId(mt.GenSeq())
	envelope.SetNeedResponse(true)
	envelope.SetTimeout(timeout)
	envelope.SetCallback(callbacks)

	// 加入等待队列
	mt.Add(envelope)

	// 发送消息,最终callback调用将在response中被执行,所以envelope会在callback执行完后自动回收
	if err := mb.receiverClient.SendRequest(envelope); err != nil {
		// 发送失败,释放资源
		mt.Remove(envelope.GetReqId())
		msgenvelope.ReleaseMsgEnvelope(envelope)
		log.SysLogger.Errorf("service[%s] send message[%s] request to client failed, error: %v", mb.senderClient.GetName(), envelope.GetMethod(), err)
		return dto.EmptyCancelRpc, errdef.RPCCallFailed
	}

	return mt.NewCancel(envelope.GetReqId()), nil
}

// Send 无返回调用
func (mb *MessageBus) Send(method string, in interface{}) error {
	defer ReleaseMessageBus(mb)
	if mb.receiverClient == nil {
		return fmt.Errorf("senderClient or receiver is nil")
	}
	if mb.err != nil {
		return mb.err
	}
	mt := monitor.GetRpcMonitor()

	// 创建请求
	envelope := msgenvelope.NewMsgEnvelope()
	envelope.SetMethod(method)
	envelope.SetReceiver(mb.receiverClient.GetPID())
	envelope.SetSenderClient(mb.senderClient)
	envelope.SetRequest(in)
	envelope.SetResponse(nil) // 容错
	envelope.SetReqId(mt.GenSeq())
	envelope.SetNeedResponse(false) // 不需要回复

	// 如果是远程调用, 则由远程调用释放资源,如果是本地调用,则由接收者自行回收
	return mb.receiverClient.SendRequestAndRelease(envelope)
}

// Cast 广播调用(实际和send是一样的)
func (mb *MessageBus) Cast(method string, in interface{}) error {
	if mb.receiverClient == nil {
		return fmt.Errorf("senderClient or receiver is nil")
	}
	if mb.err != nil {
		return mb.err
	}
	mt := monitor.GetRpcMonitor()

	// 创建请求
	envelope := msgenvelope.NewMsgEnvelope()
	envelope.SetMethod(method)
	envelope.SetReceiver(mb.receiverClient.GetPID())
	envelope.SetSenderClient(mb.senderClient)
	envelope.SetRequest(in)
	envelope.SetResponse(nil) // 容错
	envelope.SetReqId(mt.GenSeq())
	envelope.SetNeedResponse(false) // 不需要回复

	// 如果是远程调用, 则由远程调用释放资源,如果是本地调用,则由接收者自行回收
	return mb.receiverClient.SendRequestAndRelease(envelope)
}

// TODO 这个还需要修改

// MultiBus 多节点调用
type MultiBus []inf.IBus

func (m MultiBus) Call(serviceMethod string, in, out interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to call %s", serviceMethod)
		return errdef.ServiceIsUnavailable
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
		return errdef.ServiceIsUnavailable
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

func (m MultiBus) AsyncCall(serviceMethod string, timeout time.Duration, callbacks []dto.CompletionFunc, in interface{}) (dto.CancelRpc, error) {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to async call %s", serviceMethod)
		return nil, errdef.ServiceIsUnavailable
	}
	if len(m) > 1 {
		// 释放所有节点
		for _, bus := range m {
			ReleaseMessageBus(bus.(*MessageBus))
		}
		return dto.EmptyCancelRpc, fmt.Errorf("only one node can be called at a time, now got %v", len(m))
	}
	// call只允许调用一个节点
	return m[0].AsyncCall(serviceMethod, timeout, callbacks, in)
}

func (m MultiBus) Send(serviceMethod string, in interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to send %s", serviceMethod)
		return errdef.ServiceIsUnavailable
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
		return errdef.ServiceIsUnavailable
	}
	var errs []error
	for _, bus := range m {
		if err := bus.Cast(serviceMethod, in); err != nil {
			errs = append(errs, err)
		}
	}

	return errorlib.CombineErr(errs...)
}
