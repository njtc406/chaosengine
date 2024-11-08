// Package service
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/7/22 上午9:40
// @Update  yr  2024/7/22 上午9:40
package service

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine1/concurrent"
	"github.com/njtc406/chaosengine/engine1/define/inf"
	"github.com/njtc406/chaosengine/engine1/errdef"
	"github.com/njtc406/chaosengine/engine1/event"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosengine/engine1/timer"
	"reflect"
	"sync/atomic"
	"time"
)

type Module struct {
	inf.IRpcHandler
	moduleID               uint32                    // 模块ID
	moduleName             string                    // 模块名称
	parent                 inf.IModule               // 父类
	self                   inf.IModule               // 自身
	children               map[uint32]inf.IModule    // 子模块
	mapChildrenName        map[string]inf.IModule    // 子模块
	dispatcher             *timer.Dispatcher         // 定时器
	mapActiveTimer         map[timer.ITimer]struct{} // 活跃定时器
	mapActiveIDTimer       map[uint64]timer.ITimer   // 活跃定时器
	ancestor               inf.IModule               // 根模块
	seedModuleId           uint32                    //模块id种子
	descendants            map[uint32]inf.IModule    // 根模块的子模块
	eventHandler           event.IHandler            // 事件处理器
	concurrent.IConcurrent                           // 并发接口(内部可以使用多线程调用, 然后回调回来之后依然是单线程执行)
}

func (m *Module) SetModuleID(id uint32) bool {
	if m.moduleID != 0 {
		return false
	}

	m.moduleID = id
	return true
}

func (m *Module) GetModuleID() uint32 {
	return m.moduleID
}

func (m *Module) GetModuleName() string {
	return m.moduleName
}

func (m *Module) OnInit() error { return nil }
func (m *Module) OnStart()      {}
func (m *Module) OnRelease()    {}

func (m *Module) AddModule(module inf.IModule) (uint32, error) {
	// 没有事件处理器不允许加入子模块
	if m.GetEventProcessor() == nil {
		return 0, errdef.ModuleNotInitialized
	}

	pModule := module.GetBaseModule().(*Module)
	if pModule.GetModuleID() == 0 {
		pModule.moduleID = m.newModuleID()
	}

	if m.children == nil {
		m.children = make(map[uint32]inf.IModule)
		m.mapChildrenName = make(map[string]inf.IModule)
	}

	if _, ok := m.children[pModule.GetModuleID()]; ok {
		return 0, errdef.ModuleHadRegistered
	}

	pModule.IRpcHandler = m.IRpcHandler
	pModule.self = module
	pModule.parent = m.self
	pModule.dispatcher = m.GetAncestor().GetBaseModule().(*Module).dispatcher
	pModule.ancestor = m.ancestor
	pModule.moduleName = reflect.Indirect(reflect.ValueOf(module)).Type().Name()
	pModule.eventHandler = event.NewHandler()
	pModule.eventHandler.Init(m.eventHandler.GetEventProcessor())
	pModule.IConcurrent = m.IConcurrent
	if err := module.OnInit(); err != nil {
		return 0, err
	}
	m.children[pModule.GetModuleID()] = module
	m.mapChildrenName[pModule.GetModuleName()] = module
	m.ancestor.GetBaseModule().(*Module).descendants[pModule.GetModuleID()] = module

	log.SysLogger.Debugf("add module [%s] completed", pModule.GetModuleName())

	return pModule.GetModuleID(), nil
}

// ReleaseAllChildModule 释放所有子模块
func (m *Module) ReleaseAllChildModule() {
	for _, module := range m.children {
		module.ReleaseModule(module.GetModuleID())
	}
}

func (m *Module) ReleaseModule(moduleId uint32) {
	log.SysLogger.Debugf("release module %s ,id: %d", m.GetModuleName(), moduleId)
	pModule := m.GetModule(moduleId).GetBaseModule().(*Module)

	//释放子孙
	for id := range pModule.children {
		m.ReleaseModule(id)
	}

	pModule.self.OnRelease()
	pModule.GetEventHandler().Destroy()
	log.SysLogger.Debugf("Release module %s", pModule.GetModuleName())
	for pTimer := range pModule.mapActiveTimer {
		pTimer.Cancel()
	}
	for _, t := range pModule.mapActiveIDTimer {
		t.Cancel()
	}
	delete(m.children, moduleId)
	delete(m.ancestor.GetBaseModule().(*Module).descendants, moduleId)

	//清理被删除的Module
	pModule.self = nil
	pModule.parent = nil
	pModule.children = nil
	pModule.mapActiveTimer = nil
	pModule.dispatcher = nil
	pModule.ancestor = nil
	pModule.descendants = nil
	pModule.IRpcHandler = nil
	pModule.mapActiveIDTimer = nil
	pModule.mapChildrenName = nil
	pModule.eventHandler = nil
	pModule.IConcurrent = nil
	pModule.moduleID = 0
	pModule.moduleName = ""
	pModule.parent = nil
}

func (m *Module) GetModule(moduleId uint32) inf.IModule {
	iModule, ok := m.GetAncestor().GetBaseModule().(*Module).descendants[moduleId]
	if ok == false {
		return nil
	}
	return iModule
}

func (m *Module) GetAncestor() inf.IModule {
	return m.ancestor
}

