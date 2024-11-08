// Package inf
// @Title  title
// @Description  desc
// @Author  pc  2024/11/4
// @Update  pc  2024/11/4
package inf

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/utils/concurrent"
)

// IService 服务接口
// 每个服务就是一个单独的协程
type IService interface {
	concurrent.IConcurrent
	ILifecycle
	IIdentifiable
	IServiceHandler
	IChannel
	IRpcService
	//IProfiler
}

// ILifecycle 服务生命周期
type ILifecycle interface {
	Init(src interface{}, cfg interface{})
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
}

type IIdentifiable interface {
	OnSetup(svc IService)
	SetName(string)
	GetName() string
	SetUid(uid string)
	GetUid() string
	GetPID() *actor.PID
}

type IProfiler interface {
	OpenProfiler()
	//GetProfiler() *profiler.Profiler
}

type IRpcService interface {
	//GetRpcHandler() IRpcHandler
}
