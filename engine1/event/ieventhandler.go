// Package event
// @Title  事件处理器接口
// @Description  事件处理器接口
// @Author  yr  2024/7/19 下午3:19
// @Update  yr  2024/7/19 下午3:19
package event

type IHandler interface {
	Init(p IProcessor)
	GetEventProcessor() IProcessor
	NotifyEvent(IEvent)
	Destroy()
	//注册了事件
	addRegInfo(eventType int, eventProcessor IProcessor)
	removeRegInfo(eventType int, eventProcessor IProcessor)
}

type IChannel interface {
	PushEvent(ev IEvent) error
}
