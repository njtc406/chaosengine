// Package inf
// @Title  服务rpc接口
// @Description  服务rpc接口
// @Author  yr  2024/7/19 上午10:42
// @Update  yr  2024/7/19 上午10:42
package inf

import (
	"github.com/njtc406/chaosengine/engine1/actor"
	"github.com/njtc406/chaosutil/chaoserrors"
	"time"
)

type RequestHandler func(Returns []interface{}, Err string)
type CancelRpc func()

type IRpcHandler interface {
	IChannel
	IRpcClient

	GetName() string
	GetPID() *actor.PID                   // 获取服务id
	GetRpcHandler() IRpcHandler           // 获取rpc服务
	HandleRequest(msg *actor.MsgEnvelope) // 处理请求
	IsPrivate() bool                      // 是否私有服务
}

type IRpcClient interface {
	// Call 同步调用服务(serviceID如果没有设置默认就是nodeUID)
	Call(serviceID, serviceMethod string, args ...interface{}) ([]interface{}, chaoserrors.CError)
	CallNode(nodeUID, serviceID, serviceMethod string, in interface{}) ([]interface{}, chaoserrors.CError)
	CallWithTimeout(serviceID, serviceMethod string, timeout time.Duration, args ...interface{}) ([]interface{}, chaoserrors.CError)
	CallNodeWithTimeout(nodeUID, serviceID, serviceMethod string, timeout time.Duration, in interface{}) ([]interface{}, chaoserrors.CError)

	// AsyncCall 异步调用服务
	AsyncCall(serviceID, serviceMethod string, timeout time.Duration, callbacks []actor.CompletionFunc, args ...interface{}) (CancelRpc, chaoserrors.CError)
	AsyncCallNode(nodeUID, serviceID, serviceMethod string, timeout time.Duration, callbacks []actor.CompletionFunc, in interface{}) (CancelRpc, chaoserrors.CError)

	// Send 无返回调用
	Send(serviceID, serviceMethod string, args ...interface{}) chaoserrors.CError // 无返回调用
	SendNode(nodeUID, serviceID, serviceMethod string, in interface{}) chaoserrors.CError

	// CastNode 广播调用(nodeUID和serviceID为空则广播给所有注册的节点,否则给指定节点)
	CastNode(serviceMethod string, in interface{}) chaoserrors.CError
}
