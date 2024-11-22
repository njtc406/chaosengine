// Package core
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package core

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/event"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/profiler"
	"github.com/njtc406/chaosengine/engine/rpc"
	"github.com/njtc406/chaosengine/engine/utils/asynclib"
	"github.com/njtc406/chaosengine/engine/utils/concurrent"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/timer"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

type Service struct {
	Module

	pid *actor.PID // 服务基础信息

	id          string // 服务唯一id(针对本地节点的相同服务名称中唯一)
	name        string // 服务名称
	serviceType string // 服务类型
	serverId    int32  // 服务id
	version     int64  // 服务版本

	src          inf.IService       // 服务源
	cfg          interface{}        // 服务配置
	status       int32              // 服务状态(0初始化 1启动中 2启动  3关闭中 4关闭 5退休)
	goroutineNum int32              // 协程数量
	wg           sync.WaitGroup     // 协程等待
	closeSignal  chan struct{}      // 关闭信号
	mailBox      chan inf.IEvent    // 事件队列
	profiler     *profiler.Profiler // 性能分析

	rpcHandler     rpc.Handler    // rpc处理器
	eventProcessor inf.IProcessor // 事件管理器
}

func (s *Service) Init(svc interface{}, serviceInitConf *def.ServiceInitConf, cfg interface{}) {
	if serviceInitConf == nil {
		serviceInitConf = &def.ServiceInitConf{
			Name:         "",
			ServerId:     0,
			TimerSize:    def.DefaultTimerSize,
			MailBoxSize:  def.DefaultMailBoxSize,
			GoroutineNum: def.DefaultGoroutineNum,
		}
	}

	// 初始化服务数据
	s.serviceType = serviceInitConf.Type
	if s.serviceType == "" {
		s.serviceType = "Unknown"
	}
	s.serverId = serviceInitConf.ServerId
	s.src = svc.(inf.IService)
	s.cfg = cfg
	if s.closeSignal == nil {
		s.closeSignal = make(chan struct{})
	}
	if s.timerDispatcher == nil {
		var timerSize int
		if serviceInitConf.TimerSize == 0 {
			timerSize = def.DefaultTimerSize
		}
		s.timerDispatcher = timer.NewDispatcher(timerSize)
	}
	if s.mailBox == nil {
		var mailBoxSize int
		if serviceInitConf.MailBoxSize == 0 {
			mailBoxSize = def.DefaultMailBoxSize
		}
		s.mailBox = make(chan inf.IEvent, mailBoxSize)
	}
	s.goroutineNum = serviceInitConf.GoroutineNum
	if s.goroutineNum == 0 {
		s.goroutineNum = def.DefaultGoroutineNum
	}

	// rpc处理器
	s.rpcHandler.Init(svc.(inf.IRpcHandler))

	// 初始化根模块
	s.self = svc.(inf.IModule)
	s.root = s.self
	s.rootContains = make(map[uint32]inf.IModule)
	s.moduleIdSeed = 1
	s.eventProcessor = event.NewProcessor()
	s.eventProcessor.Init(s)
	s.eventHandler = event.NewHandler()
	s.eventHandler.Init(s.eventProcessor)
	s.IConcurrent = concurrent.NewConcurrent()
}

func (s *Service) Start() error {
	// 按理说服务都应该是单线程的被初始化,所以应该不需要这样变更状态的
	if !atomic.CompareAndSwapInt32(&s.status, def.SvcStatusInit, def.SvcStatusStarting) {
		return fmt.Errorf("service[%s] status[%d] has inited", s.GetName(), s.status)
	}
	var waitRun sync.WaitGroup

	s.self.(inf.IService).OnStart()
	for i := int32(0); i < s.goroutineNum; i++ {
		s.wg.Add(1)
		waitRun.Add(1)
		asynclib.Go(func() {
			waitRun.Done()
			s.run()
		})
	}

	waitRun.Wait()

	// 所有服务都注册到服务列表
	s.pid = endpoints.GetEndpointManager().AddService(s.serverId, s.id, s.serviceType, s.name, s.version, s.GetRpcHandler())
	log.SysLogger.Infof(" service[%s] pid: %s", s.GetName(), s.pid.String())
	return nil
}

