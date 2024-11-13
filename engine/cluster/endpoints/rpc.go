// Package endpoints
// @Title  title
// @Description  desc
// @Author  yr  2024/11/8
// @Update  yr  2024/11/8
package endpoints

import (
	"context"
	"errors"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/monitor"
	"github.com/njtc406/chaosengine/engine/msgenvelope"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/serializer"
)

type RPCResponse struct{}

type RPCListener struct{}

func (rm *RPCListener) RPCCall(_ context.Context, req *actor.Message, _ *RPCResponse) error {
	if req.Reply {
		// 回复
		// 需要回复的信息都会加入monitor中,找到对应的信封
		if envelope := monitor.GetRpcMonitor().Remove(req.ReqId); envelope != nil {
			// 解析回复数据
			response, err := serializer.Deserialize(req.Response, req.TypeName, req.TypeId)
			if err != nil {
				return err
			}
			envelope.Response = response
			envelope.Err = errors.New(req.Err)
			if envelope.NeedCallback() {
				// 异步回调,直接发送到对应服务处理,服务处理完后会自己释放envelope
				return envelope.SenderClient.PushRequest(envelope)
			} else {
				// 同步回调,回复结果
				envelope.Done()
				return nil
			}
		} else {
			// 已经超时,丢弃返回
			log.SysLogger.Warnf("rpc call timeout, future not found: %+v", req)
			msgenvelope.ReleaseMsgEnvelope(envelope)
			return nil
		}
	} else {
		// 调用
		// 构建消息
		envelope := msgenvelope.NewMsgEnvelope()
		envelope.Sender = req.Sender
		envelope.Receiver = req.Receiver
		envelope.Method = req.Method
		envelope.Reply = req.Reply
		envelope.ReqID = req.ReqId
		envelope.Header = req.MessageHeader
		envelope.NeedResp = req.NeedResp
		envelope.SetErrStr(req.Err)
		request, err := serializer.Deserialize(req.Request, req.TypeName, req.TypeId)
		if err != nil {
			return err
		}
		envelope.Request = request

		return endMgr.GetClient(req.Receiver.GetServiceUid()).SendMessage(envelope)
	}
}
