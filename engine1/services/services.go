// Package services
// @Title  系统服务
// @Description  这里放的都是系统级服务
// @Author  yr  2024/7/22 下午2:30
// @Update  yr  2024/7/22 下午2:30
package services

import (
	"github.com/njtc406/chaosengine/engine1/define/inf"
	"github.com/njtc406/chaosengine/engine1/log"
	"sync"
)

var (
	lock           = sync.RWMutex{}
	mapServiceName map[string]inf.IService
	list           []inf.IService
)

func init() {
	mapServiceName = make(map[string]inf.IService)
}

// Setup 注册服务
func Setup(services ...inf.IService) {
	for _, svc := range services {
		svc.OnSetup(svc)
		mapServiceName[svc.GetName()] = svc
		list = append(list, svc)
	}
}

func Init() {
	lock.RLock()
	defer lock.RUnlock()
	for _, svc := range list {
		if err := svc.OnInit(); err != nil {
			log.SysLogger.Panicf("Service %s init error: %v", svc.GetName(), err)
		}
	}
}

func Start() {
	lock.RLock()
	defer lock.RUnlock()
	for _, svc := range list {
		log.SysLogger.Infof("Start Service: %s", svc.GetName())
		svc.Start()
	}
}

func StopAll() {
	lock.RLock()
	defer lock.RUnlock()
	for i := len(list) - 1; i >= 0; i-- {
		log.SysLogger.Infof("Stop Service: %s", list[i].GetName())
		list[i].Stop()
	}
}
