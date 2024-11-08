// Package inf
// @Title 调用监控
// @Description  用于监控一些具有超时属性的调用
// @Author  yr  2024/9/4 下午7:49
// @Update  yr  2024/9/4 下午7:49
package inf

import (
	"github.com/njtc406/chaosengine/engine/actor"
)

type IMonitor interface {
	Init(fun func(f *actor.Future))
	Start()
	Stop()
	Add(call *actor.Future)
	Remove(seq uint64) *actor.Future
	Get(seq uint64) *actor.Future
	GenSeq() uint64
}
