package monitor

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/errdef"
	"testing"
	"time"
)

func TestRpcMonitor_Add(t *testing.T) {
	rm := NewRpcMonitor()
	rm.Init(nil)
	rm.Start()
	defer rm.Stop()
	f := actor.NewFuture()
	f.SetTimeout(time.Second)
	rm.Add(f)
}

func callFail(f *actor.Future) {
	if !f.IsRef() {
		//log.SysLogger.Errorf("future is not ref,pid:%s", f.GetSender().String())
		return // 已经被释放,丢弃
	}

	//if f.NeedCallback() {
	//	// TODO 需要回复,那么构建一个回复信封
	//}

	f.SetResult(nil, errdef.RPCCallTimeout)
}

func TestRpcMonitor_Get(t *testing.T) {
	rm := NewRpcMonitor()
	rm.Init(callFail)
	rm.Start()
	defer rm.Stop()
	f := actor.NewFuture()
	f.SetReqID(1)
	f.SetTimeout(time.Second)
	rm.Add(f)
	f2 := rm.Get(rm.GenSeq())
	if f2 == nil {
		t.Error("get future failed")
	}
	f2.Wait()
}

func TestRpcMonitor_Remove(t *testing.T) {
	rm := NewRpcMonitor()
	rm.Init(callFail)
	rm.Start()
	defer rm.Stop()
	f := actor.NewFuture()
	f.SetReqID(1)
	f.SetTimeout(time.Second)
	rm.Add(f)
	rm.Remove(rm.GenSeq())
	nf := rm.Get(rm.GenSeq())
	if nf != nil {
		t.Error("remove failed")
	}
}
