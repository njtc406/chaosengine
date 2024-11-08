// Package inf
// @Title
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/9/4 下午7:49
// @Update  yr  2024/9/4 下午7:49
package inf

import (
	"github.com/njtc406/chaosengine/engine1/actor"
)

type ICallSet interface {
	AddPending(call *actor.Future)
	RemovePending(seq uint64) *actor.Future
	FindPending(seq uint64) *actor.Future
	GenerateSeq() uint64
}
