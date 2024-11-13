// Package inf
// @Title 调用监控
// @Description  用于监控一些具有超时属性的调用
// @Author  yr  2024/9/4 下午7:49
// @Update  yr  2024/9/4 下午7:49
package inf

import (
	"github.com/njtc406/chaosengine/engine/msgenvelope"
)

type IMonitor interface {
	Init(fun func(f *msgenvelope.MsgEnvelope))
	Start()
	Stop()
	Add(call *msgenvelope.MsgEnvelope)
	Remove(seq uint64) *msgenvelope.MsgEnvelope
	Get(seq uint64) *msgenvelope.MsgEnvelope
	GenSeq() uint64
}
