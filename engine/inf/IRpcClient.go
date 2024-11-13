// Package inf
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/7/29 下午4:47
// @Update  yr  2024/7/29 下午4:47
package inf

import (
	"github.com/njtc406/chaosengine/engine/msgenvelope"
)

func EmptyCancelRpc() {}

type IClient interface {
	IRpcHandler
	Close()
	// SendMessage 发送消息
	SendMessage(envelope *msgenvelope.MsgEnvelope) error
	SendMessageWithFuture(envelope *msgenvelope.MsgEnvelope)
	AsyncSendMessage(envelope *msgenvelope.MsgEnvelope, completions []msgenvelope.CompletionFunc) (CancelRpc, error)
}
