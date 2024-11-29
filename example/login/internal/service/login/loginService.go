// Package login
// 模块名: 模块名
// 功能描述: 描述
// 作者:  yr  2024/11/16 0016 16:03
// 最后更新:  yr  2024/11/16 0016 16:03
package login

import (
	"github.com/gin-gonic/gin"
	sysConfig "github.com/njtc406/chaosengine/engine/config"
	"github.com/njtc406/chaosengine/engine/core"
	"github.com/njtc406/chaosengine/engine/event"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/services"
	"github.com/njtc406/chaosengine/engine/sysModule/httpmodule"
	"github.com/njtc406/chaosengine/engine/sysModule/httpmodule/router_center"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/example/login/config"
	"github.com/njtc406/chaosengine/example/login/internal/def"
)

func init() {
	services.SetService("LoginService", func() inf.IService { return &LoginService{} })
	sysConfig.RegisterServiceConf(&sysConfig.ServiceConfig{
		ServiceName: "LoginService",
		ConfName:    "login",
		ConfPath:    "",
		ConfType:    "yaml",
		CfgCreator: func() interface{} {
			return &config.LoginServiceConf{}
		},
		Cfg: &config.LoginServiceConf{},
	})
}

type LoginService struct {
	core.Service

	httpModule *httpmodule.HttpModule
}

func (l *LoginService) OnInit() error {
	log.SysLogger.Infof("login service init")
	l.httpModule = httpmodule.NewHttpModule(l.GetServiceCfg().(*config.LoginServiceConf).LoginServerConf, log.SysLogger, sysConfig.GetStatus())
	l.httpModule.SetRouter(l.initRouter())
	l.httpModule.SetModuleID(def.LoginHttpModuleId)
	l.AddModule(l.httpModule)

	l.GetEventProcessor().RegEventReceiverFunc(event.SysEventHttpMsg, l.GetEventHandler(), l.httpHandler)
	return nil
}

func (l *LoginService) OnStart() error {
	return l.httpModule.OnStart()
}

func (l *LoginService) OnRelease() {
	l.httpModule.OnRelease()
	l.ReleaseModule(def.LoginHttpModuleId)
}

func (l *LoginService) initRouter() *router_center.GroupHandlerPool {
	router := router_center.NewGroupHandlerPool()
	router.RegisterGroupHandler(router_center.DefaultGroup, l.routerHandlerV1)
	return router
}

func (l *LoginService) routerHandlerV1(rg *gin.RouterGroup) {
	rg.POST("/auth", l.OnAuth)
	rg.POST("/serverList", l.OnServerList)
	rg.POST("/announcementList", l.OnAnnouncementList)
	rg.POST("/selectServer", l.OnSelectServer)
}
