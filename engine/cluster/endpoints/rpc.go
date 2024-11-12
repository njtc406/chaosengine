// Package endpoints
// @Title  title
// @Description  desc
// @Author  yr  2024/11/8
// @Update  yr  2024/11/8
package endpoints

import (
	"context"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/serializer"
)

type RPCResponse struct{}

type RPCListener struct{}

func (rm *RPCListener) RPCCall(ctx context.Context, req *actor.Message, _ *RPCResponse) error {
	// 根据receiver的pid找到对应的client
	client := endMgr.GetClient(req.Receiver.GetServiceUid())
	if client == nil {
		// 服务已经下线
		log.SysLogger.Warnf("client not found: %+v", req)
		return errdef.ServiceNotFound
	}

	// 构建消息
	envelope := actor.NewMsgEnvelope()
	envelope.Sender = req.Sender
	envelope.Receiver = req.Receiver
	envelope.Method = req.Method
	envelope.Reply = req.Reply
	envelope.ReqID = req.ReqId
	envelope.Header = req.MessageHeader
	envelope.SetErrStr(req.Err)

	// 反序列化

	if req.Reply {
		response, err := serializer.Deserialize(req.Response, req.TypeName, req.TypeId)
		if err != nil {
			return err
		}
		envelope.Response = response
	} else {
		request, err := serializer.Deserialize(req.Request, req.TypeName, req.TypeId)
		if err != nil {
			return err
		}
		envelope.Request = request
	}

	if envelope.IsReply() {
		// 如果是回复,先找到对应的future
		if future := client.Get(envelope.ReqID); future != nil {
			if future.NeedCallback() {
				// 异步回调,直接发送到对应handler处理
				// envelope会在send之后自动回收
				return client.SendMessage(envelope)
			} else {
				// 同步回调,直接向future设置结果
				future.SetResult(envelope.Response, envelope.GetError())
				actor.ReleaseMsgEnvelope(envelope)
				return nil
			}
		} else {
			// 已经超时,丢弃返回
			log.SysLogger.Warnf("rpc call timeout, future not found: %+v", req)
			actor.ReleaseMsgEnvelope(envelope)
			return nil
		}
	}

	return client.SendMessage(envelope)
}
