// Package client
// @Title  本地服务的Client
// @Description  本地服务的Client
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package client

import (
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/errdef"
	"time"

	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/errdef/errcode"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/asynclib"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosutil/chaoserrors"
)

// LClient 本地服务的Client
type LClient struct {
	Client
}

func NewLClient(pid *actor.PID, monitor inf.IMonitor, rpcHandler inf.IRpcHandler) inf.IClient {
	lClient := &LClient{}
	lClient.pid = pid
	lClient.IMonitor = monitor
	lClient.IRpcHandler = rpcHandler
	return lClient
}

func (lc *LClient) Close() {}

func (lc *LClient) SendMessage(envelope *actor.MsgEnvelope) error {
	if envelope.IsReply() && !envelope.IsRpc && !envelope.NeedCallback() {
		defer actor.ReleaseMsgEnvelope(envelope)
		// 本地Call的回复, 通过reqId找到future,直接返回结果
		future := lc.Get(envelope.ReqID)
		if future == nil {
			// 已经超时了,丢弃
			return errdef.RPCCallTimeout
		}

		future.SetResult(envelope.Request.(interface{}), envelope.GetError())

		return nil
	}

	// 根据自己的pid找到对应的rpcHandler
	rpcHandler := lc.GetRpcHandler()
	if rpcHandler == nil || rpcHandler.IsClosed() {
		// 按道理不会是nil,因为如果服务关闭,会释放掉client,如果走到这里,那么只能是刚好收到一条消息,此时服务又关闭了
		// 已经被释放了
		actor.ReleaseMsgEnvelope(envelope)
		return errdef.ServiceNotFound
	}

	// 直接push消息(envelope会在对应service的rpcHandler调用完成后回收)
	return rpcHandler.PushRequest(envelope)
}

func (lc *LClient) SendMessageWithFuture(envelope *actor.MsgEnvelope, timeout time.Duration) *actor.Future {
	if timeout == 0 {
		timeout = def.DefaultRpcTimeout
	}
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)

	// 根据自己的pid找到对应的rpcHandler
	rpcHandler := lc.GetRpcHandler()
	if rpcHandler == nil || rpcHandler.IsClosed() {
		actor.ReleaseMsgEnvelope(envelope)
		future.SetResult(nil, chaoserrors.NewErrCode(errcode.ServiceNotExist, "service rpc handler not exist"))
		return future
	}

	if timeout > 0 {
		lc.Add(future)
	}

	// 这里的envelope如果发生成功,将在rpcRequest处理完成后回收
	if err := rpcHandler.PushRequest(envelope); err != nil {
		// 发送错误,直接回收信封
		actor.ReleaseMsgEnvelope(envelope)
		// 设置错误结果
		future.SetResult(nil, chaoserrors.NewErrCode(errcode.RpcErr, err))
		return future
	}

	return future
}

func (lc *LClient) AsyncSendMessage(envelope *actor.MsgEnvelope, timeout time.Duration, completions []actor.CompletionFunc) (inf.CancelRpc, error) {
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)
	future.SetCompletions(completions...)

	// 根据自己的pid找到对应的rpcHandler
	rpcHandler := lc.GetRpcHandler()
	if rpcHandler == nil || rpcHandler.IsClosed() {
		actor.ReleaseMsgEnvelope(envelope)
		actor.ReleaseFuture(future)
		return inf.EmptyCancelRpc, errdef.ServiceNotFound
	}

	cancelRpc := inf.EmptyCancelRpc
	if timeout > 0 {
		lc.Add(future)
		cancelRpc = NewRpcCancel(lc, envelope.ReqID)
	}

	// 这里的envelope如果发送成功,将在rpcRequest处理完成后回收
	if err := rpcHandler.PushRequest(envelope); err != nil {
		// 发送错误,直接回收信封
		actor.ReleaseMsgEnvelope(envelope)
		lc.Remove(future.GetReqID())
		actor.ReleaseFuture(future)
		return inf.EmptyCancelRpc, err
	}

	asynclib.Go(func() {
		// 异步等待结果后释放
		future.Wait()
		log.SysLogger.Debug("asynclib send message future call back, release future")
		lc.Remove(future.GetReqID())
		actor.ReleaseFuture(future)
	})

	return cancelRpc, nil
}
