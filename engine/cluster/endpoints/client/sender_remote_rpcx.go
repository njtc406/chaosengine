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

type rpcxSender struct {
	SenderBase
	rpcClient client.XClient
}

func newRemoteClient(pid *actor.PID, _ inf.IRpcHandler) inf.IRpcSender {
	d, _ := client.NewPeer2PeerDiscovery("tcp@"+pid.GetAddress(), "")
	// 如果调用失败,会自动重试3次
	rpcClient := client.NewXClient("RpcListener", client.Failtry, client.RandomSelect, d, client.Option{
		Retries:             3, // 重试3次
		RPCPath:             share.DefaultRPCPath,
		ConnectTimeout:      time.Second,           // 连接超时
		SerializeType:       protocol.ProtoBuffer,  // 序列化方式
		CompressType:        protocol.None,         // 压缩方式
		BackupLatency:       50 * time.Millisecond, // 延迟时间(上一个请求在这个时间内没有回复,则会发送第二次请求) 这个需要考虑一下
		MaxWaitForHeartbeat: 30 * time.Second,      // 心跳时间
		TCPKeepAlivePeriod:  time.Minute,           // tcp keepalive
		BidirectionalBlock:  false,                 // 是否允许双向阻塞(true代表发送过去的消息必须消费之后才会再次发送,否则通道阻塞)
		TimeToDisallow:      time.Minute,
	})

	remoteClient := &rpcxSender{
		rpcClient: rpcClient,
	}
	remoteClient.pid = pid

	log.SysLogger.Infof("create remote client success : %s", pid.String())
	return remoteClient
}

func (rc *rpcxSender) Close() {
	if rc.rpcClient == nil {
		return
	}
	if err := rc.rpcClient.Close(); err != nil {
		log.SysLogger.Errorf("close remote client is error : %s", err)
	}
	rc.rpcClient = nil
}

func (rc *rpcxSender) send(envelope inf.IEnvelope) error {
	if rc.rpcClient == nil {
		return errdef.RPCHadClosed
	}
	// 这里仅仅代表消息发送成功
	ctx, cancel := context.WithTimeout(context.Background(), envelope.GetTimeout())
	defer cancel()

	// 构建发送消息
	msg := envelope.ToProtoMsg()
	if msg == nil {
		return errdef.MsgSerializeFailed
	}

	if _, err := rc.rpcClient.Go(ctx, "RPCCall", msg, nil, nil); err != nil {
		log.SysLogger.Errorf("send message[%+v] to %s is error: %s", envelope, rc.GetPID().GetServiceUid(), err)
		return errdef.RPCCallFailed
	}

	return nil
}

func (rc *rpcxSender) SendRequest(envelope inf.IEnvelope) error {
	// 这里不能释放envelope,因为调用方需要使用
	return rc.send(envelope)
}

func (rc *rpcxSender) SendRequestAndRelease(envelope inf.IEnvelope) error {
	defer msgenvelope.ReleaseMsgEnvelope(envelope)
	return rc.send(envelope)
}

func (rc *rpcxSender) SendResponse(envelope inf.IEnvelope) error {
	defer msgenvelope.ReleaseMsgEnvelope(envelope)
	return rc.send(envelope)
}
