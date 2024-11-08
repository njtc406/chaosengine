// Package endpoints
// @Title  远程服务的Client
// @Description  远程服务的Client
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package endpoints

import (
	"context"
	"github.com/njtc406/chaosengine/engine1/actor"
	interfacedef2 "github.com/njtc406/chaosengine/engine1/define/inf"
	"github.com/njtc406/chaosengine/engine1/errdef/errcode"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosengine/engine1/utils/asynclib"
	"github.com/njtc406/chaosutil/chaoserrors"
	"github.com/smallnest/rpcx/client"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

// 远程服务的Client

type RemoteClient struct {
	selfClient *Client
	rpcClient  client.XClient
}

func NewRemoteClient(pid *actor.PID, callSet interfacedef2.ICallSet, rpcHandler interfacedef2.IRpcHandler) *Client {
	d, _ := client.NewPeer2PeerDiscovery("tcp@"+pid.GetAddress(), "")
	rpcClient := client.NewXClient("RPCMonitor", client.Failtry, client.RandomSelect, d, client.DefaultOption)

	remoteClient := &RemoteClient{
		rpcClient: rpcClient,
	}
	c := NewClient(pid, callSet, remoteClient, rpcHandler)
	c.IRealClient = remoteClient
	c.ICallSet = callSet
	c.pid = pid
	remoteClient.selfClient = c

	log.SysLogger.Infof("create remote client success : %s", pid.String())
	return c
}

func (rc *RemoteClient) Close() {
	if rc.rpcClient == nil {
		return
	}
	if err := rc.rpcClient.Close(); err != nil {
		log.SysLogger.Errorf("close remote client is error : %s", err)
	}
}

func (rc *RemoteClient) send(envelope *actor.MsgEnvelope, timeout time.Duration) chaoserrors.CError {
	if timeout == 0 {
		timeout = time.Second
	}
	// 这里仅仅代表消息发送成功
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 回收envelope
	defer actor.ReleaseMsgEnvelope(envelope)

	errCh := make(chan error, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 构建发送消息
		msg := &actor.Message{
			TypeID:        0, // 默认使用protobuf(后面有其他需求再修改这里)
			Receiver:      proto.Clone(envelope.Receiver).(*actor.PID),
			Method:        envelope.Method,
			MessageHeader: envelope.Header,
			Reply:         envelope.Reply,
			ReqID:         envelope.ReqID,
			IsRpc:         envelope.IsRpc,
			Err:           envelope.Err,
		}
		//log.SysLogger.Debugf("----------->>>>>>>>>>send message[%+v] to %s", envelope, rc.selfClient.GetPID().GetId())
		if envelope.Sender != nil {
			msg.Sender = proto.Clone(envelope.Sender).(*actor.PID)
		}

		byteData, typeName, err := Serialize(envelope.Request, msg.TypeID)
		if err != nil {
			log.SysLogger.Errorf("serialize message[%+v] is error: %s", envelope, err)
			errCh <- err
			return
		}
		msg.Request = byteData
		msg.TypeName = typeName

		errCh <- rc.rpcClient.Call(ctx, "RPCCall", msg, nil)
		//log.SysLogger.Debugf("----------->>>>>>>>>>send message[%+v] to %s is success", msg, rc.selfClient.GetPID().GetId())
	}()
	wg.Wait()
	select {
	case err := <-errCh:
		if err != nil {
			log.SysLogger.Errorf("send message[%v] to %s is error: %s", envelope, rc.selfClient.GetPID().GetId(), err)
			return chaoserrors.NewErrCode(errcode.RpcErr, err)
		}
	case <-ctx.Done():
		log.SysLogger.Errorf("send message[%v] to %s is timeout", envelope, rc.selfClient.GetPID().GetId())
		return chaoserrors.NewErrCode(errcode.RpcCallTimeout, "service method call timeout")
	}

	return nil
}

func (rc *RemoteClient) SendMessage(envelope *actor.MsgEnvelope) chaoserrors.CError {
	return rc.send(envelope, 0)
}

func (rc *RemoteClient) SendMessageWithFuture(envelope *actor.MsgEnvelope, timeout time.Duration, reply interface{}) *actor.Future {
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)

	if timeout > 0 {
		rc.selfClient.AddPending(future)
	}

	if err := rc.send(envelope, timeout); err != nil {
		log.SysLogger.Errorf("remote send message[%v] to %s is error: %s", envelope, rc.selfClient.GetPID().GetId(), err)
		future.SetResult(nil, err)
		return future
	}

	//log.SysLogger.Debugf("remote send message[%+v] to %s is success", envelope, rc.selfClient.GetPID().GetId())

	return future
}

func (rc *RemoteClient) AsyncSendMessage(envelope *actor.MsgEnvelope, timeout time.Duration, completions []actor.CompletionFunc) (interfacedef2.CancelRpc, chaoserrors.CError) {
	future := actor.NewFuture()
	future.SetSender(envelope.Sender)
	future.SetMethod(envelope.Method)
	future.SetTimeout(timeout)
	future.SetReqID(envelope.ReqID)
	future.SetCompletions(completions...)

	cancelRpc := EmptyCancelRpc
	if timeout > 0 {
		rc.selfClient.AddPending(future)
		cancelRpc = NewRpcCancel(rc.selfClient, envelope.ReqID)
	}

	if err := rc.send(envelope, timeout); err != nil {
		rc.selfClient.RemovePending(envelope.ReqID)
		actor.ReleaseFuture(future)
		return cancelRpc, err
	}

	asynclib.Go(func() {
		future.Wait()
		rc.selfClient.RemovePending(future.GetReqID())
		actor.ReleaseFuture(future)
	})

	return cancelRpc, nil
}
