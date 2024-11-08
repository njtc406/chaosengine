// Package inf
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/7/29 下午4:47
// @Update  yr  2024/7/29 下午4:47
package inf

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"time"
)

func EmptyCancelRpc() {}

type IClient interface {
	IMonitor
	IRpcHandler
	Close()
	// SendMessage 发送消息
	SendMessage(envelope *actor.MsgEnvelope) error
	SendMessageWithFuture(envelope *actor.MsgEnvelope, timeout time.Duration) *actor.Future
	AsyncSendMessage(envelope *actor.MsgEnvelope, timeout time.Duration, completions []actor.CompletionFunc) (CancelRpc, error)
}
