// Package inf
// @Title  title
// @Description  desc
// @Author  pc  2024/11/4
// @Update  pc  2024/11/4
package inf

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/config"
	"github.com/njtc406/chaosengine/engine/profiler"
	"github.com/njtc406/chaosengine/engine/utils/concurrent"
)

// IService 服务接口
// 每个服务就是一个单独的协程
type IService interface {
	concurrent.IConcurrent
	ILifecycle
	IIdentifiable
	IServiceHandler
	IEventChannel
	IProfiler
}

// ILifecycle 服务生命周期
type ILifecycle interface {
	Init(src interface{}, serviceInitConf *config.ServiceInitConf, cfg interface{})
	Start() error
	Stop()
	OnInit() error
	OnStart() error
	OnRelease()
}

type IServiceHandler interface {
	SetGoRoutineNum(num int32)
	GetServiceCfg() interface{}
	GetServiceEventChannelNum() int
	GetServiceTimerChannelNum() int
	GetRpcHandler() IRpcHandler
}

type IIdentifiable interface {
	OnSetup(svc IService)
	SetName(string)
	GetName() string
	SetServiceId(id string)
	GetServiceId() string
	GetPid() *actor.PID
	GetServerId() int32
	IsClosed() bool // 服务是否已经关闭
}

type IProfiler interface {
	OpenProfiler()
	GetProfiler() *profiler.Profiler
}
