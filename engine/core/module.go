// Package core
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package core

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/event"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/concurrent"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/timer"
	"reflect"
	"sync/atomic"
	"time"
)

type Module struct {
	inf.IRpcHandler
	concurrent.IConcurrent

	moduleId     uint32 // 模块ID
	moduleName   string // 模块名称
	moduleIdSeed uint32 // 模块ID种子(如果没有给模块ID，则子模块从该种子开始分配)

	self         inf.IModule            // 自身
	parent       inf.IModule            // 父模块
	children     map[uint32]inf.IModule // 子模块列表 map[moduleId]module
	root         inf.IModule            // 根模块
	rootContains map[uint32]inf.IModule // 根模块下所有模块(包括所有的子模块)

	eventHandler inf.IEventHandler // 事件处理器

	timerDispatcher  *timer.Dispatcher         // 定时器调度器
	mapActiveTimer   map[timer.ITimer]struct{} // 活跃定时器
	mapActiveIDTimer map[uint64]timer.ITimer   // 活跃定时器id
}

func (m *Module) AddModule(module inf.IModule) (uint32, error) {
	if m.GetEventProcessor() == nil {
		return 0, errdef.ModuleNotInitialized
	}

	pModule := module.GetBaseModule().(*Module)

	if pModule.GetModuleID() == 0 {
		pModule.moduleId = m.newModuleID()
	}

	if m.children == nil {
		m.children = make(map[uint32]inf.IModule)
	}

	if _, ok := m.children[pModule.GetModuleID()]; ok {
		return 0, errdef.ModuleHadRegistered
	}

	pModule.IRpcHandler = m.IRpcHandler
	pModule.self = module
	pModule.parent = m.self
	pModule.timerDispatcher = m.GetRoot().GetBaseModule().(*Module).timerDispatcher
	pModule.root = m.root
	pModule.moduleName = reflect.Indirect(reflect.ValueOf(module)).Type().Name()
	pModule.eventHandler = event.NewHandler()
	pModule.eventHandler.Init(m.eventHandler.GetEventProcessor())
	pModule.IConcurrent = m.IConcurrent
	if err := module.OnInit(); err != nil {
		return 0, err
	}
	m.children[pModule.GetModuleID()] = module
	m.GetRoot().GetBaseModule().(*Module).rootContains[pModule.GetModuleID()] = module

	log.SysLogger.Debugf("add module [%s] completed", pModule.GetModuleName())

	return pModule.moduleId, nil
}

func (m *Module) OnInit() error {
	return nil
}

func (m *Module) OnRelease() {}

func (m *Module) newModuleID() uint32 {
	return atomic.AddUint32(&m.root.GetBaseModule().(*Module).moduleIdSeed, 1)
}

func (m *Module) NewModuleID() uint32 {
	return m.newModuleID()
}

func (m *Module) SetModuleID(id uint32) bool {
	if m.moduleId != 0 {
		return false
	}
	m.moduleId = id
	return true
}

func (m *Module) GetModuleID() uint32 {
	return m.moduleId
}

func (m *Module) GetModuleName() string {
	return m.moduleName
}

func (m *Module) GetModule(moduleId uint32) inf.IModule {
	iModule, ok := m.GetRoot().GetBaseModule().(*Module).rootContains[moduleId]
	if !ok {
		return nil
	}
	return iModule
}

func (m *Module) GetRoot() inf.IModule {
	return m.root
}

func (m *Module) GetParent() inf.IModule {
	return m.parent
}

func (m *Module) GetBaseModule() inf.IModule {
	return m
}

func (m *Module) GetService() inf.IService {
	return m.GetRoot().(inf.IService)
}

func (m *Module) GetEventProcessor() inf.IEventProcessor {
	return m.eventHandler.GetEventProcessor()
}

func (m *Module) GetEventHandler() inf.IEventHandler {
	return m.eventHandler
}

