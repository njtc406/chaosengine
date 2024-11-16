// Package inf
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package inf

// EventCallBack 事件接受器
type EventCallBack func(event IEvent)

type EventType int

type IEvent interface {
	GetType() EventType
}

type IChannel interface {
	PushEvent(ev IEvent) error
}

type IProcessor interface {
	IChannel

	Init(eventChannel IChannel)
	EventHandler(ev IEvent)
	RegEventReceiverFunc(eventType EventType, receiver IHandler, callback EventCallBack)
	UnRegEventReceiverFun(eventType EventType, receiver IHandler)

	CastEvent(event IEvent) //广播事件
	AddBindEvent(eventType EventType, receiver IHandler, callback EventCallBack)
	AddListen(eventType EventType, receiver IHandler)
	RemoveBindEvent(eventType EventType, receiver IHandler)
	RemoveListen(eventType EventType, receiver IHandler)
}

type IHandler interface {
	Init(p IProcessor)
	GetEventProcessor() IProcessor
	NotifyEvent(IEvent)
	Destroy()
	//注册了事件
	AddRegInfo(eventType EventType, eventProcessor IProcessor)
	RemoveRegInfo(eventType EventType, eventProcessor IProcessor)
}

type IEventHandler interface {
	Init(processor IEventProcessor)
	GetEventProcessor() IEventProcessor //获得事件
	NotifyEvent(IEvent)
	Destroy()
	//注册了事件
	AddRegInfo(eventType EventType, eventProcessor IEventProcessor)
	RemoveRegInfo(eventType EventType, eventProcessor IEventProcessor)
}

type IEventChannel interface {
	PushEvent(ev IEvent) error
}

type IEventProcessor interface {
	IEventChannel

	Init(eventChannel IEventChannel)
	EventHandler(ev IEvent)
	RegEventReceiverFunc(eventType EventType, receiver IEventHandler, callback EventCallBack)
	UnRegEventReceiverFun(eventType EventType, receiver IEventHandler)

	CastEvent(event IEvent) //广播事件
	AddBindEvent(eventType EventType, receiver IEventHandler, callback EventCallBack)
	AddListen(eventType EventType, receiver IEventHandler)
	RemoveBindEvent(eventType EventType, receiver IEventHandler)
	RemoveListen(eventType EventType, receiver IEventHandler)
}
