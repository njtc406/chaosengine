// Package client
// @Title  title
// @Description  desc
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package client

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
)

type SenderBase struct {
	pid *actor.PID
	inf.IRpcHandler
}

func (c *SenderBase) GetPID() *actor.PID {
	return c.pid
}
