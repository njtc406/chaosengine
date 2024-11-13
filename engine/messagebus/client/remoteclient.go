// Package client
// @Title  远程服务的Client
// @Description  远程服务的Client
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package client

import (
	"context"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/msgenvelope"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"
	"time"

	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/smallnest/rpcx/client"
)

// 远程服务的Client

type RemoteClient struct {
	Client
	rpcClient client.XClient
}

func NewRemoteClient(pid *actor.PID) inf.IClient {
	d, _ := client.NewPeer2PeerDiscovery("tcp@"+pid.GetAddress(), "")
	// 如果调用失败,会自动重试3次
	rpcClient := client.NewXClient("RPCMonitor", client.Failtry, client.RandomSelect, d, client.Option{
		Retries:             3, // 重试3次
		RPCPath:             share.DefaultRPCPath,
		ConnectTimeout:      time.Second,           // 连接超时
		SerializeType:       protocol.ProtoBuffer,  // 序列化方式
		CompressType:        protocol.None,         // 压缩方式
		BackupLatency:       50 * time.Millisecond, // 延迟时间(第一个请求在这个时间内没有回复,则会发送第二次请求)
		MaxWaitForHeartbeat: 30 * time.Second,      // 心跳时间
		TCPKeepAlivePeriod:  time.Minute,           // tcp keepalive
		BidirectionalBlock:  false,                 // 是否允许双向阻塞(true代表发送过去的消息必须消费之后才会再次发送,否则通道阻塞)
		TimeToDisallow:      time.Minute,
	})

	remoteClient := &RemoteClient{
		rpcClient: rpcClient,
	}
	remoteClient.pid = pid

	log.SysLogger.Infof("create remote client success : %s", pid.String())
	return remoteClient
}

func (rc *RemoteClient) Close() {
	if rc.rpcClient == nil {
		return
	}
	if err := rc.rpcClient.Close(); err != nil {
		log.SysLogger.Errorf("close remote client is error : %s", err)
	}
	rc.rpcClient = nil
}

func (rc *RemoteClient) send(envelope *msgenvelope.MsgEnvelope) error {
	if rc.rpcClient == nil {
		return errdef.RPCHadClosed
	}
	// 这里仅仅代表消息发送成功
	ctx, cancel := context.WithTimeout(context.Background(), envelope.Timeout)
	defer cancel()

	// 构建发送消息
	msg := envelope.ToProtoMsg()
	if msg == nil {
		return errdef.MsgSerializeFailed
	}

	if _, err := rc.rpcClient.Go(ctx, "RPCCall", msg, nil, nil); err != nil {
		log.SysLogger.Errorf("send message[%v] to %s is error: %s", envelope, rc.GetPID().GetServiceUid(), err)
		return err
	}

	return nil
}

func (rc *RemoteClient) SendMessage(envelope *msgenvelope.MsgEnvelope) error {
	// 回收envelope
	return rc.send(envelope)
}

func (rc *RemoteClient) SendMessageWithFuture(envelope *msgenvelope.MsgEnvelope) {
	if err := rc.send(envelope); err != nil {
		log.SysLogger.Errorf("remote send message[%v] to %s is error: %s", envelope, rc.GetPID().GetServiceUid(), err)
		envelope.Err = errdef.RPCCallFailed
		envelope.Done()
		return
	}

	return
}

func (rc *RemoteClient) AsyncSendMessage(envelope *msgenvelope.MsgEnvelope, completions []msgenvelope.CompletionFunc) (inf.CancelRpc, error) {
	defer msgenvelope.ReleaseMsgEnvelope(envelope)

	cancelRpc := inf.EmptyCancelRpc
	if envelope.NeedResp {
		rc.Add(envelope)
		cancelRpc = NewRpcCancel(rc, envelope.ReqID)
	}

	if err := rc.send(envelope); err != nil {
		rc.Remove(envelope.ReqID)
		return inf.EmptyCancelRpc, err
	}

	return cancelRpc, nil
}
