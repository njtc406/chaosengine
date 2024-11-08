// Package service
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/7/19 上午10:42
// @Update  yr  2024/7/19 上午10:42
package service

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine1/actor"
	"github.com/njtc406/chaosengine/engine1/cluster/endpoints"
	"github.com/njtc406/chaosengine/engine1/concurrent"
	"github.com/njtc406/chaosengine/engine1/define/consts"
	"github.com/njtc406/chaosengine/engine1/define/inf"
	"github.com/njtc406/chaosengine/engine1/errdef"
	"github.com/njtc406/chaosengine/engine1/event"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosengine/engine1/profiler"
	"github.com/njtc406/chaosengine/engine1/rpc"
	"github.com/njtc406/chaosengine/engine1/timer"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var maxServiceEventChannelNum = 2000000

type Service struct {
	Module

	pid          *actor.PID         // 服务唯一标识
	name         string             // 服务名称
	id           string             // 服务id(用于区分相同服务的不同对象)
	base         interface{}        // 服务对象
	serviceCfg   interface{}        // 服务配置
	status       bool               // 运行状态
	goroutineNum int32              // 最大并发数
	wg           sync.WaitGroup     // 等待goroutine退出
	closeSig     chan struct{}      // 关闭信号
	profiler     *profiler.Profiler // 性能监控

	rpcHandler     rpc.Handler       // rpc处理器
	eventCh        chan event.IEvent // 消息队列
	eventProcessor event.IProcessor  // 事件管理器(用于分配事件)
}

func (s *Service) RegEventReceiverFunc(eventType int, receiver event.IHandler, callback event.CallBack) {
	s.eventProcessor.RegEventReceiverFunc(eventType, receiver, callback)
}

func (s *Service) GetName() string {
	return s.name
}

func (s *Service) SetName(name string) {
	s.name = name
}

func (s *Service) GetID() string {
	return s.id
}

func (s *Service) SetID(id string) {
	s.id = id
}

func (s *Service) SetPID(pid *actor.PID) {
	s.pid = pid
}

func (s *Service) GetPID() *actor.PID {
	return s.pid
}

func (s *Service) GetServiceCfg() interface{} {
	return s.serviceCfg
}

func (s *Service) OnSetup(svc inf.IService) {
	if svc.GetName() == "" {
		s.name = reflect.Indirect(reflect.ValueOf(svc)).Type().Name()
	}
}

func (s *Service) getUniqueKey() string {
	var name = s.GetName()
	if s.GetID() != "" {
		name = fmt.Sprintf("%s-%s", name, s.GetID())
	}
	return name
}

func (s *Service) OpenProfiler() {
	s.profiler = profiler.RegProfiler(s.getUniqueKey())
	if s.profiler == nil {
		log.SysLogger.Fatalf("profiler %s reg fail", s.GetName())
	}
}

func (s *Service) CloseProfiler() {
	if s.profiler != nil {
		profiler.UnRegProfiler(s.getUniqueKey())
	}
}

func (s *Service) GetProfiler() *profiler.Profiler {
	return s.profiler
}

func (s *Service) Init(svc interface{}, cfg interface{}) {
	// TODO: 如果之后需要不同配置,这里就再加一个config参数,根据config参数来初始化
	s.serviceCfg = cfg
	s.closeSig = make(chan struct{})
	s.dispatcher = timer.NewDispatcher(100)
	if s.eventCh == nil {
		s.eventCh = make(chan event.IEvent, 10240)
	}
	s.rpcHandler.InitHandler(svc.(inf.IRpcHandler), svc.(inf.IChannel))
	s.IRpcHandler = &s.rpcHandler
	s.self = svc.(inf.IModule)
	s.base = svc
	s.goroutineNum = 1
	s.ancestor = svc.(inf.IModule)
	s.seedModuleId = consts.InitModuleId
	s.descendants = make(map[uint32]inf.IModule)
	s.eventProcessor = event.NewProcessor()
	s.eventProcessor.Init(s)
	s.eventHandler = event.NewHandler()
	s.eventHandler.Init(s.eventProcessor)
	s.Module.IConcurrent = &concurrent.Concurrent{} // 这里加不加Module有什么区别?
	//s.moduleName = reflect.Indirect(reflect.ValueOf(svc)).Type().Name()
}

func (s *Service) OnInit() error {
	return nil
}

func (s *Service) PushRequest(c *actor.MsgEnvelope) error {
	ev := event.NewEvent()
	ev.Type = event.SysEventRpc
	ev.Data = c

	return s.pushEvent(ev)
}

func (s *Service) PushClientMsg(msg interface{}) error {
	ev := event.NewEvent()
	ev.Type = event.SysEventClientMsg
	ev.Data = msg
	return s.pushEvent(ev)
}

func (s *Service) PushEvent(ev event.IEvent) error {
	return s.pushEvent(ev)
}

func (s *Service) pushEvent(e event.IEvent) error {
	if len(s.eventCh) >= maxServiceEventChannelNum {
		log.SysLogger.Errorf("service[%s] event channel full", s.GetName())
		return errdef.EventChannelIsFull
	}
	s.eventCh <- e
	return nil
}

