package event

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/pool"
	"runtime"
	"sync"
)

type Event struct {
	def.DataRef
	Type inf.EventType
	Data interface{}
}

var emptyEvent Event

func (e *Event) Reset() {
	*e = emptyEvent
}

func (e *Event) GetType() inf.EventType {
	return e.Type
}

type EventHandler struct {
	//已经注册的事件类型
	eventProcessor inf.IEventProcessor

	//已经注册的事件
	locker      sync.RWMutex
	mapRegEvent map[inf.EventType]map[inf.IEventProcessor]interface{} //向其他事件处理器监听的事件类型
}

type EventProcessor struct {
	inf.IEventChannel

	locker              sync.RWMutex
	mapListenerEvent    map[inf.EventType]map[inf.IEventProcessor]int             //监听者信息
	mapBindHandlerEvent map[inf.EventType]map[inf.IEventHandler]inf.EventCallBack //收到事件处理
}

func NewEventHandler() inf.IEventHandler {
	eh := EventHandler{}
	eh.mapRegEvent = map[inf.EventType]map[inf.IEventProcessor]interface{}{}

	return &eh
}

func NewEventProcessor() inf.IEventProcessor {
	ep := EventProcessor{}
	ep.mapListenerEvent = map[inf.EventType]map[inf.IEventProcessor]int{}
	ep.mapBindHandlerEvent = map[inf.EventType]map[inf.IEventHandler]inf.EventCallBack{}

	return &ep
}

func (handler *EventHandler) AddRegInfo(eventType inf.EventType, eventProcessor inf.IEventProcessor) {
	handler.locker.Lock()
	defer handler.locker.Unlock()
	if handler.mapRegEvent == nil {
		handler.mapRegEvent = map[inf.EventType]map[inf.IEventProcessor]interface{}{}
	}

	if _, ok := handler.mapRegEvent[eventType]; ok == false {
		handler.mapRegEvent[eventType] = map[inf.IEventProcessor]interface{}{}
	}
	handler.mapRegEvent[eventType][eventProcessor] = nil
}

func (handler *EventHandler) RemoveRegInfo(eventType inf.EventType, eventProcessor inf.IEventProcessor) {
	if _, ok := handler.mapRegEvent[eventType]; ok == true {
		delete(handler.mapRegEvent[eventType], eventProcessor)
	}
}

func (handler *EventHandler) GetEventProcessor() inf.IEventProcessor {
	return handler.eventProcessor
}

func (handler *EventHandler) NotifyEvent(ev inf.IEvent) {
	handler.GetEventProcessor().CastEvent(ev)
}

func (handler *EventHandler) Init(processor inf.IEventProcessor) {
	handler.eventProcessor = processor
	handler.mapRegEvent = map[inf.EventType]map[inf.IEventProcessor]interface{}{}
}

func (processor *EventProcessor) Init(eventChannel inf.IEventChannel) {
	processor.IEventChannel = eventChannel
}

func (processor *EventProcessor) AddBindEvent(eventType inf.EventType, receiver inf.IEventHandler, callback inf.EventCallBack) {
	processor.locker.Lock()
	defer processor.locker.Unlock()

	if _, ok := processor.mapBindHandlerEvent[eventType]; ok == false {
		processor.mapBindHandlerEvent[eventType] = map[inf.IEventHandler]inf.EventCallBack{}
	}

	processor.mapBindHandlerEvent[eventType][receiver] = callback
}

func (processor *EventProcessor) AddListen(eventType inf.EventType, receiver inf.IEventHandler) {
	processor.locker.Lock()
	defer processor.locker.Unlock()

	if _, ok := processor.mapListenerEvent[eventType]; ok == false {
		processor.mapListenerEvent[eventType] = map[inf.IEventProcessor]int{}
	}

	processor.mapListenerEvent[eventType][receiver.GetEventProcessor()] += 1
}

func (processor *EventProcessor) RemoveBindEvent(eventType inf.EventType, receiver inf.IEventHandler) {
	processor.locker.Lock()
	defer processor.locker.Unlock()
	if _, ok := processor.mapBindHandlerEvent[eventType]; ok == true {
		delete(processor.mapBindHandlerEvent[eventType], receiver)
	}
}

func (processor *EventProcessor) RemoveListen(eventType inf.EventType, receiver inf.IEventHandler) {
	processor.locker.Lock()
	defer processor.locker.Unlock()
	if _, ok := processor.mapListenerEvent[eventType]; ok == true {
		processor.mapListenerEvent[eventType][receiver.GetEventProcessor()] -= 1
		if processor.mapListenerEvent[eventType][receiver.GetEventProcessor()] <= 0 {
			delete(processor.mapListenerEvent[eventType], receiver.GetEventProcessor())
		}
	}
}

func (processor *EventProcessor) RegEventReceiverFunc(eventType inf.EventType, receiver inf.IEventHandler, callback inf.EventCallBack) {
	//记录receiver自己注册过的事件
	receiver.AddRegInfo(eventType, processor)
	//记录当前所属IEventProcessor注册的回调
	receiver.GetEventProcessor().AddBindEvent(eventType, receiver, callback)
	//将注册加入到监听中
	processor.AddListen(eventType, receiver)
}

func (processor *EventProcessor) UnRegEventReceiverFun(eventType inf.EventType, receiver inf.IEventHandler) {
	processor.RemoveListen(eventType, receiver)
	receiver.GetEventProcessor().RemoveBindEvent(eventType, receiver)
	receiver.RemoveRegInfo(eventType, processor)
}

func (handler *EventHandler) Destroy() {
	handler.locker.Lock()
	defer handler.locker.Unlock()
	for eventTyp, mapEventProcess := range handler.mapRegEvent {
		if mapEventProcess == nil {
			continue
		}

		for eventProcess := range mapEventProcess {
			eventProcess.UnRegEventReceiverFun(eventTyp, handler)
		}
	}
}

func (processor *EventProcessor) EventHandler(ev inf.IEvent) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			errString := fmt.Sprint(r)
			log.SysLogger.Errorf("%s error %s", string(buf[:l]), errString)
		}
	}()

	mapCallBack, ok := processor.mapBindHandlerEvent[ev.GetType()]
	if ok == false {
		return
	}
	for _, callback := range mapCallBack {
		callback(ev)
	}
}

func (processor *EventProcessor) CastEvent(event inf.IEvent) {
	if processor.mapListenerEvent == nil {
		log.SysLogger.Error("mapListenerEvent not init!")
		return
	}

	eventProcessor, ok := processor.mapListenerEvent[event.GetType()]
	if ok == false || processor == nil {
		return
	}

	for proc := range eventProcessor {
		proc.PushEvent(event)
	}
}

var eventPool = pool.NewPoolEx(make(chan pool.IPoolData, 10240), func() pool.IPoolData {
	return &Event{}
})

func NewEvent() *Event {
	return eventPool.Get().(*Event)
}

func ReleaseEvent(e *Event) {
	eventPool.Put(e)
}
