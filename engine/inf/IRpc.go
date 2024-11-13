// Package inf
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package inf

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/msgenvelope"
)

type RequestHandler func(Returns []interface{}, Err string)
type CancelRpc func() // 异步调用时的取消函数,可用于取消调用

type IRpcHandler interface {
	IRpcChannel
	IRpcSelector

	GetName() string
	GetPID() *actor.PID                          // 获取服务id
	GetRpcHandler() IRpcHandler                  // 获取rpc服务
	HandleRequest(msg *msgenvelope.MsgEnvelope)  // 处理请求
	HandleResponse(msg *msgenvelope.MsgEnvelope) // 处理回复
	IsPrivate() bool                             // 是否私有服务
	IsClosed() bool                              // 服务是否已经关闭
}

type IRpcSelector interface {
	Select(serverId int32, serviceId, serviceName string) IBus

	SelectSameServer(serviceUid, serviceName string) IBus

	SelectByPid(receiver *actor.PID) IBus

	// SelectByRule 根据自定义规则选择服务
	SelectByRule(rule func(pid *actor.PID) bool) IBus
}

type IRpcChannel interface {
	PushRequest(req *msgenvelope.MsgEnvelope) error
}
