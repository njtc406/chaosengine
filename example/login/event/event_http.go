// Package event
// @Title  title
// @Description  desc
// @Author  yr  2024/11/20
// @Update  yr  2024/11/20
package event

import (
	sysDto "github.com/njtc406/chaosengine/engine/dto"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/pool"
	"github.com/njtc406/chaosengine/example/login/internal/dto"
)

type HttpEvent struct {
	sysDto.DataRef
	Type inf.EventType
	Data interface{}
	done chan struct{}
	Resp *dto.HttpResponse
}

func (h *HttpEvent) Reset() {
	h.Type = 0
	h.Data = nil
	h.Resp = nil
	if h.done == nil {
		h.done = make(chan struct{}, 1)
	} else {
		if len(h.done) > 0 {
			<-h.done
		}
	}
}

func (h *HttpEvent) GetType() inf.EventType {
	return h.Type
}

func (h *HttpEvent) Done() {
	if h.done != nil {
		h.done <- struct{}{}
	}
}

func (h *HttpEvent) Wait() <-chan struct{} {
	return h.done
}

var httpEventPool = pool.NewPoolEx(make(chan pool.IPoolData, 2048), func() pool.IPoolData {
	return &HttpEvent{}
})

func NewHttpEvent() *HttpEvent {
	return httpEventPool.Get().(*HttpEvent)
}

func ReleaseHttpEvent(e *HttpEvent) {
	e.Reset()
	httpEventPool.Put(e)
}
