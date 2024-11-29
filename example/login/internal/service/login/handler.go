// Package login
// @Title  title
// @Description  desc
// @Author  yr  2024/11/20
// @Update  yr  2024/11/20
package login

import (
	sysEvent "github.com/njtc406/chaosengine/engine/event"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/timelib"
	"github.com/njtc406/chaosengine/example/login/internal/service/login/sdk"
	"reflect"

	commmsg "server/msg/comm"
	msg "server/msg/login"
)

var handlerMap = map[msg.MsgId]func(l *LoginService, ev *event.HttpEvent){
	msg.MsgId_C2S_SignIn:        Msg_CS_SignIn,        // 登录校验
	msg.MsgId_C2S_DeleteAccount: Msg_CS_DeleteAccount, // 删除账号
}

func (l *LoginService) httpHandler(e inf.IEvent) {
	ev, ok := e.(*sysEvent.Event)
	if !ok {
		log.SysLogger.Errorf("httpHandler event type error: got %v", reflect.TypeOf(e))
		return
	}

	req, ok := ev.Data.(*event.HttpEvent)
	if !ok {
		log.SysLogger.Errorf("httpHandler event data type error: got %v", reflect.TypeOf(ev.Data))
		return
	}

	handle, ok := handlerMap[msg.MsgId(req.GetType())]
	if ok {
		handle(l, req)
	} else {
		req.Resp.SetCode(commmsg.ErrCode_ApiNotFound)
	}
	req.Done()
}

// Msg_CS_SignIn 登录校验
func Msg_CS_SignIn(l *LoginService, ev *event.HttpEvent) {
	resp := ev.Resp
	req, ok := ev.Data.(*msg.C2S_SignIn)
	if !ok {
		resp.SetCode(commmsg.ErrCode_ParamsInvalid)
		return
	}
	sdkHandler := sdk.GetSDKHandler(req.Type)
	if sdkHandler == nil {
		resp.SetCode(commmsg.ErrCode_SdkNotFound)
		return
	}
	user := sdkHandler.Login(l.GetRpcHandler(), req)
	if user == nil {
		resp.SetCode(commmsg.ErrCode_UserNotFound)
		return
	}

	// 检查用户的删除标记
	if user.DeleteTime > 0 {
		if user.DeleteTime > timelib.GetTimeUnix() {
			// 删除标记未过期,则删除标记
			user.DeleteTime = 0
			// 保存数据
			if err := l.Select(l.GetPID().GetServerId(), "", "DBService").Send("APIExecuteMysqlFun", []interface{}{user.Update}); err != nil {
				log.SysLogger.Errorf("Msg_CS_SignIn APIExecuteMysqlFun error: %v", err)
			}
		} else {
			// 删除标记已过期,删除用户
			if err := l.Select(l.GetPID().GetServerId(), "", "DBService").Send("APIExecuteMysqlFun", []interface{}{user.Delete}); err != nil {
				log.SysLogger.Errorf("Msg_CS_SignIn APIExecuteMysqlFun error: %v", err)
				resp.SetCode(commmsg.ErrCode_UserStatusAbnormal)
				return
			}

			// 重新登录
			user = sdkHandler.Login(l.GetRpcHandler(), req)
			if user == nil {
				resp.SetCode(commmsg.ErrCode_UserNotFound)
				return
			}
		}
	}

	// 生成token
	token, err := tokenlib.CreateJwtToken(user.ID)
	if err != nil {
		log.SysLogger.Errorf("Msg_CS_SignIn CreateJwtToken error: %v", err)
		resp.SetCode(commmsg.ErrCode_TokenCreateFailed)
		return
	}

	resp.SetData(&msg.S2C_SignIn{
		Token: token,
	})
}

func Msg_CS_DeleteAccount(l *LoginService, ev *event.HttpEvent) {
	resp := ev.Resp
	//req, ok := ev.Data.(*msg.Msg_CS_DeleteAccount)
	//if !ok {
	//	resp.SetCode(msg.ErrCode_ParamsInvalid)
	//	return
	//}

	resp.SetCode(commmsg.ErrCode_ApiDeveloping)
}
