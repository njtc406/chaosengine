// Package endpoints
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/7/29 下午4:47
// @Update  yr  2024/7/29 下午4:47
package endpoints

import (
	"github.com/njtc406/chaosengine/engine1/actor"
	interfacedef2 "github.com/njtc406/chaosengine/engine1/define/inf"
	"github.com/njtc406/chaosutil/chaoserrors"
	"time"
)

const (
	DefaultRpcConnNum           = 1
	DefaultRpcLenMsgLen         = 4
	DefaultRpcMinMsgLen         = 2
	DefaultMaxCheckCallRpcCount = 1000
	DefaultMaxPendingWriteNum   = 1000000

	DefaultConnectInterval             = 2 * time.Second
	DefaultCheckRpcCallTimeoutInterval = 1 * time.Second
	DefaultRpcTimeout                  = 15 * time.Second
)

var clientSeq uint32

type IWriter interface {
	WriteMsg(nodeId string, args ...[]byte) error
	IsConnected() bool
}

func EmptyCancelRpc() {}

type IRealClient interface {
	Close()
	// SendMessage 发送消息
	SendMessage(envelope *actor.MsgEnvelope) chaoserrors.CError
	SendMessageWithFuture(envelope *actor.MsgEnvelope, timeout time.Duration, reply interface{}) *actor.Future
	AsyncSendMessage(envelope *actor.MsgEnvelope, timeout time.Duration, completions []actor.CompletionFunc) (interfacedef2.CancelRpc, chaoserrors.CError)
}

type Client struct {
	pid *actor.PID
	//compressBytesLen int

	interfacedef2.ICallSet
	IRealClient
	interfacedef2.IRpcHandler
}

func NewClient(pid *actor.PID, callSet interfacedef2.ICallSet, realClient IRealClient, rpcHandler interfacedef2.IRpcHandler) *Client {
	return &Client{
		pid:         pid,
		ICallSet:    callSet,
		IRealClient: realClient,
		IRpcHandler: rpcHandler,
	}
}

func (c *Client) GetPID() *actor.PID {
	return c.pid
}

func (c *Client) GetRpcHandler() interfacedef2.IRpcHandler {
	return c.IRpcHandler
}
