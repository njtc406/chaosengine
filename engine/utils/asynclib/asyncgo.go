// Package asynclib
// Mode Name: 异步执行
// Mode Desc: 使用协程池中的协程执行任务,防止出现瞬间创建大量协程,出现性能问题
package asynclib

import (
	"fmt"
	"github.com/njtc406/chaosutil/chaoserrors"
	"github.com/panjf2000/ants/v2"
)

// antsPool 协程池
var antsPool *ants.Pool

func InitAntsPool(size int) {
	if antsPool == nil {
		var err error
		antsPool, err = ants.NewPool(size, ants.WithPreAlloc(true))
		if err != nil {
			panic(err)
		}
	}
}

func Go(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			errString := fmt.Sprint(r)
			err = chaoserrors.NewErrCode(-1, errString, nil)
		}
	}()

	return antsPool.Submit(f)
}
