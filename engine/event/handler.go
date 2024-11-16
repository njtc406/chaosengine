// Package event
// @Title  事件处理器
// @Description  用于给事件一个注册绑定,标注这个事件在哪里注册过
// @Author  yr  2024/7/19 下午3:31
// @Update  yr  2024/7/19 下午3:31
package event

import (
	"github.com/njtc406/chaosengine/engine/inf"
	"sync"
)

type Handler struct {
	sync.RWMutex
	processor   inf.IProcessor
	mapRegEvent map[inf.EventType]map[inf.IProcessor]interface{}
}

func NewHandler() inf.IHandler {
	return &Handler{}
}

func (h *Handler) Init(p inf.IProcessor) {
	h.processor = p
	h.mapRegEvent = make(map[inf.EventType]map[inf.IProcessor]interface{})
}

func (h *Handler) GetEventProcessor() inf.IProcessor {
	return h.processor
}

func (h *Handler) NotifyEvent(ev inf.IEvent) {
	h.GetEventProcessor().CastEvent(ev)
}

func (h *Handler) Destroy() {
	h.Lock()
	defer h.Unlock()
	for eventTyp, mapEventProcess := range h.mapRegEvent {
		if mapEventProcess == nil {
			continue
		}

		for eventProcess := range mapEventProcess {
			eventProcess.UnRegEventReceiverFun(eventTyp, h)
		}
	}
}

func (h *Handler) AddRegInfo(eventType inf.EventType, eventProcessor inf.IProcessor) {
	h.Lock()
	defer h.Unlock()
	if h.mapRegEvent == nil {
		h.mapRegEvent = map[inf.EventType]map[inf.IProcessor]interface{}{}
	}

	if _, ok := h.mapRegEvent[eventType]; ok == false {
		h.mapRegEvent[eventType] = map[inf.IProcessor]interface{}{}
	}
	h.mapRegEvent[eventType][eventProcessor] = nil
}

func (h *Handler) RemoveRegInfo(eventType inf.EventType, eventProcessor inf.IProcessor) {
	if _, ok := h.mapRegEvent[eventType]; ok == true {
		delete(h.mapRegEvent[eventType], eventProcessor)
	}
}
