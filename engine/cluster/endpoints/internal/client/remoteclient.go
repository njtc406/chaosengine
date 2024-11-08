// Package client
// @Title  远程服务的Client
// @Description  远程服务的Client
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package client

import (
	"context"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/errdef"
	"sync"
	"time"

	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/asynclib"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/serializer"
	"github.com/smallnest/rpcx/client"
	"google.golang.org/protobuf/proto"
)

// 远程服务的Client

type RemoteClient struct {
	Client
	rpcClient client.XClient
}

func NewRemoteClient(pid *actor.PID, monitor inf.IMonitor) inf.IClient {
	d, _ := client.NewPeer2PeerDiscovery("tcp@"+pid.GetAddress(), "")
	rpcClient := client.NewXClient("RPCMonitor", client.Failtry, client.RandomSelect, d, client.DefaultOption)

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

func (rc *RemoteClient) send(envelope *actor.MsgEnvelope) error {
	// 这里仅仅代表消息发送成功
	ctx := context.Background()
	// 回收envelope
	defer actor.ReleaseMsgEnvelope(envelope)

	// 构建发送消息
	msg := &actor.Message{
		TypeId:        0, // 默认使用protobuf(后面有其他需求再修改这里)
		Receiver:      proto.Clone(envelope.Receiver).(*actor.PID),
		Method:        envelope.Method,
		MessageHeader: envelope.Header,
		Reply:         envelope.Reply,
		ReqId:         envelope.ReqID,
		IsRpc:         envelope.IsRpc,
		Err:           envelope.GetErrStr(),
	}
	//log.SysLogger.Debugf("----------->>>>>>>>>>send message[%+v] to %s", envelope, rc.c.GetPID().GetId())
	if envelope.Sender != nil {
		msg.Sender = proto.Clone(envelope.Sender).(*actor.PID)
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

func (rc *RemoteClient) call(envelope *actor.MsgEnvelope, timeout time.Duration) error {
	// 这里仅仅代表消息发送成功
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 回收envelope
	defer actor.ReleaseMsgEnvelope(envelope)

	errCh := make(chan error, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	asynclib.Go(func() {
		defer wg.Done()
		// 构建发送消息
		msg := &actor.Message{
			TypeId:        0, // 默认使用protobuf(后面有其他需求再修改这里)
			Receiver:      proto.Clone(envelope.Receiver).(*actor.PID),
			Method:        envelope.Method,
			MessageHeader: envelope.Header,
			Reply:         envelope.Reply,
			ReqId:         envelope.ReqID,
			IsRpc:         envelope.IsRpc,
			Err:           envelope.GetErrStr(),
		}

		if envelope.Sender != nil {
			msg.Sender = proto.Clone(envelope.Sender).(*actor.PID)
		}

		byteData, typeName, err := serializer.Serialize(envelope.Request, msg.TypeId)
		if err != nil {
			log.SysLogger.Errorf("serialize message[%+v] is error: %s", envelope, err)
			errCh <- err
			return
		}
		msg.Request = byteData
		msg.TypeName = typeName

		errCh <- rc.rpcClient.Call(ctx, "RPCCall", msg, nil)
	})
	wg.Wait()
	select {
	case err := <-errCh:
		if err != nil {
			log.SysLogger.Errorf("send message[%v] to %s is error: %s", envelope, rc.GetPID().GetServiceUid(), err)
			return err
		}
	case <-ctx.Done():
		log.SysLogger.Errorf("send message[%v] to %s is timeout", envelope, rc.GetPID().GetServiceUid())
		return errdef.RPCCallTimeout
	}

	return nil
}

func (rc *RemoteClient) SendMessage(envelope *actor.MsgEnvelope) error {
	return rc.send(envelope)
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

	if err := rc.call(envelope, timeout); err != nil {
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

	if err := rc.call(envelope, timeout); err != nil {
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