func (s *Service) run() {
	defer s.wg.Done()

	var bStop bool

	concurrent := s.IConcurrent.(*concurrent.Concurrent)
	concurrentCBChannel := concurrent.GetCallBackChannel()

	for {
		var analyzer *profiler.Analyzer
		select {
		case <-s.closeSignal:
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
		case ev := <-s.mailBox:
			// 事件处理
			switch ev.GetType() {
			case event.SysEventRpc:
				// rpc调用
				cEvent, ok := ev.(*event.Event)
				if !ok {
					log.SysLogger.Error("event type error")
					break
				}
				c := cEvent.Data.(inf.IEnvelope)
				log.SysLogger.Debugf("dddddddddddddddddddddddddd")
				if c.IsReply() {
					if s.profiler != nil {
						analyzer = s.profiler.Push(fmt.Sprintf("[RPCResponse]%s", c.GetMethod()))
					}
					log.SysLogger.Debugf("eeeeeeeeeeeeeeeeeeeeeeeeeeeee")
					// 回复
					s.rpcHandler.HandleResponse(c)
				} else {
					if s.profiler != nil {
						analyzer = s.profiler.Push(fmt.Sprintf("[RPCRequest]%s", c.GetMethod()))
					}
					log.SysLogger.Debugf("fffffffffffffffffffffffffff")
					// rpc调用
					s.rpcHandler.HandleRequest(c)
				}

				event.ReleaseEvent(cEvent)
				if analyzer != nil {
					analyzer.Pop()
					analyzer = nil
				}
			default:
				if s.profiler != nil {
					analyzer = s.profiler.Push(fmt.Sprintf("[SvcEvent][%d]", ev.GetType()))
				}
				s.eventProcessor.EventHandler(ev)
				if analyzer != nil {
					analyzer.Pop()
					analyzer = nil
				}
			}
		case t := <-s.timerDispatcher.ChanTimer:
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
			if len(s.mailBox) > 0 || len(s.timerDispatcher.ChanTimer) > 0 {
				continue
			}

			if atomic.AddInt32(&s.goroutineNum, -1) <= 0 {
				s.release() // 关闭最后一个协程的时候才调用
			}
			break
		}
	}
}

func (s *Service) Stop() {
	close(s.closeSignal)
	s.status = def.SvcStatusClosing
	s.wg.Wait()
	s.status = def.SvcStatusClosed
}

func (s *Service) release() {
	defer func() {
		if err := recover(); err != nil {
			log.SysLogger.Errorf("service [%s] release error: %v\ntrace:%s", s.GetName(), err, debug.Stack())
		}
	}()

	s.self.OnRelease()
	s.closeProfiler()

	// 服务关闭,从服务移除(等待其他释放完再移除,防止在释放的时候有同步调用,例如db等,会导致调用失败)
	endpoints.GetEndpointManager().RemoveService(s.GetPID())
}

func (s *Service) pushEvent(e inf.IEvent) error {
	select {
	case s.mailBox <- e:
	default:
		log.SysLogger.Errorf("service[%s] event channel full", s.GetName())
		return errdef.EventChannelIsFull
	}

	return nil
}

func (s *Service) PushEvent(e inf.IEvent) error {
	return s.pushEvent(e)
}

func (s *Service) PushRequest(c inf.IEnvelope) error {
	log.SysLogger.Debugf("bbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	ev := event.NewEvent()
	ev.Type = event.SysEventRpc
	ev.Data = c
	return s.pushEvent(ev)
}

func (s *Service) PushHttpEvent(e interface{}) error {
	ev := event.NewEvent()
	ev.Type = event.SysEventHttpMsg
	ev.Data = e
	return s.pushEvent(ev)
}

func (s *Service) SetName(name string) {
	s.name = name
}

func (s *Service) GetName() string {
	return s.name
}

func (s *Service) SetServiceId(id string) {
	s.id = id
}

func (s *Service) GetServiceId() string {
	return s.id
}

func (s *Service) GetPID() *actor.PID {
	return s.pid
}

func (s *Service) GetRpcHandler() inf.IRpcHandler {
	return &s.rpcHandler
}

func (s *Service) GetServiceEventChannelNum() int {
	return len(s.mailBox)
}

func (s *Service) GetServiceTimerChannelNum() int {
	return len(s.timerDispatcher.ChanTimer)
}

func (s *Service) OnInit() error {
	return nil
}

func (s *Service) OnStart() error {
	return nil
}

func (s *Service) OnRelease() {}

func (s *Service) SetGoRoutineNum(goroutineNum int32) {
	if atomic.LoadInt32(&s.status) != def.SvcStatusInit { // 已经启动的不允许修改
		return
	}

	s.goroutineNum = goroutineNum
}

func (s *Service) OnSetup(svc inf.IService) {
	if svc.GetName() == "" {
		s.name = reflect.Indirect(reflect.ValueOf(svc)).Type().Name()
	}
}

func (s *Service) IsClosed() bool {
	return atomic.LoadInt32(&s.status) > def.SvcStatusRunning
}

func (s *Service) getUniqueKey() string {
	var name = s.GetName()
	if s.id != "" {
		name = fmt.Sprintf("%s-%s", name, s.id)
	}
	return name
}

func (s *Service) OpenProfiler() {
	s.profiler = profiler.RegProfiler(s.getUniqueKey())
	if s.profiler == nil {
		log.SysLogger.Fatalf("profiler %s reg fail", s.GetName())
	}
}

func (s *Service) GetProfiler() *profiler.Profiler {
	return s.profiler
}

func (s *Service) closeProfiler() {
	if s.profiler != nil {
		profiler.UnRegProfiler(s.getUniqueKey())
	}
}

func (s *Service) GetServiceCfg() interface{} {
	return s.cfg
}
