// Package inf
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package inf

// EventCallBack 事件接受器
type EventCallBack func(event IEvent)

type EventType int32

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

type IEventChannel interface {
	PushEvent(ev IEvent) error
}
