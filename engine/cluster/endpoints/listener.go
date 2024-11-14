// Package endpoints
// @Title  title
// @Description  desc
// @Author  yr  2024/11/8
// @Update  yr  2024/11/8
package endpoints

import (
	"context"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/monitor"
	"github.com/njtc406/chaosengine/engine/msgenvelope"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/serializer"
)

type RPCListener struct{}

// TODO 处理ReqId重复发送的问题

func (rm *RPCListener) RPCCall(_ context.Context, req *actor.Message, _ *def.RPCResponse) error {
	if req.Reply {
		// 回复
		// 需要回复的信息都会加入monitor中,找到对应的信封
		if envelope := monitor.GetRpcMonitor().Remove(req.ReqId); envelope != nil {
			// 异步回调,直接发送到对应服务处理,服务处理完后会自己释放envelope
			sender := envelope.GetSenderClient()
			if sender != nil && sender.IsClosed() {
				// 调用者已经下线,丢弃回复
				msgenvelope.ReleaseMsgEnvelope(envelope)
				return nil
			}
			// 解析回复数据
			response, err := serializer.Deserialize(req.Response, req.TypeName, req.TypeId)
			if err != nil {
				msgenvelope.ReleaseMsgEnvelope(envelope)
				return err
			}
			envelope.SetReply()
			envelope.SetRequest(nil)
			envelope.SetResponse(response)
			envelope.SetErrStr(req.Err)
			if envelope.NeedCallback() {
				return sender.PushRequest(envelope)
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
		request, err := serializer.Deserialize(req.Request, req.TypeName, req.TypeId)
		if err != nil {
			return err
		}

		mt := monitor.GetRpcMonitor()

		// 构建消息
		envelope := msgenvelope.NewMsgEnvelope()
		envelope.SetMethod(req.Method)
		envelope.SetReceiver(req.Receiver)
		envelope.SetSenderClient(endMgr.GetClient(req.Sender.GetServiceUid()))
		envelope.SetRequest(request)
		envelope.SetResponse(nil)
		envelope.SetReqId(mt.GenSeq())
		envelope.SetNeedResponse(req.NeedResp)

		return endMgr.GetClient(req.Receiver.GetServiceUid()).SendRequestAndRelease(envelope)
	}
}
