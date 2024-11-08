// Package event
// @Title  事件管理器
// @Description  这里管理着所有已经注册的事件,一般是一个service一个processor,事件触发时分发到不同的handler，并执行回调
// @Author  yr  2024/7/19 下午3:33
// @Update  yr  2024/7/19 下午3:33
package event

import (
	"sync"
)

// CallBack 事件接受器
type CallBack func(event IEvent)

type IProcessor interface {
	IChannel

	Init(eventChannel IChannel)
	EventHandler(ev IEvent)
	RegEventReceiverFunc(eventType int, receiver IHandler, callback CallBack)
	UnRegEventReceiverFun(eventType int, receiver IHandler)

	castEvent(event IEvent) //广播事件
	addBindEvent(eventType int, receiver IHandler, callback CallBack)
	addListen(eventType int, receiver IHandler)
	removeBindEvent(eventType int, receiver IHandler)
	removeListen(eventType int, receiver IHandler)
}

type Processor struct {
	IChannel

	locker              sync.RWMutex
	mapListenerEvent    map[int]map[IProcessor]int    //监听者信息
	mapBindHandlerEvent map[int]map[IHandler]CallBack //收到事件处理
}

func NewProcessor() IProcessor {
	p := &Processor{
		mapListenerEvent:    make(map[int]map[IProcessor]int),
		mapBindHandlerEvent: make(map[int]map[IHandler]CallBack),
	}
	return p
}

func (p *Processor) Init(eventChannel IChannel) {
	p.IChannel = eventChannel
}

// EventHandler 事件处理
func (p *Processor) EventHandler(ev IEvent) {
	eventType := ev.GetEventType()
	mapCallBack, ok := p.mapBindHandlerEvent[eventType]
	if !ok {
		return
	}
	for _, callback := range mapCallBack {
		callback(ev)
	}
}

// RegEventReceiverFunc 注册事件处理函数
func (p *Processor) RegEventReceiverFunc(eventType int, receiver IHandler, callback CallBack) {
	//记录receiver自己注册过的事件
	receiver.addRegInfo(eventType, p)
	//记录当前所属IEventProcessor注册的回调
	receiver.GetEventProcessor().addBindEvent(eventType, receiver, callback)
	//将注册加入到监听中
	p.addListen(eventType, receiver)
}

// UnRegEventReceiverFun 取消注册
func (p *Processor) UnRegEventReceiverFun(eventType int, receiver IHandler) {
	p.removeListen(eventType, receiver)
	receiver.GetEventProcessor().removeBindEvent(eventType, receiver)
	receiver.removeRegInfo(eventType, p)
}

// castEvent 广播事件
func (p *Processor) castEvent(event IEvent) {
	if p.mapListenerEvent == nil {
		//log.Error("mapListenerEvent not init!")
		return
	}

	eventProcessor, ok := p.mapListenerEvent[event.GetEventType()]
	if ok == false || p == nil {
		return
	}

	for proc := range eventProcessor {
		proc.PushEvent(event)
	}
}

// addListen 添加监听
func (p *Processor) addListen(eventType int, receiver IHandler) {
	p.locker.Lock()
	defer p.locker.Unlock()

	if _, ok := p.mapListenerEvent[eventType]; ok == false {
		p.mapListenerEvent[eventType] = map[IProcessor]int{}
	}

	p.mapListenerEvent[eventType][receiver.GetEventProcessor()] += 1
}

// addBindEvent 添加绑定事件
func (p *Processor) addBindEvent(eventType int, receiver IHandler, callback CallBack) {
	p.locker.Lock()
	defer p.locker.Unlock()

	if _, ok := p.mapBindHandlerEvent[eventType]; ok == false {
		p.mapBindHandlerEvent[eventType] = map[IHandler]CallBack{}
	}

	p.mapBindHandlerEvent[eventType][receiver] = callback
}

// removeBindEvent 移除绑定事件
func (p *Processor) removeBindEvent(eventType int, receiver IHandler) {
	p.locker.Lock()
	defer p.locker.Unlock()
	if _, ok := p.mapBindHandlerEvent[eventType]; ok == true {
		delete(p.mapBindHandlerEvent[eventType], receiver)
	}
}

// removeListen 移除监听
func (p *Processor) removeListen(eventType int, receiver IHandler) {
	p.locker.Lock()
	defer p.locker.Unlock()
	if _, ok := p.mapListenerEvent[eventType]; ok == true {
		p.mapListenerEvent[eventType][receiver.GetEventProcessor()] -= 1
		if p.mapListenerEvent[eventType][receiver.GetEventProcessor()] <= 0 {
			delete(p.mapListenerEvent[eventType], receiver.GetEventProcessor())
		}
	}
}
