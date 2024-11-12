// Package client
// @Title  远程服务的Client
// @Description  远程服务的Client
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package client

import (
	"context"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"
	"time"

	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/asynclib"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/serializer"
	"github.com/smallnest/rpcx/client"
)

// 远程服务的Client

type RemoteClient struct {
	Client
	rpcClient client.XClient
}

func NewRemoteClient(pid *actor.PID, monitor inf.IMonitor) inf.IClient {
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
	remoteClient.IMonitor = monitor
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
}

func (rc *RemoteClient) send(envelope *actor.MsgEnvelope, timeout time.Duration) error {
	// 这里仅仅代表消息发送成功
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 回收envelope
	defer actor.ReleaseMsgEnvelope(envelope)

	// 构建发送消息
	msg := &actor.Message{
		TypeId:        0, // 默认使用protobuf(后面有其他需求再修改这里)
		Sender:        envelope.Sender,
		Receiver:      envelope.Receiver,
		Method:        envelope.Method,
		MessageHeader: envelope.Header,
		Reply:         envelope.Reply,
		ReqId:         envelope.ReqID,
		Err:           envelope.GetErrStr(),
	}

	byteData, typeName, err := serializer.Serialize(envelope.Request, msg.TypeId)
	if err != nil {
		log.SysLogger.Errorf("serialize message[%+v] is error: %s", envelope, err)
		return err
	}
	msg.Request = byteData
	msg.TypeName = typeName

	if _, err = rc.rpcClient.Go(ctx, "RPCCall", msg, nil, nil); err != nil {
		log.SysLogger.Errorf("send message[%v] to %s is error: %s", envelope, rc.GetPID().GetServiceUid(), err)
		return err
	}

	return nil
}

func (rc *RemoteClient) SendMessage(envelope *actor.MsgEnvelope) error {
	return rc.send(envelope, def.DefaultRpcTimeout)
}

func (rc *RemoteClient) SendMessageWithFuture(envelope *actor.MsgEnvelope, timeout time.Duration) *actor.Future {
	if timeout == 0 {
		timeout = def.DefaultRpcTimeout
	}
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)

	if timeout > 0 {
		rc.Add(future)
	}

	if err := rc.send(envelope, timeout); err != nil {
		log.SysLogger.Errorf("remote send message[%v] to %s is error: %s", envelope, rc.GetPID().GetServiceUid(), err)
		future.SetResult(nil, err)
		return future
	}

	return future
}

func (rc *RemoteClient) AsyncSendMessage(envelope *actor.MsgEnvelope, timeout time.Duration, completions []actor.CompletionFunc) (inf.CancelRpc, error) {
	if timeout == 0 {
		timeout = def.DefaultRpcTimeout
	}
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)
	future.SetCompletions(completions...)

	cancelRpc := inf.EmptyCancelRpc
	if timeout > 0 {
		rc.Add(future)
		cancelRpc = NewRpcCancel(rc, envelope.ReqID)
	}

	if err := rc.send(envelope, timeout); err != nil {
		rc.Remove(envelope.ReqID)
		actor.ReleaseFuture(future)
		return cancelRpc, err
	}

	asynclib.Go(func() {
		future.Wait()
		rc.Remove(future.GetReqID())
		actor.ReleaseFuture(future)
	})

	return cancelRpc, nil
}
