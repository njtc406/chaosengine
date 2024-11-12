// Package endpoints
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/8/29 下午6:24
// @Update  yr  2024/8/29 下午6:24
package endpoints

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints/internal/client"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints/repository"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/event"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/rpc/monitor"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"google.golang.org/protobuf/encoding/protojson"
)

var endMgr = &EndpointManager{}

type EndpointManager struct {
	event.IProcessor
	event.IHandler

	remote     *Remote                // 通讯器
	stopped    bool                   // 是否已停止
	rpcMonitor inf.IMonitor           // rpc调用的监视器
	repository *repository.Repository // 服务存储仓库
}

func GetEndpointManager() *EndpointManager {
	return endMgr
}

func (em *EndpointManager) Init(nodeUID, addr string, eventProcessor event.IProcessor) {
	em.remote = NewRemote(nodeUID, addr)
	em.remote.Init()

	em.IProcessor = eventProcessor

	// 事件管理
	em.IProcessor = eventProcessor
	em.IHandler = event.NewHandler()
	em.IHandler.Init(em.IProcessor)

	em.repository = repository.NewRepository()

	em.rpcMonitor = monitor.NewRpcMonitor()
	em.rpcMonitor.Init(futureCallTimeout)
}

func (em *EndpointManager) Start() {
	// 启动rpc监听服务器
	if err := em.remote.Serve(); err != nil {
		log.SysLogger.Fatalf("start rpc server error: %v", err)
	}
	// 新增、修改服务事件
	em.IProcessor.RegEventReceiverFunc(event.SysEventETCDPut, em.IHandler, em.updateServiceInfo)
	// 删除服务事件
	em.IProcessor.RegEventReceiverFunc(event.SysEventETCDDel, em.IHandler, em.removeServiceInfo)
	em.rpcMonitor.Start()
}

func (em *EndpointManager) Stop() {
	em.stopped = true
	em.rpcMonitor.Stop()
}

func (em *EndpointManager) GetUID() string {
	return em.remote.GetNodeUID()
}

// updateServiceInfo 更新远程服务信息事件
func (em *EndpointManager) updateServiceInfo(e event.IEvent) {
	//log.SysLogger.Debugf("endpoints receive update service event: %+v", e)
	ev := e.(*event.Event)
	kv := ev.Data.(*mvccpb.KeyValue)
	if kv.Value != nil {
		var pid actor.PID
		if err := protojson.Unmarshal(kv.Value, &pid); err != nil {
			log.SysLogger.Errorf("unmarshal pid error: %v", err)
			return
		}
		if pid.GetNodeUid() == em.remote.GetNodeUID() {
			log.SysLogger.Debugf("local service: %s, pid: %+v", pid.String(), &pid)
			// 本地服务,忽略
			return
		}

		em.repository.Add(client.NewRemoteClient(&pid, em.rpcMonitor))
	}
}

// removeServiceInfo 删除远程服务信息事件
func (em *EndpointManager) removeServiceInfo(e event.IEvent) {
	ev := e.(*event.Event)
	kv := ev.Data.(*mvccpb.KeyValue)
	if kv.Value != nil {
		var pid actor.PID
		if err := protojson.Unmarshal(kv.Value, &pid); err != nil {
			log.SysLogger.Errorf("unmarshal pid error: %v", err)
			return
		}

		em.repository.Remove(&pid)
	}
}

// AddService 添加本地服务到服务发现中
func (em *EndpointManager) AddService(serverId int32, serviceID, serviceName string, version int32, rpcHandler inf.IRpcHandler) *actor.PID {
	pid := actor.NewPID(em.remote.GetNodeUID(), em.remote.GetAddress(), serverId, serviceID, serviceName, version)
	log.SysLogger.Debugf("add local service: %s, pid: %v", pid.String(), rpcHandler)
	em.repository.Add(client.NewLClient(pid, em.rpcMonitor, rpcHandler))

	// 私有服务不发布到etcd
	if rpcHandler.IsPrivate() {
		return pid
	}

	log.SysLogger.Debugf("add service to etcd ,pid: %v", pid.String())

	// 将服务信息发布到etcd
	ev := event.NewEvent()
	ev.Type = event.SysEventServiceReg
	ev.Data = pid
	em.IProcessor.EventHandler(ev)

	return pid
}

