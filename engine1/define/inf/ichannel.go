// Package inf
// @Title  rpc通道
// @Description  rpc通道
// @Author  yr  2024/7/19 上午10:42
// @Update  yr  2024/7/19 上午10:42
package inf

import (
	"github.com/njtc406/chaosengine/engine1/actor"
)

type IChannel interface {
	PushRequest(msg *actor.MsgEnvelope) error
}
