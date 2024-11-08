// Package endpoints
// @Title  本地服务的Client
// @Description  本地服务的Client
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package endpoints

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine1/actor"
	interfacedef2 "github.com/njtc406/chaosengine/engine1/define/inf"
	"github.com/njtc406/chaosengine/engine1/errdef/errcode"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosengine/engine1/utils/asynclib"
	"github.com/njtc406/chaosutil/chaoserrors"
	"time"
)

// LClient 本地服务的Client
type LClient struct {
	selfClient *Client
}

func NewLClient(pid *actor.PID, callSet interfacedef2.ICallSet, rpcHandler interfacedef2.IRpcHandler) *Client {
	lClient := &LClient{}
	client := NewClient(pid, callSet, lClient, rpcHandler)
	client.IRealClient = lClient
	client.ICallSet = callSet
	client.pid = pid
	lClient.selfClient = client
	return client
}

func (lc *LClient) Close() {}

func (lc *LClient) SendMessage(envelope *actor.MsgEnvelope) chaoserrors.CError {
	//log.SysLogger.Debugf("is reply: %v, is rpc: %v, need callback: %v", envelope.IsReply(), envelope.IsRpc, envelope.NeedCallback())
	if envelope.IsReply() && !envelope.IsRpc && !envelope.NeedCallback() {
		defer actor.ReleaseMsgEnvelope(envelope)
		// 本地Call的回复, 通过reqId找到future,直接返回结果
		future := lc.selfClient.FindPending(envelope.ReqID)
		if future == nil {
			// 已经超时了,丢弃
			return chaoserrors.NewErrCode(errcode.RpcCallTimeout, "service method call timeout")
		}
		var cErr chaoserrors.CError
		if envelope.Err != "" {
			log.SysLogger.Errorf("envelope.err reqID %d method %s sender %v receiver %v is_reply %v is_rpc %v need_callback %v errdef [%#v]", envelope.ReqID, envelope.Method, envelope.Sender, envelope.Receiver, envelope.IsReply(), envelope.IsRpc, envelope.NeedCallback(), envelope.Err)
			cErr = chaoserrors.NewErrCode(errcode.RpcCallResponseErr, envelope.Err)
		}
		future.SetResult(envelope.Request.([]interface{}), cErr)

		return nil
	}

	// 根据自己的pid找到对应的rpcHandler
	rpcHandler := lc.selfClient.GetRpcHandler()
	if rpcHandler == nil {
		return chaoserrors.NewErrCode(errcode.ServiceMethodError, fmt.Sprintf("service %s can not found", lc.selfClient.GetPID().GetName()), nil)
	}

	// 直接push消息(envelope会在对应service的rpcHandler调用完成后回收)
	return rpcHandler.PushRequest(envelope)
}

func (lc *LClient) SendMessageWithFuture(envelope *actor.MsgEnvelope, timeout time.Duration, reply interface{}) *actor.Future {
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)

	// 根据自己的pid找到对应的rpcHandler
	rpcHandler := lc.selfClient.GetRpcHandler()
	//log.SysLogger.Errorf("SendMessageWithFuture  method-[%s] sender-[%v]  completions-[%v] reqID-[%v] rpcHandler-[%v]", future.GetMethod(), future.GetSender(), future.GetCompletions(), future.GetReqID(), rpcHandler.GetName())
	if rpcHandler == nil {
		actor.ReleaseMsgEnvelope(envelope)
		future.SetResult(nil, chaoserrors.NewErrCode(errcode.ServiceNotExist, "service rpc handler not exist"))
		return future
	}

	if timeout > 0 {
		lc.selfClient.AddPending(future)
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

func (lc *LClient) AsyncSendMessage(envelope *actor.MsgEnvelope, timeout time.Duration, completions []actor.CompletionFunc) (interfacedef2.CancelRpc, chaoserrors.CError) {
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)
	future.SetCompletions(completions...)

	// 根据自己的pid找到对应的rpcHandler
	rpcHandler := lc.selfClient.GetRpcHandler()
	if rpcHandler == nil {
		actor.ReleaseMsgEnvelope(envelope)
		actor.ReleaseFuture(future)
		return EmptyCancelRpc, chaoserrors.NewErrCode(errcode.ServiceNotExist, "service rpc handler not exist")
	}

	cancelRpc := EmptyCancelRpc
	if timeout > 0 {
		lc.selfClient.AddPending(future)
		cancelRpc = NewRpcCancel(lc.selfClient, envelope.ReqID)
	}

	// 这里的envelope如果发生成功,将在rpcRequest处理完成后回收
	if err := rpcHandler.PushRequest(envelope); err != nil {
		// 发送错误,直接回收信封
		actor.ReleaseMsgEnvelope(envelope)
		lc.selfClient.RemovePending(envelope.ReqID)
		actor.ReleaseFuture(future)
		return EmptyCancelRpc, chaoserrors.NewErrCode(errcode.RpcErr, err)
	}

	asynclib.Go(func() {
		// 异步等待结果后释放
		future.Wait()
		log.SysLogger.Debug("asynclib send message future call back, release future")
		lc.selfClient.RemovePending(future.GetReqID())
		actor.ReleaseFuture(future)
	})

	return cancelRpc, nil
}