func (em *EndpointManager) RemoveService(pid *actor.PID) {
	em.repository.Remove(pid)
}

func (em *EndpointManager) GetClient(serviceUid string) inf.IClient {
	return em.repository.SelectBySvcUid(serviceUid)
}

// SelectBySvcUid 根据服务 UID 选择服务
func (em *EndpointManager) SelectBySvcUid(serviceUid string) inf.IRpcInvoker {
	invoker := em.repository.SelectBySvcUid(serviceUid)
	if invoker != nil {
		return invoker
	}
	return inf.NewEmptyMultiRpcInvoker()
}

// SelectByNodeUidAndSvcName 根据节点 UID 和服务名称选择服务
func (em *EndpointManager) SelectByNodeUidAndSvcName(nodeUID string, serviceName string) inf.IRpcInvoker {
	invokerList := em.repository.SelectByNodeUidAndSvcName(nodeUID, serviceName)
	var ret inf.MultiRpcInvoker
	for _, c := range invokerList {
		ret = append(ret, c)
	}
	return ret
}

// SelectAllByName 选择某个服务名称的所有服务
func (em *EndpointManager) SelectAllByName(serviceName string) inf.IRpcInvoker {
	invokerList := em.repository.SelectAllByName(serviceName)
	var ret inf.MultiRpcInvoker
	for _, c := range invokerList {
		ret = append(ret, c)
	}
	return ret
}

// SelectRandomByName 随机选择某个服务名称的服务
func (em *EndpointManager) SelectRandomByName(serviceName string) inf.IRpcInvoker {
	c := em.repository.SelectRandomByName(serviceName)
	if c != nil {
		return c
	}
	return inf.NewEmptyMultiRpcInvoker()
}

// SelectNameByRule 根据规则选择某个服务名称的服务
func (em *EndpointManager) SelectNameByRule(serviceName string, rule func(pid *actor.PID) bool) inf.IRpcInvoker {
	invokerList := em.repository.SelectNameByRule(serviceName, rule)
	var ret inf.MultiRpcInvoker
	for _, c := range invokerList {
		ret = append(ret, c)
	}
	return ret
}

// SelectByRule 根据自定义规则选择服务
func (em *EndpointManager) SelectByRule(rule func(pid *actor.PID) bool) inf.IRpcInvoker {
	invokerList := em.repository.SelectByRule(rule)
	var ret inf.MultiRpcInvoker
	for _, c := range invokerList {
		ret = append(ret, c)
	}
	return ret
}

func futureCallTimeout(f *actor.Future) {
	if !f.IsRef() {
		log.SysLogger.Errorf("future is not ref,pid:%s", f.GetSender().String())
		return // 已经被释放,丢弃
	}

	if f.NeedCallback() {
		resp := actor.NewMsgEnvelope()
		resp.SetReply() // 这是一条回复信息
		resp.SetError(errdef.RPCCallTimeout)
		resp.AddCompletion(f.GetCompletions()...)
		resp.Receiver = f.GetSender()

		// 获取send
		client := endMgr.GetClient(f.GetSender().GetServiceUid())
		if client == nil {
			log.SysLogger.Errorf("client is nil,pid:%s", f.GetSender().String())
			actor.ReleaseMsgEnvelope(resp)
			return
		}
		// (这里的envelope会在两个地方回收,如果是本地调用,那么会在requestHandler执行完成后自动回收
		// 如果是远程调用,那么在远程client将消息发送完成后自动回收)
		if err := client.PushRequest(resp); err != nil {
			actor.ReleaseMsgEnvelope(resp)
			log.SysLogger.Errorf("send call timeout response error:%s", err.Error())
		}
	}

	f.SetResult(nil, errdef.RPCCallTimeout)
}
