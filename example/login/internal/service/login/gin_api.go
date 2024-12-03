// Package login
// 模块名: 模块名
// 功能描述: 描述
// 作者:  yr  2024/11/16 0016 18:07
// 最后更新:  yr  2024/11/16 0016 18:07
package login

import (
	"context"
	"github.com/gin-gonic/gin"
	sysInf "github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/timelib"
	"github.com/njtc406/chaosengine/example/login/event"
	"github.com/njtc406/chaosengine/example/login/internal/def"
	"github.com/njtc406/chaosengine/example/login/internal/dto"
	commmsg "github.com/njtc406/chaosengine/example/msg/comm"
	msg "github.com/njtc406/chaosengine/example/msg/login"
	"github.com/njtc406/chaosengine/example/utils/tokenlib"
)

func (l *LoginService) OnAuth(c *gin.Context) {
	startTime := timelib.GetTime()
	resp := dto.NewResponse()
	var req *msg.C2S_SignIn
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.SetCode(commmsg.ErrCode_ParamsInvalid).DoResponse(c)
		return
	}

	ev := event.NewHttpEvent()
	ev.Type = sysInf.EventType(msg.MsgId_C2S_SignIn)
	ev.Data = req
	ev.Resp = resp

	// 放入login队列中执行,防止并发
	if err := l.PushHttpEvent(ev); err != nil {
		resp.SetCode(commmsg.ErrCode_ApiError).DoResponse(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), def.DefaultHttpReqTimeout*3)
	defer cancel()

	select {
	case <-ctx.Done():
		resp.SetCode(commmsg.ErrCode_ApiRunTimeout)
	case <-ev.Wait():
		break
	}

	log.SysLogger.Debugf("resp: %+v", resp)
	event.ReleaseHttpEvent(ev)
	log.SysLogger.Debugf(">>>>>>>>>>>>>>>>>>>>>>>>tm: %+v", timelib.Since(startTime))
	resp.DoResponse(c)
}

func (l *LoginService) OnServerList(c *gin.Context) {
	resp := dto.NewResponse()
	var req *msg.C2S_ServerList
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.SetCode(commmsg.ErrCode_ParamsInvalid).DoResponse(c)
		return
	}

	// token验证
	if claims, err := tokenlib.ParseJwtToken(req.Token); err != nil {
		resp.SetCode(commmsg.ErrCode_TokenInvalid)
	} else if claims.ExpireTime <= timelib.GetTimeUnix() {
		resp.SetCode(commmsg.ErrCode_TokenExpired)
	}

	// TODO 直接请求gm后台,这个数据可以并发

	resp.SetData(&msg.S2C_ServerList{
		List: []*msg.ServerInfo{
			{
				Id:        1001,
				Name:      "测试1服",
				StartTime: 0,
				Status:    commmsg.ServerStatus_Normal,
			},
		},
	}).DoResponse(c)

	//ev := event.NewHttpEvent()
	//ev.Type = sysInf.EventType(msg.MsgId_C2S_ServerList)
	//ev.Data = req
	//ev.Resp = resp
	//
	//// 放入login队列中执行,防止并发
	//if err := l.PushHttpEvent(ev); err != nil {
	//	resp.SetCode(commmsg.ErrCode_ApiError).DoResponse(c)
	//}
	//
	//ctx, cancel := context.WithTimeout(context.Background(), def.DefaultHttpReqTimeout)
	//defer cancel()
	//
	//select {
	//case <-ctx.Done():
	//	resp.SetCode(commmsg.ErrCode_ApiRunTimeout)
	//case <-ev.Wait():
	//	break
	//}
	//event.ReleaseHttpEvent(ev)
	//resp.DoResponse(c)
}

func (l *LoginService) OnAnnouncementList(c *gin.Context) {
	resp := dto.NewResponse()
	var req *msg.C2S_AnnouncementList
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.SetCode(commmsg.ErrCode_ParamsInvalid).DoResponse(c)
		return
	}

	// token验证
	if claims, err := tokenlib.ParseJwtToken(req.Token); err != nil {
		resp.SetCode(commmsg.ErrCode_TokenInvalid)
	} else if claims.ExpireTime <= timelib.GetTimeUnix() {
		resp.SetCode(commmsg.ErrCode_TokenExpired)
	}

	// TODO 直接请求gm后台,这个数据可以并发
	resp.SetData(&msg.S2C_AnnouncementList{
		List: []*msg.AnnouncementInfo{
			{
				Title:      "测试数据标题",
				Content:    "测试数据内容",
				Group:      "测试",
				ShowTime:   0,
				ExpireTime: 0,
				Sort:       1,
				Url:        "",
			},
		},
	}).DoResponse(c)

	//ev := event.NewHttpEvent()
	//ev.Type = sysInf.EventType(msg.MsgId_C2S_AnnouncementList)
	//ev.Data = req
	//ev.Resp = resp
	//
	//// 放入login队列中执行,防止并发
	//if err := l.PushHttpEvent(ev); err != nil {
	//	resp.SetCode(commmsg.ErrCode_ApiError).DoResponse(c)
	//}
	//
	//ctx, cancel := context.WithTimeout(context.Background(), def.DefaultHttpReqTimeout)
	//defer cancel()
	//
	//select {
	//case <-ctx.Done():
	//	resp.SetCode(commmsg.ErrCode_ApiRunTimeout)
	//case <-ev.Wait():
	//	break
	//}
	//event.ReleaseHttpEvent(ev)
	//resp.DoResponse(c)
}

func (l *LoginService) OnSelectServer(c *gin.Context) {
	resp := dto.NewResponse()
	var req *msg.C2S_SelectServer
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.SetCode(commmsg.ErrCode_ParamsInvalid).DoResponse(c)
		return
	}

	// token验证
	if claims, err := tokenlib.ParseJwtToken(req.Token); err != nil {
		resp.SetCode(commmsg.ErrCode_TokenInvalid)
	} else if claims.ExpireTime <= timelib.GetTimeUnix() {
		resp.SetCode(commmsg.ErrCode_TokenExpired)
	}

	// TODO 从服务发现中根据规则找到一个gateService的地址,并且将登录token发送到对应的gateService中缓存,最大缓存时间需要配置
	resp.SetData(&msg.S2C_SelectServer{
		Address: "127.0.0.1:8901",
	}).DoResponse(c)

	//ev := event.NewHttpEvent()
	//ev.Type = sysInf.EventType(msg.MsgId_C2S_SelectServer)
	//ev.Data = req
	//ev.Resp = resp
	//
	//// 放入login队列中执行,防止并发
	//if err := l.PushHttpEvent(ev); err != nil {
	//	resp.SetCode(commmsg.ErrCode_ApiError).DoResponse(c)
	//}
	//
	//ctx, cancel := context.WithTimeout(context.Background(), def.DefaultHttpReqTimeout)
	//defer cancel()
	//
	//select {
	//case <-ctx.Done():
	//	resp.SetCode(commmsg.ErrCode_ApiRunTimeout)
	//case <-ev.Wait():
	//	break
	//}
	//event.ReleaseHttpEvent(ev)
	//resp.DoResponse(c)
}
