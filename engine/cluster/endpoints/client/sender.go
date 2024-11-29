// Package client
// @Title  title
// @Description  desc
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package client

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/inf"
)

type SenderCreator func(pid *actor.PID, handler inf.IRpcHandler) inf.IRpcSender

var senderMap = map[string]SenderCreator{
	def.DefaultRpcTypeLocal: newLClient,
	def.DefaultRpcTypeRpcx:  newRemoteClient,
}

func Register(tp string, creator SenderCreator) {
	senderMap[tp] = creator
}

type SenderBase struct {
	pid *actor.PID
	inf.IRpcHandler
}

func (c *SenderBase) GetPID() *actor.PID {
	return c.pid
}

func NewSender(tp string) SenderCreator {
	creator, ok := senderMap[tp]
	if !ok {
		return nil
	}
	return creator
}
