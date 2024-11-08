// Package event
// @Title  事件处理器
// @Description  用于给事件一个注册绑定,标注这个事件在哪里注册过
// @Author  yr  2024/7/19 下午3:31
// @Update  yr  2024/7/19 下午3:31
package event

import (
	"sync"
)

type Handler struct {
	sync.RWMutex
	processor   IProcessor
	mapRegEvent map[int]map[IProcessor]interface{}
}

func NewHandler() IHandler {
	return &Handler{}
}

func (h *Handler) Init(p IProcessor) {
	h.processor = p
	h.mapRegEvent = make(map[int]map[IProcessor]interface{})
}

func (h *Handler) GetEventProcessor() IProcessor {
	return h.processor
}

func (h *Handler) NotifyEvent(ev IEvent) {
	h.GetEventProcessor().castEvent(ev)
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

func (h *Handler) addRegInfo(eventType int, eventProcessor IProcessor) {
	h.Lock()
	defer h.Unlock()
	if h.mapRegEvent == nil {
		h.mapRegEvent = map[int]map[IProcessor]interface{}{}
	}

	if _, ok := h.mapRegEvent[eventType]; ok == false {
		h.mapRegEvent[eventType] = map[IProcessor]interface{}{}
	}
	h.mapRegEvent[eventType][eventProcessor] = nil
}

func (h *Handler) removeRegInfo(eventType int, eventProcessor IProcessor) {
	if _, ok := h.mapRegEvent[eventType]; ok == true {
		delete(h.mapRegEvent[eventType], eventProcessor)
	}
}