func (m *Module) ReleaseAllChildModule() {
	// 释放所有子模块
	for id := range m.children {
		m.ReleaseModule(id)
	}
}

func (m *Module) reset() {
	m.moduleId = 0
	m.moduleName = ""
	m.moduleIdSeed = 0
	m.self = nil
	m.parent = nil
	m.children = nil
	m.mapActiveTimer = nil
	m.timerDispatcher = nil
	m.root = nil
	m.rootContains = nil
	m.IRpcHandler = nil
	m.mapActiveIDTimer = nil
	m.eventHandler = nil
	m.IConcurrent = nil
}

func (m *Module) ReleaseModule(moduleId uint32) {
	pModule := m.GetModule(moduleId).GetBaseModule().(*Module)
	if pModule == nil {
		log.SysLogger.Errorf("module %d not found", moduleId)
		return
	}

	log.SysLogger.Debugf("release module %s ,id: %d name:%s", m.GetModuleName(), moduleId, pModule.GetModuleName())

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
	delete(m.GetRoot().GetBaseModule().(*Module).rootContains, moduleId)

	//清理被删除的Module
	pModule.reset()
}

func (m *Module) NotifyEvent(e inf.IEvent) {
	m.eventHandler.NotifyEvent(e)
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

func (m *Module) AfterFunc(d time.Duration, cb func(timer.ITimer)) timer.ITimer {
	if m.mapActiveTimer == nil {
		m.mapActiveTimer = map[timer.ITimer]struct{}{}
	}

	return m.timerDispatcher.AfterFunc(d, nil, cb, m.OnCloseTimer, m.OnAddTimer)
}

func (m *Module) CronFunc(cronExpr *timer.CronExpr, cb func(timer.ITimer)) timer.ITimer {
	if m.mapActiveTimer == nil {
		m.mapActiveTimer = map[timer.ITimer]struct{}{}
	}

	return m.timerDispatcher.CronFunc(cronExpr, nil, cb, m.OnCloseTimer, m.OnAddTimer)
}

func (m *Module) NewTicker(d time.Duration, cb func(timer.ITimer)) timer.ITimer {
	if m.mapActiveTimer == nil {
		m.mapActiveTimer = map[timer.ITimer]struct{}{}
	}

	return m.timerDispatcher.TickerFunc(d, nil, cb, m.OnCloseTimer, m.OnAddTimer)
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
	if timerId == nil {
		return
	}
	if m.mapActiveIDTimer == nil {
		m.mapActiveIDTimer = map[uint64]timer.ITimer{}
	}

	if *timerId != 0 {
		m.CancelTimerId(timerId)
	}

	*timerId = m.GenTimerId()
	t := m.timerDispatcher.AfterFunc(d, cb, nil, m.OnCloseTimer, m.OnAddTimer)
	t.AdditionData = AdditionData
	t.Id = *timerId
	m.mapActiveIDTimer[*timerId] = t
}

func (m *Module) SafeCronFunc(cronId *uint64, cronExpr *timer.CronExpr, AdditionData interface{}, cb func(uint64, interface{})) {
	if m.mapActiveIDTimer == nil {
		m.mapActiveIDTimer = map[uint64]timer.ITimer{}
	}

	*cronId = m.GenTimerId()
	c := m.timerDispatcher.CronFunc(cronExpr, cb, nil, m.OnCloseTimer, m.OnAddTimer)
	c.AdditionData = AdditionData
	c.Id = *cronId
	m.mapActiveIDTimer[*cronId] = c
}

func (m *Module) SafeNewTicker(tickerId *uint64, d time.Duration, AdditionData interface{}, cb func(uint64, interface{})) {
	if m.mapActiveIDTimer == nil {
		m.mapActiveIDTimer = map[uint64]timer.ITimer{}
	}

	*tickerId = m.GenTimerId()
	t := m.timerDispatcher.TickerFunc(d, cb, nil, m.OnCloseTimer, m.OnAddTimer)
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
