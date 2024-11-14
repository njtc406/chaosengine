// Package client
// @Title  本地服务的Client
// @Description  本地服务的Client,调用时直接使用rpcHandler发往对应的service
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package client

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/monitor"
)

// LocalSender 本地服务的Client
type LocalSender struct {
	SenderBase
}

func NewLClient(pid *actor.PID, handler inf.IRpcHandler) inf.IRpcSender {
	lClient := &LocalSender{}
	lClient.pid = pid
	lClient.IRpcHandler = handler
	return lClient
}

func (lc *LocalSender) Close() {}

func (lc *LocalSender) SendRequest(envelope inf.IEnvelope) error {
	if lc.IsClosed() {
		return errdef.ServiceNotFound
	}

	return lc.PushRequest(envelope)
}

func (lc *LocalSender) SendResponse(envelope inf.IEnvelope) error {
	monitor.GetRpcMonitor().Remove(envelope.GetReqId()) // 回复时先移除监控,防止超时
	if lc.IsClosed() {
		return errdef.ServiceNotFound
	}

	if envelope.NeedCallback() {
		// 本地调用的回复消息,直接发送到对应service的邮箱处理
		return lc.PushRequest(envelope)
	} else {
		// 同步调用,直接设置调用结束
		envelope.Done()
	}
	return nil
}

func (lc *LocalSender) SendRequestAndRelease(envelope inf.IEnvelope) error {
	// 本地调用envelope在接收者处理后释放
	return lc.SendRequest(envelope)
}
