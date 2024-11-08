// Package endpoints
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package endpoints

import (
	"context"
	"github.com/njtc406/chaosengine/engine1/actor"
	"github.com/njtc406/chaosengine/engine1/errdef/errcode"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosutil/chaoserrors"
	"github.com/smallnest/rpcx/server"
)

type Remote struct {
	address string // 服务监听地址
	nodeUID string // 节点唯一ID
	svr     *server.Server
}

func NewRemote(nodeUID, address string) *Remote {
	return &Remote{
		address: address,
		nodeUID: nodeUID,
	}
}

func (r *Remote) Init() {
	r.svr = server.NewServer()
}

func (r *Remote) Serve() error {
	// 注册rpc监听服务
	if err := r.svr.Register(new(RPCMonitor), ""); err != nil {
		return err
	}

	go func() {
		if err := r.svr.Serve("tcp", r.address); err != nil {
			log.SysLogger.Errorf("rpc serve stop: %v", err)
		}
	}()

	return nil
}

func (r *Remote) GetAddress() string {
	return r.address
}

func (r *Remote) GetNodeUID() string {
	return r.nodeUID
}

type RPCResponse struct{}

type RPCMonitor struct{}

func (rm *RPCMonitor) RPCCall(ctx context.Context, req *actor.Message, _ *RPCResponse) error {
	// 根据receiver的pid找到对应的client
	client := endpointManager.GetClientByPID(req.Receiver)
	if client == nil {
		// 服务已经下线
		log.SysLogger.Warnf("client not found: %+v", req)
		return chaoserrors.NewErrCode(errcode.ServiceMethodNotExist, "服务不存在")
	}

	// 构建消息
	envelope := actor.NewMsgEnvelope()
	envelope.Sender = req.Sender
	envelope.Receiver = req.Receiver
	envelope.Method = req.Method
	envelope.Reply = req.Reply
	envelope.IsRpc = req.IsRpc
	envelope.ReqID = req.ReqID
	envelope.Header = req.MessageHeader

	// 反序列化
	request, err := Deserialize(req.Request, req.TypeName, req.TypeID)
	if err != nil {
		return err
	}
	envelope.Request = request

	if envelope.IsReply() {
		// 如果是回复,先找到对应的future
		if future := client.FindPending(envelope.ReqID); future != nil {
			if future.NeedCallback() {
				// 异步回调,直接发送到对应handler处理
				// envelope会在send之后自动回收
				return client.SendMessage(envelope)
			} else {
				// 同步回调,直接向future设置结果
				var resp []interface{}
				if envelope.Request != nil {
					resp = []interface{}{envelope.Request}
				}
				var cErr chaoserrors.CError
				if envelope.Err != "" {
					cErr = chaoserrors.NewErrCode(errcode.RpcCallResponseErr, envelope.Err)
				}
				actor.ReleaseMsgEnvelope(envelope)
				future.SetResult(resp, cErr)
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
