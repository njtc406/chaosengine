// Package inf
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package inf

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/utils/errorlib"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"time"
)

type RequestHandler func(Returns []interface{}, Err string)
type CancelRpc func() // 异步调用时的取消函数,可用于取消调用
type SelectType int   // 选择类型 0随机 1哈希 2轮询 3加权轮询 4最少连接 5权重随机 (目前只实现0和1)

type IRpcInvoker interface {
	// Call 同步调用服务 (serviceMethod格式为 "ServiceName.MethodName")
	Call(serviceMethod string, in, out interface{}) error
	CallWithTimeout(serviceMethod string, timeout time.Duration, in, out interface{}) error
	// AsyncCall 异步调用服务
	AsyncCall(serviceMethod string, timeout time.Duration, callbacks []actor.CompletionFunc, in, out interface{}) (CancelRpc, error)
	// Send 无返回调用
	Send(serviceMethod string, in interface{}) error
	// Cast 广播调用
	Cast(serviceMethod string, in interface{}) error
}

type IRpcHandler interface {
	IRpcChannel

	GetName() string
	GetPID() *actor.PID            // 获取服务id
	GetRpcHandler() IRpcHandler    // 获取rpc服务
	Handle(msg *actor.MsgEnvelope) // 处理请求
	IsPrivate() bool               // 是否私有服务
	IsClosed() bool                // 服务是否已经关闭
}

// MultiRpcInvoker 多选器(主要是方便调用,所有的调用都可以使用rpc.Selector().Call()的方式来写,不用对返回多个节点的选择器进行处理)
type MultiRpcInvoker []IRpcInvoker

func NewEmptyMultiRpcInvoker() MultiRpcInvoker {
	return MultiRpcInvoker{}
}

func NewMultiRpcInvokerWithCap(list []IRpcInvoker) MultiRpcInvoker {
	return list
}

func (m MultiRpcInvoker) Call(serviceMethod string, in, out interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to call %s", serviceMethod)
		return nil
	}
	// call只允许调用一个节点
	return m[0].Call(serviceMethod, in, out)
}

func (m MultiRpcInvoker) CallWithTimeout(serviceMethod string, timeout time.Duration, in, out interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to call timeout %s", serviceMethod)
		return nil
	}
	// call只允许调用一个节点
	return m[0].CallWithTimeout(serviceMethod, timeout, in, out)
}

func (m MultiRpcInvoker) AsyncCall(serviceMethod string, timeout time.Duration, callbacks []actor.CompletionFunc, in, out interface{}) (CancelRpc, error) {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to async call %s", serviceMethod)
		return nil, nil
	}
	// call只允许调用一个节点
	return m[0].AsyncCall(serviceMethod, timeout, callbacks, in, out)
}

func (m MultiRpcInvoker) Send(serviceMethod string, in interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to send %s", serviceMethod)
		return nil
	}
	var errs []error
	for _, selector := range m {
		if err := selector.Send(serviceMethod, in); err != nil {
			errs = append(errs, err)
		}
	}

	return errorlib.CombineErr(errs...)
}

func (m MultiRpcInvoker) Cast(serviceMethod string, in interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to cast %s", serviceMethod)
		return nil
	}
	var errs []error
	for _, selector := range m {
		if err := selector.Cast(serviceMethod, in); err != nil {
			errs = append(errs, err)
		}
	}

	return errorlib.CombineErr(errs...)
}

type IRpcSelector interface {
	// SelectByUID 根据 服务UID 选择服务
	SelectByUID(uid string) IRpcInvoker

	// SelectByName 根据服务名称选择服务
	SelectByName(serviceName string, selectType SelectType) IRpcInvoker

	// SelectByNodeUID 根据节点 UID 和服务 UID 选择服务
	SelectByNodeUID(nodeUID, serviceUID string) IRpcInvoker

	// SelectByRule 根据自定义规则选择服务
	SelectByRule(rule func(service *actor.PID) bool) IRpcInvoker
}

type IRpcChannel interface {
	PushRequest(req *actor.MsgEnvelope) error
}
