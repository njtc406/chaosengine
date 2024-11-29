// Package inf
// 模块名: 模块名
// 功能描述: 描述
// 作者:  yr  2024/11/16 0016 17:39
// 最后更新:  yr  2024/11/16 0016 17:39
package inf

import (
	sysInf "github.com/njtc406/chaosengine/engine/inf"
	"server/login/internal/data/db"
)

type ISdk interface {
	Login(rpcHandler sysInf.IRpcHandler, req interface{}) *db.User
	DeleteAccount(rpcHandler sysInf.IRpcHandler, req interface{}) error
}
