// Package asynclib
// @Title  异步库
// @Description  异步库
// @Author  yr  2024/6/27 下午4:05
// @Update  yr  2024/6/27 下午4:05
package asynclib

import (
	"github.com/panjf2000/ants/v2"
)

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
	return antsPool.Submit(f)
}
