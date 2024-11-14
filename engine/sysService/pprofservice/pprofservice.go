// Package pprofservice
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/8/21 下午5:00
// @Update  yr  2024/8/21 下午5:00
package pprofservice

import (
	"github.com/gin-gonic/gin"
	"github.com/njtc406/chaosengine/engine/core"
	"github.com/njtc406/chaosengine/engine/node"
	nodeConfig "github.com/njtc406/chaosengine/engine/node/config"
	"github.com/njtc406/chaosengine/engine/sysModule/httpmodule"
	"github.com/njtc406/chaosengine/engine/sysModule/httpmodule/router_center"
	"github.com/njtc406/chaosengine/engine/sysService/pprofservice/config"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"net/http"
	_ "net/http/pprof"
)

func init() {
	node.SetupBase(&PprofService{})
}

type PprofService struct {
	core.Service

	httpModule *httpmodule.HttpModule
}

func (ps *PprofService) getConf() *config.PprofConf {
	return ps.GetServiceCfg().(*config.PprofConf)
}

func (ps *PprofService) OnInit() error {
	ps.httpModule = httpmodule.NewHttpModule(ps.getConf().Conf, log.SysLogger, nodeConfig.Conf.SystemStatus)
	ps.httpModule.SetRouter(ps.initRouter())
	_, err := ps.AddModule(ps.httpModule)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PprofService) initRouter() *router_center.GroupHandlerPool {
	router := router_center.NewGroupHandlerPool()
	router.RegisterGroupHandler("", ps.routerHandler)
	return router
}

func (ps *PprofService) routerHandler(r *gin.RouterGroup) {
	r.GET("/debug/pprof", gin.WrapH(http.DefaultServeMux))
	r.GET("/debug/pprof/*pprof", gin.WrapH(http.DefaultServeMux))
}

func (ps *PprofService) OnStart() error {
	if err := ps.httpModule.OnStart(); err != nil {
		log.SysLogger.Panic(err)
	}
	return nil
}

func (ps *PprofService) OnRelease() {
	ps.httpModule.OnRelease()
}
