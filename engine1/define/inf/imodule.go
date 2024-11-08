// Package service
// @Title  模块
// @Description  模块
// @Author  yr  2024/7/22 上午10:08
// @Update  yr  2024/7/22 上午10:08
package inf

import (
	"github.com/njtc406/chaosengine/engine1/concurrent"
	"github.com/njtc406/chaosengine/engine1/event"
	"github.com/njtc406/chaosengine/engine1/timer"
	"time"
)

type IModuleHandler interface {
	OnInit() error // 初始化
	OnRelease()    // 释放
}

type IModule interface {
	IModuleHandler
	concurrent.IConcurrent

	SetModuleID(uint32) bool           // 设置模块ID
	GetModuleID() uint32               // 获取模块ID
	GetModuleName() string             // 获取模块名称
	NewModuleID() uint32               // 生成模块ID
	AddModule(IModule) (uint32, error) // 添加子模块
	ReleaseAllChildModule()            // 释放所有子模块
	ReleaseModule(moduleId uint32)     // 释放指定模块
	GetModule(uint32) IModule          // 获取指定模块
	GetModuleByName(string) IModule    // 通过模块名获取指定模块
	GetAncestor() IModule              // 获取祖先模块(即service自带的那个模块)

	GetBaseModule() IModule // 获取基础模块
	GetParent() IModule     // 获取父模块
	GetService() IService   // 获取服务

	GetEventProcessor() event.IProcessor // 获取事件处理器
	NotifyEvent(event.IEvent)            // 通知事件
}

type IModuleTimer interface {
	AfterFunc(d time.Duration, cb func(*timer.Timer)) *timer.Timer
	CronFunc(cronExpr *timer.CronExpr, cb func(*timer.Cron)) *timer.Cron
	NewTicker(d time.Duration, cb func(*timer.Ticker)) *timer.Ticker
}
