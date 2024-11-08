// Package monitor
// @Title  rpc调用监视器
// @Description  用于监控rpc的call调用,当超时发生时自动回调,防止一直阻塞
// @Author  pc  2024/11/6
// @Update  pc  2024/11/6
package monitor

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"sync"
	"sync/atomic"
	"time"
)

type RpcMonitor struct {
	closed      chan struct{}
	locker      sync.Mutex
	seed        uint64
	waitMap     map[uint64]*actor.Future
	th          CallTimerHeap // 由于请求很频繁,所以这里使用单独的timer来处理
	ticker      *time.Ticker
	callFailFun func(f *actor.Future)
}

func NewRpcMonitor() *RpcMonitor {
	return &RpcMonitor{}
}

func (rm *RpcMonitor) Init(fun func(f *actor.Future)) {
	rm.closed = make(chan struct{})
	rm.waitMap = make(map[uint64]*actor.Future)
	rm.callFailFun = fun
	rm.th.Init()
	rm.ticker = time.NewTicker(time.Millisecond * 100)
}

func (rm *RpcMonitor) Start() {
	go rm.listen()
}

func (rm *RpcMonitor) Stop() {
	close(rm.closed)
	rm.ticker.Stop()
}

func (rm *RpcMonitor) listen() {
	for {
		select {
		case <-rm.ticker.C:
			rm.tick()
		case <-rm.closed:
			return
		}
	}
}

func (rm *RpcMonitor) tick() {
	for i := 0; i < 1000; i++ { // 每个tick 最多处理1000个超时的rpc
		rm.locker.Lock() // 放里面,防止tick期间一直占用锁
		id := rm.th.PopTimeout()
		if id == 0 {
			rm.locker.Unlock()
			break
		}

		f := rm.waitMap[id]
		if f == nil {
			rm.locker.Unlock()
			log.SysLogger.Errorf("call seq is not find,seq:%d", id)
			continue
		}

		delete(rm.waitMap, id)
		//log.SysLogger.Debugf("RPC call takes more than %d seconds,method is %s", int64(f.GetTimeout().Seconds()), f.GetMethod())
		fmt.Printf("RPC call takes more than %d seconds,method is %s", int64(f.GetTimeout().Seconds()), f.GetMethod())
		// 调用超时,执行超时回调
		if rm.callFailFun != nil {
			rm.callFailFun(f)
		}
		rm.locker.Unlock()
		continue
	}
}

func (rm *RpcMonitor) GenSeq() uint64 {
	return atomic.AddUint64(&rm.seed, 1)
}

func (rm *RpcMonitor) Add(f *actor.Future) {
	rm.locker.Lock()
	defer rm.locker.Unlock()

	id := f.GetReqID()
	if id == 0 {
		return
	}
	rm.waitMap[id] = f
	rm.th.AddTimer(id, f.GetTimeout())
}

func (rm *RpcMonitor) remove(id uint64) *actor.Future {
	f, ok := rm.waitMap[id]
	if !ok {
		return nil
	}

	rm.th.Cancel(id)
	delete(rm.waitMap, id)
	return f
}

func (rm *RpcMonitor) Remove(id uint64) *actor.Future {
	if id == 0 {
		return nil
	}
	rm.locker.Lock()
	f := rm.remove(id)
	rm.locker.Unlock()
	return f
}

func (rm *RpcMonitor) Get(id uint64) *actor.Future {
	rm.locker.Lock()
	defer rm.locker.Unlock()

	return rm.waitMap[id]
}
