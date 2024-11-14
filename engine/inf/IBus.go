// Package inf
// @Title  title
// @Description  desc
// @Author  yr  2024/11/12
// @Update  yr  2024/11/12
package inf

import (
	"github.com/njtc406/chaosengine/engine/def"
	"time"
)

type IBus interface {
	// Call 同步调用服务 (serviceMethod格式为 "ServiceName.MethodName")
	Call(method string, in, out interface{}) error
	CallWithTimeout(method string, timeout time.Duration, in, out interface{}) error
	// AsyncCall 异步调用服务
	AsyncCall(method string, timeout time.Duration, callbacks []def.CompletionFunc, in interface{}) (def.CancelRpc, error)
	// Send 无返回调用
	Send(method string, in interface{}) error
	// Cast 广播调用
	Cast(method string, in interface{}) error
}
