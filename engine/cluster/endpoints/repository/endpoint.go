// Package repository
// @Title  单个服务信息
// @Description  desc
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package repository

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints/internal/client"
	"github.com/njtc406/chaosengine/engine/utils/synclib"
)

type Endpoint struct {
	synclib.DataRef
	pid    *actor.PID
	client *client.Client
}

func (e *Endpoint) Init(pid *actor.PID, client *client.Client) *Endpoint {
	e.pid = pid
	e.client = client
	return e
}

func (e *Endpoint) Reset() {
	e.pid = nil
	if e.client != nil {
		e.client.Close()
	}
	e.client = nil
}

var endpointPool = synclib.NewPoolEx(make(chan synclib.IPoolData, 1024), func() synclib.IPoolData {
	return &Endpoint{}
})

func NewEndpoint() *Endpoint {
	return endpointPool.Get().(*Endpoint)
}

func releaseEndpoint(e *Endpoint) {
	if e != nil {
		e.Reset()
		endpointPool.Put(e)
	}
}

func (e *Endpoint) GetPID() *actor.PID {
	return e.pid
}

func (e *Endpoint) GetClient() *client.Client {
	return e.client
}