func (m *Module) GetBaseModule() inf.IModule {
	return m
}

func (m *Module) GetModuleByName(name string) inf.IModule {
	return m.mapChildrenName[name]
}

func (m *Module) GetParent() inf.IModule {
	return m.parent
}

func (m *Module) GetService() inf.IService {
	return m.GetAncestor().(inf.IService)
}

func (m *Module) GetEventProcessor() event.IProcessor {
	return m.eventHandler.GetEventProcessor()
}

func (m *Module) newModuleID() uint32 {
	m.ancestor.GetBaseModule().(*Module).seedModuleId += 1
	return m.ancestor.GetBaseModule().(*Module).seedModuleId
}

func (m *Module) NewModuleID() uint32 {
	return m.newModuleID()
}

func (m *Module) NotifyEvent(ev event.IEvent) {
	m.eventHandler.NotifyEvent(ev)
}

func (m *Module) GetEventHandler() event.IHandler {
	return m.eventHandler
}

func (m *Module) OnCloseTimer(t timer.ITimer) {
	delete(m.mapActiveIDTimer, t.GetId())
	delete(m.mapActiveTimer, t)
}

func (m *Module) OnAddTimer(t timer.ITimer) {
	if t != nil {
		if m.mapActiveTimer == nil {
			m.mapActiveTimer = map[timer.ITimer]struct{}{}
		}

		m.mapActiveTimer[t] = struct{}{}
	}
}

func (m *Module) AfterFunc(d time.Duration, cb func(*timer.Timer)) *timer.Timer {
	if m.mapActiveTimer == nil {
		m.mapActiveTimer = map[timer.ITimer]struct{}{}
	}

	return m.dispatcher.AfterFunc(d, nil, cb, m.OnCloseTimer, m.OnAddTimer)
}

func (m *Module) CronFunc(cronExpr *timer.CronExpr, cb func(*timer.Cron)) *timer.Cron {
	if m.mapActiveTimer == nil {
		m.mapActiveTimer = map[timer.ITimer]struct{}{}
	}

	return m.dispatcher.CronFunc(cronExpr, nil, cb, m.OnCloseTimer, m.OnAddTimer)
}

func (m *Module) NewTicker(d time.Duration, cb func(*timer.Ticker)) *timer.Ticker {
	if m.mapActiveTimer == nil {
		m.mapActiveTimer = map[timer.ITimer]struct{}{}
	}

	return m.dispatcher.TickerFunc(d, nil, cb, m.OnCloseTimer, m.OnAddTimer)
}

func (m *Module) cb(*timer.Timer) {

}

var timerSeedId uint32

func (m *Module) GenTimerId() uint64 {
	for {
		newTimerId := (uint64(m.GetModuleID()) << 32) | uint64(atomic.AddUint32(&timerSeedId, 1))
		if _, ok := m.mapActiveIDTimer[newTimerId]; ok == true {
			continue
		}

		return newTimerId
	}
}

func (m *Module) SafeAfterFunc(timerId *uint64, d time.Duration, AdditionData interface{}, cb func(uint64, interface{})) {
	if m.mapActiveIDTimer == nil {
		m.mapActiveIDTimer = map[uint64]timer.ITimer{}
	}

	if *timerId != 0 {
		m.CancelTimerId(timerId)
	}

	*timerId = m.GenTimerId()
	t := m.dispatcher.AfterFunc(d, cb, nil, m.OnCloseTimer, m.OnAddTimer)
	t.AdditionData = AdditionData
	t.Id = *timerId
	m.mapActiveIDTimer[*timerId] = t
}

func (m *Module) SafeCronFunc(cronId *uint64, cronExpr *timer.CronExpr, AdditionData interface{}, cb func(uint64, interface{})) {
	if m.mapActiveIDTimer == nil {
		m.mapActiveIDTimer = map[uint64]timer.ITimer{}
	}

	*cronId = m.GenTimerId()
	c := m.dispatcher.CronFunc(cronExpr, cb, nil, m.OnCloseTimer, m.OnAddTimer)
	c.AdditionData = AdditionData
	c.Id = *cronId
	m.mapActiveIDTimer[*cronId] = c
}

func (m *Module) SafeNewTicker(tickerId *uint64, d time.Duration, AdditionData interface{}, cb func(uint64, interface{})) {
	if m.mapActiveIDTimer == nil {
		m.mapActiveIDTimer = map[uint64]timer.ITimer{}
	}

	*tickerId = m.GenTimerId()
	t := m.dispatcher.TickerFunc(d, cb, nil, m.OnCloseTimer, m.OnAddTimer)
	t.AdditionData = AdditionData
	t.Id = *tickerId
	m.mapActiveIDTimer[*tickerId] = t
}

func (m *Module) CancelTimerId(timerId *uint64) bool {
	if timerId == nil || *timerId == 0 {
		log.SysLogger.Warn("timerId is invalid")
		return false
	}

	if m.mapActiveIDTimer == nil {
		log.SysLogger.Error("mapActiveIdTimer is nil")
		return false
	}

	t, ok := m.mapActiveIDTimer[*timerId]
	if ok == false {
		log.SysLogger.Debugf(fmt.Sprintf("cannot find timer id %d", *timerId))
		return false
	}

	t.Cancel()
	*timerId = 0
	return true
}
