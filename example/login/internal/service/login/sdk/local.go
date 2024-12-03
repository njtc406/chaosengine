// Package sdk
// 模块名: 模块名
// 功能描述: 描述
// 作者:  yr  2024/11/16 0016 17:38
// 最后更新:  yr  2024/11/16 0016 17:38
package sdk

import "github.com/njtc406/chaosengine/example/login/internal/data/db"

type Local struct{}

func (l *Local) Login(req interface{}) *db.User {
	return nil
}
