// Package inf
// @Title  title
// @Description  desc
// @Author  yr  2024/11/12
// @Update  yr  2024/11/12
package inf

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/utils/errorlib"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"time"
)

type IBus interface {
	// Call 同步调用服务 (serviceMethod格式为 "ServiceName.MethodName")
	Call(method string, in, out interface{}) error
	CallWithTimeout(method string, timeout time.Duration, in, out interface{}) error
	// AsyncCall 异步调用服务
	AsyncCall(method string, timeout time.Duration, callbacks []actor.CompletionFunc, in interface{}) (CancelRpc, error)
	// Send 无返回调用
	Send(method string, in interface{}) error
	// Cast 广播调用
	Cast(method string, in interface{}) error
}

// MultiBus 多节点调用
type MultiBus []IBus

func (m MultiBus) Call(serviceMethod string, in, out interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to call %s", serviceMethod)
		return nil
	}
	// call只允许调用一个节点
	return m[0].Call(serviceMethod, in, out)
}

func (m MultiBus) CallWithTimeout(serviceMethod string, timeout time.Duration, in, out interface{}) error {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to call timeout %s", serviceMethod)
		return nil
	}
	// call只允许调用一个节点
	return m[0].CallWithTimeout(serviceMethod, timeout, in, out)
}

func (m MultiBus) AsyncCall(serviceMethod string, timeout time.Duration, callbacks []actor.CompletionFunc, in interface{}) (CancelRpc, error) {
	if len(m) == 0 {
		log.SysLogger.Errorf("===========select empty service to async call %s", serviceMethod)
		return nil, nil
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