func (s *Service) Start() {
	s.status = true
	var waitRun sync.WaitGroup

	s.self.(inf.IService).OnStart()
	for i := int32(0); i < s.goroutineNum; i++ {
		s.wg.Add(1)
		waitRun.Add(1)
		go func() {
			waitRun.Done()
			s.run()
		}()
	}

	waitRun.Wait()

	// 所有服务都注册到本地服务列表
	//Register(s.GetName(), s.GetID(), s.base)
	s.pid = endpoints.GetEndpointManager().AddService(s.id, s.name, s.GetRpcHandler())
	log.SysLogger.Infof(" service[%s] pid: %s", s.GetName(), s.pid.String())
}

func (s *Service) run() {
	log.SysLogger.Infof(" service[%s] begin running", s.GetName())
	defer s.wg.Done()
	var bStop bool

	concurrent := s.IConcurrent.(*concurrent.Concurrent)
	concurrentCBChannel := concurrent.GetCallBackChannel()

	for {
		var analyzer *profiler.Analyzer
		select {
		case <-s.closeSig:
			if !bStop {
				bStop = true // 关闭信号
				concurrent.Close()
			}
		case cb := <-concurrentCBChannel:
			if s.profiler != nil {
				analyzer = s.profiler.Push(fmt.Sprintf("[Concurrent]%s", reflect.TypeOf(cb).String()))
			}
			concurrent.DoCallback(cb) // 异步执行
			if analyzer != nil {
				analyzer.Pop()
				analyzer = nil
			}
		case ev := <-s.eventCh:
			// 事件处理
			switch ev.GetEventType() {
			case event.SysEventRpc:
				// rpc调用
				cEvent, ok := ev.(*event.Event)
				if !ok {
					log.SysLogger.Error("event type error")
					break
				}
				c := cEvent.Data.(*actor.MsgEnvelope)
				if c.IsReply() {
					if s.profiler != nil {
						analyzer = s.profiler.Push(fmt.Sprintf("[RPCReply]%s", c.GetMethod()))
					}
					// 回复
					s.rpcHandler.HandlerResponse(c)
				} else {
					if s.profiler != nil {
						analyzer = s.profiler.Push(fmt.Sprintf("[RPCRequest]%s", c.GetMethod()))
					}
					// rpc调用
					s.rpcHandler.HandleRequest(c)
				}

				actor.ReleaseMsgEnvelope(c)
				event.ReleaseEvent(cEvent)
				if analyzer != nil {
					analyzer.Pop()
					analyzer = nil
				}
			case event.SysEventClientMsg:
				cEvent, ok := ev.(*event.Event)
				if !ok {
					log.SysLogger.Error("event type error")
					break
				}
				if s.profiler != nil {
					analyzer = s.profiler.Push(fmt.Sprintf("[ClientMsg][%d]", cEvent.GetEventType()))
				}
				s.rpcHandler.HandlerClientMsg(cEvent.Data)
				event.ReleaseEvent(cEvent)
				if analyzer != nil {
					analyzer.Pop()
					analyzer = nil
				}
			default:
				if s.profiler != nil {
					analyzer = s.profiler.Push(fmt.Sprintf("[SvcEvent][%d]", ev.GetEventType()))
				}
				s.eventProcessor.EventHandler(ev)
				if analyzer != nil {
					analyzer.Pop()
					analyzer = nil
				}
			}
		case t := <-s.dispatcher.ChanTimer:
			// 定时器处理
			if s.profiler != nil {
				analyzer = s.profiler.Push("[timer]" + s.GetName() + "." + t.GetName())
			}
			t.Do()
			if analyzer != nil {
				analyzer.Pop()
				analyzer = nil
			}
		}

		if bStop {
			// 等待所有channel处理完成后关闭
			if len(s.eventCh) > 0 || len(s.dispatcher.ChanTimer) > 0 {
				continue
			}

			if atomic.AddInt32(&s.goroutineNum, -1) <= 0 {
				s.status = false
				s.release()
			}
			break
		}
	}
}

func (s *Service) OnStart() {

}

func (s *Service) Stop() {
	close(s.closeSig)
	s.wg.Wait()
}

func (s *Service) release() {
	defer func() {
		if err := recover(); err != nil {
			log.SysLogger.Errorf("service [%s] release error: %v\ntrace:%s", s.GetName(), err, debug.Stack())
		}
	}()

	s.self.OnRelease() // 这里是为了能调用到子类的方法
	s.CloseProfiler()

	// 服务关闭,从本地服务移除
	//UnRegister(s.GetName(), s.GetID())
	endpoints.GetEndpointManager().RemoveService(s.GetPID())
}

func (s *Service) OnRelease() {}

func (s *Service) RegRawHandler(rawRpcCB rpc.RawCallback) {
	s.rpcHandler.RegisterRawHandler(rawRpcCB)
}

// SetGoRoutineNum 设置服务协程数量
func (s *Service) SetGoRoutineNum(goroutineNum int32) bool {
	//已经开始状态不允许修改协程数量,打开性能分析器不允许开多线程
	if s.status == true || s.profiler != nil {
		log.SysLogger.Errorf("service [%s] can not set goroutine num", s.GetName())
		return false
	}

	s.goroutineNum = goroutineNum
	return true
}

func (s *Service) GetServiceEventChannelNum() int {
	return len(s.eventCh)
}
func (s *Service) GetServiceTimerChannelNum() int {
	return len(s.dispatcher.ChanTimer)
}
