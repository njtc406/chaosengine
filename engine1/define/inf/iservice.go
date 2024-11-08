// Package inf
// @Title  服务接口
// @Description  服务接口
// @Author  yr  2024/7/19 上午10:42
// @Update  yr  2024/7/19 上午10:42
package inf

import (
	"github.com/njtc406/chaosengine/engine1/actor"
	"github.com/njtc406/chaosengine/engine1/concurrent"
	"github.com/njtc406/chaosengine/engine1/event"
	"github.com/njtc406/chaosengine/engine1/profiler"
)

type IService interface {
	concurrent.IConcurrent
	event.IChannel

	Init(svc interface{}, cfg interface{})
	Start()
	Stop()

	OnSetup(svc IService)
	OnInit() error
	OnStart()
	OnRelease()

	SetName(string)
	GetName() string
	SetID(id string)
	GetID() string

	GetPID() *actor.PID
	GetRpcHandler() IRpcHandler
	GetServiceCfg() interface{}
	OpenProfiler()
	GetProfiler() *profiler.Profiler
	GetServiceEventChannelNum() int
	GetServiceTimerChannelNum() int
}
