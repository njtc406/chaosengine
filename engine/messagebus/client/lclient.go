// Package client
// @Title  本地服务的Client
// @Description  本地服务的Client,调用时直接使用rpcHandler发往对应的service
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package client

import (
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/msgenvelope"
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

func NewLClient(pid *actor.PID, handler inf.IRpcHandler) inf.IClient {
	lClient := &LClient{}
	lClient.pid = pid
	lClient.IRpcHandler = handler
	return lClient
}

func (lc *LClient) Close() {}

func (lc *LClient) SendMessage(envelope *msgenvelope.MsgEnvelope) error {
	if envelope.IsReply() && !envelope.NeedCallback() {
		defer msgenvelope.ReleaseMsgEnvelope(envelope)
		// 本地Call的回复, 通过reqId找到future,直接返回结果
		future := lc.Remove(envelope.ReqID)
		if future == nil {
			// 已经超时了,丢弃
			return errdef.RPCCallTimeout
		}

		future.SetResult(envelope.Response, envelope.GetError())

		return nil
	}

	if lc.IsClosed() {
		msgenvelope.ReleaseMsgEnvelope(envelope)
		return errdef.ServiceNotFound
	}

	return lc.PushRequest(envelope)
}

func (lc *LClient) SendMessageWithFuture(envelope *msgenvelope.MsgEnvelope, timeout time.Duration) *actor.Future {
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)

	if lc.IsClosed() {
		msgenvelope.ReleaseMsgEnvelope(envelope)
		future.SetResult(nil, chaoserrors.NewErrCode(errcode.ServiceNotExist, "service rpc handler not exist"))
		return future
	}

	if envelope.NeedResp {
		lc.Add(future)
	}

	// 这里的envelope如果发生成功,将在rpcRequest处理完成后回收
	if err := lc.PushRequest(envelope); err != nil {
		//if err := rpcHandler.PushRequest(envelope); err != nil {
		// 发送错误,直接回收信封
		msgenvelope.ReleaseMsgEnvelope(envelope)
		// 设置错误结果
		future.SetResult(nil, chaoserrors.NewErrCode(errcode.RpcErr, err))
		return future
	}

	return future
}

func (lc *LClient) AsyncSendMessage(envelope *msgenvelope.MsgEnvelope, timeout time.Duration, completions []msgenvelope.CompletionFunc) (inf.CancelRpc, error) {
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)
	future.SetCompletions(completions...)

	if lc.IsClosed() {
		msgenvelope.ReleaseMsgEnvelope(envelope)
		actor.ReleaseFuture(future)
		return inf.EmptyCancelRpc, errdef.ServiceNotFound
	}

	cancelRpc := inf.EmptyCancelRpc
	if envelope.NeedResp {
		lc.Add(future)
		cancelRpc = NewRpcCancel(lc, envelope.ReqID)
	}

	// 这里的envelope如果发送成功,将在rpcRequest处理完成后回收
	if err := lc.PushRequest(envelope); err != nil {
		//if err := rpcHandler.PushRequest(envelope); err != nil {
		// 发送错误,直接回收信封
		msgenvelope.ReleaseMsgEnvelope(envelope)
		lc.Remove(future.GetReqID())
		actor.ReleaseFuture(future)
		return inf.EmptyCancelRpc, err
	}

	asynclib.Go(func() {
		// 异步等待结果后释放
		future.Wait()
		log.SysLogger.Debug("asynclib send message future call back, release future")
		actor.ReleaseFuture(future)
	})

	return cancelRpc, nil
}
