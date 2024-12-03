// Package sdk
// 模块名: 模块名
// 功能描述: 描述
// 作者:  yr  2024/11/16 0016 17:38
// 最后更新:  yr  2024/11/16 0016 17:38
package sdk

import (
	"github.com/google/uuid"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/example/login/internal/data/db"
	"github.com/njtc406/chaosengine/example/login/internal/def"
	msg "github.com/njtc406/chaosengine/example/msg/login"
)

func init() {
	registerSDK(def.LoginTypeGuest, &Guest{})
}

type Guest struct{}

func (g *Guest) Login(rpcHandler inf.IRpcHandler, params interface{}) *db.User {
	if rpcHandler == nil {
		log.SysLogger.Debug(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	}
	req, ok := params.(*msg.C2S_SignIn)
	if !ok {
		log.SysLogger.Errorf("login failed, params is not *msg.Msg_CS_SignIn")
		return nil
	}
	user := new(db.User)
	var exists bool
	s := rpcHandler.SelectSameServer("", "DBService")
	if s == nil {
		log.SysLogger.Debugf("000000000000000000000000000000000000000000000000000000000")
	}
	if err := rpcHandler.SelectSameServer("", "DBService").
		Call("APIExecuteMysqlFun", []interface{}{user.Get, map[string]interface{}{"account": req.UserName}}, &exists); err != nil {
		log.SysLogger.Errorf("get user info failed, err: %s", err)
		return nil
	}

	if !exists {
		// 账号不存在,创建
		user.Channel = req.Channel
		user.ChildChannel = req.ChildChannel
		user.Account = req.UserName
		user.Secret = uuid.NewString()

		if err := rpcHandler.SelectSameServer("", "DBService").
			Call("APIExecuteMysqlFun", []interface{}{user.Insert}, nil); err != nil {
			log.SysLogger.Errorf("create user failed, err: %s", err)
			return nil
		}
	}

	return user
}

func (g *Guest) DeleteAccount(rpcHandler inf.IRpcHandler, req interface{}) error {
	return nil
}
