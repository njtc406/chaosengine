package endpoints

import (
	"github.com/njtc406/chaosengine/engine1/actor"
	"github.com/njtc406/chaosengine/engine1/define/inf"
	"github.com/njtc406/chaosengine/engine1/errdef/errcode"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosutil/chaoserrors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type CallSet struct {
	pendingLock          sync.RWMutex
	startSeq             uint64
	pending              map[uint64]*actor.Future
	callRpcTimeout       time.Duration
	maxCheckCallRpcCount int

	callTimerHeap CallTimerHeap
}

func (cs *CallSet) Init() {
	cs.pendingLock.Lock()
	cs.callTimerHeap.Init()
	cs.pending = make(map[uint64]*actor.Future, 4096)

	cs.maxCheckCallRpcCount = DefaultMaxCheckCallRpcCount
	cs.callRpcTimeout = DefaultRpcTimeout

	go cs.checkRpcCallTimeout()
	cs.pendingLock.Unlock()
}

func (cs *CallSet) makeCallFail(future *actor.Future) {
	if !future.IsRef() {
		log.SysLogger.Errorf("future is not ref,pid:%s", future.GetSender().String())
		// 已经被释放,丢弃
		return
	}

	if future.NeedCallback() {
		// 需要回复,异步回调,需要构建response
		resp := actor.NewMsgEnvelope()
		resp.SetReply() // 这是一条回复信息
		resp.SetError(chaoserrors.NewErrCode(errcode.RpcCallTimeout, "RPC call timeout"))
		resp.AddCompletion(future.GetCompletions()...)
		resp.Receiver = future.GetSender()

		// 获取send
		client := endpointManager.GetClientByPID(future.GetSender())
		if client == nil {
			log.SysLogger.Errorf("client is nil,pid:%s", future.GetSender().String())
			actor.ReleaseMsgEnvelope(resp)
			return
		}
		// (这里的envelope会在两个地方回收,如果是本地调用,那么会在requestHandler执行完成后自动回收
		// 如果是远程调用,那么在远程client将消息发送完成后自动回收)
		if err := client.SendMessage(resp); err != nil {
			actor.ReleaseMsgEnvelope(resp)
			log.SysLogger.Errorf("send call timeout response error:%s", err.Error())
		}
	}

	future.SetResult(nil, chaoserrors.NewErrCode(errcode.RpcCallTimeout, "RPC call timeout1"))
}

func (cs *CallSet) checkRpcCallTimeout() {
	for {
		time.Sleep(DefaultCheckRpcCallTimeoutInterval)
		for i := 0; i < cs.maxCheckCallRpcCount; i++ {
			cs.pendingLock.Lock()

			reqID := cs.callTimerHeap.PopTimeout()
			if reqID == 0 {
				cs.pendingLock.Unlock()
				break
			}

			future := cs.pending[reqID]
			if future == nil {
				cs.pendingLock.Unlock()
				log.SysLogger.Errorf("call seq is not find,seq:%d", reqID)
				continue
			}

			delete(cs.pending, reqID) // 删除

			log.SysLogger.Errorf("RPC call takes more than %s seconds,method is %s", strconv.FormatInt(int64(future.GetTimeout().Seconds()), 10), future.GetMethod())

			cs.makeCallFail(future)
			cs.pendingLock.Unlock()
			continue
		}
	}
}

func (cs *CallSet) AddPending(future *actor.Future) {
	cs.pendingLock.Lock()
	defer cs.pendingLock.Unlock()
	reqID := future.GetReqID()

	if reqID == 0 {
		return
	}
	cs.pending[reqID] = future
	cs.callTimerHeap.AddTimer(reqID, future.GetTimeout())
}

func (cs *CallSet) RemovePending(seq uint64) *actor.Future {
	if seq == 0 {
		return nil
	}
	cs.pendingLock.Lock()
	call := cs.removePending(seq)
	cs.pendingLock.Unlock()
	return call
}

func (cs *CallSet) removePending(seq uint64) *actor.Future {
	v, ok := cs.pending[seq]
	if ok == false {
		return nil
	}

	cs.callTimerHeap.Cancel(seq)
	delete(cs.pending, seq)
	return v
}

func (cs *CallSet) FindPending(seq uint64) *actor.Future {
	if seq == 0 {
		return nil
	}

	cs.pendingLock.Lock()
	pCall := cs.pending[seq]
	cs.pendingLock.Unlock()

	return pCall
}

func (cs *CallSet) cleanPending() {
	cs.pendingLock.Lock()
	for {
		callSeq := cs.callTimerHeap.PopFirst()
		if callSeq == 0 {
			break
		}
		pCall := cs.pending[callSeq]
		if pCall == nil {
			//log.Error("call Seq is not find", log.Uint64("seq", callSeq))
			continue
		}

		delete(cs.pending, callSeq)
		//pCall.err = errors.New("nodeid is disconnect ")
		cs.makeCallFail(pCall)
	}

	cs.pendingLock.Unlock()
}

func (cs *CallSet) GenerateSeq() uint64 {
	return atomic.AddUint64(&cs.startSeq, 1)
}

type RpcCancel struct {
	Cli     *Client
	CallSeq uint64
}

func (rc *RpcCancel) CancelRpc() {
	rc.Cli.RemovePending(rc.CallSeq)
}

func NewRpcCancel(cli *Client, seq uint64) inf.CancelRpc {
	cancel := &RpcCancel{Cli: cli, CallSeq: seq}
	return cancel.CancelRpc
}
