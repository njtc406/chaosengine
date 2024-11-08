// Package ip2regoin
// @Title
// @Description  ip2regoin
// @Author sly 2024/8/31
// @Created sly 2024/8/31
package ip2regoin

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine1/ip2regoin/xdb"
)

var XSearcher *xdb.Searcher

func InitXdbSearcher(dbPath string) {
	// 1、从 dbPath 加载整个 xdb 到内存
	cBuff, err := xdb.LoadContentFromFile(dbPath)
	if err != nil {
		fmt.Printf("failed to load content from `%s`: %s\n", dbPath, err)
		return
	}

	// 2、用全局的 cBuff 创建完全基于内存的查询对象。
	XSearcher, err = xdb.NewWithBuffer(cBuff)
	if err != nil {
		fmt.Printf("failed to create searcher with content: %s\n", err)
		return
	}

	return
}
