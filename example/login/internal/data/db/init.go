// Package db
// 模块名: 数据模块
// 功能描述:
// 作者:  yr  2024/11/16 0016 17:42
// 最后更新:  yr  2024/11/16 0016 17:42
package db

import (
	"github.com/njtc406/chaosengine/engine/sysModule/mysqlmodule"
)

func init() {
	mysqlmodule.RegisterTable(
		&User{},
	)
}
