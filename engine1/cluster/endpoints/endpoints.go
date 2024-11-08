// Package endpoints
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/8/29 下午6:24
// @Update  yr  2024/8/29 下午6:24
package endpoints

import (
	"github.com/njtc406/chaosengine/engine1/actor"
	interfacedef2 "github.com/njtc406/chaosengine/engine1/define/inf"
	"github.com/njtc406/chaosengine/engine1/event"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosengine/engine1/synclib"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"google.golang.org/protobuf/encoding/protojson"
	"sync"
)

type endpoint struct {
	synclib.DataRef
	pid    *actor.PID
	client *Client
}

func (e *endpoint) Reset() {
	e.pid = nil
	if e.client != nil {
		e.client.Close()
	}
	e.client = nil
}

var endpointPool = synclib.NewPoolEx(make(chan synclib.IPoolData, 1024), func() synclib.IPoolData {
	return &endpoint{}
})

func newEndpoint() *endpoint {
	return endpointPool.Get().(*endpoint)
}

func releaseEndpoint(e *endpoint) {
	if e != nil {
		e.Reset()
		endpointPool.Put(e)
	}
}

func (e *endpoint) GetPID() *actor.PID {
	return e.pid
}

func (e *endpoint) GetClient() *Client {
	return e.client
}

var endpointManager = &EndpointManager{}

type EndpointManager struct {
	event.IProcessor
	event.IHandler
	interfacedef2.ICallSet

	remote      *Remote   // 通讯器
	stopped     bool      // 是否已停止
	mapPID      *sync.Map // 服务 [*pid]*endpoint
	mapNodeLock sync.RWMutex
	mapNode     map[string]map[string]map[string][]*endpoint // map[nodeUID]map[serviceName]map[serviceID][]*endpoint
}

func GetEndpointManager() *EndpointManager {
	return endpointManager
}

func (em *EndpointManager) Init(nodeUID, addr string, eventProcessor event.IProcessor, callSet interfacedef2.ICallSet) {
	em.remote = NewRemote(nodeUID, addr)
	em.remote.Init()

	em.IProcessor = eventProcessor
	em.ICallSet = callSet

	// 事件管理
	em.IProcessor = eventProcessor
	em.IHandler = event.NewHandler()
	em.IHandler.Init(em.IProcessor)

	em.mapPID = &sync.Map{}
	em.mapNode = make(map[string]map[string]map[string][]*endpoint)
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
}

func (em *EndpointManager) Stop() {
	em.stopped = true
}

func (em *EndpointManager) GetUID() string {
	return em.remote.GetNodeUID()
}

// updateServiceInfo 更新远程服务信息
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
		if pid.Address == em.remote.GetAddress() {
			log.SysLogger.Debugf("local service: %s, pid: %+v", pid.String(), &pid)
			// 本地服务,忽略
			return
		}

		client := NewRemoteClient(&pid, em.ICallSet, nil)
		ep := newEndpoint()
		ep.pid = &pid
		ep.client = client
		em.addService(&pid, ep)
	}
}

// removeServiceInfo 删除远程服务信息
func (em *EndpointManager) removeServiceInfo(e event.IEvent) {
	ev := e.(*event.Event)
	kv := ev.Data.(*mvccpb.KeyValue)
	if kv.Value != nil {
		var pid actor.PID
		if err := protojson.Unmarshal(kv.Value, &pid); err != nil {
			log.SysLogger.Errorf("unmarshal pid error: %v", err)
			return
		}
		ep := em.GetServiceByPID(&pid)
		if ep == nil {
			log.SysLogger.Warnf("service not found: %s", pid.String())
			return
		}

		em.removeService(&pid)
	}
}

func (em *EndpointManager) addService(pid *actor.PID, ep *endpoint) {
	em.mapPID.Store(pid.GetUid(), ep)
	em.mapNodeLock.Lock()
	defer em.mapNodeLock.Unlock()
	nodeInfo, ok := em.mapNode[pid.GetNodeUID()]
	if !ok {
		em.mapNode[pid.GetNodeUID()] = make(map[string]map[string][]*endpoint)
		nodeInfo = em.mapNode[pid.GetNodeUID()]
	}
	serviceInfo, ok := nodeInfo[pid.GetName()]
	if !ok {
		nodeInfo[pid.GetName()] = make(map[string][]*endpoint)
		serviceInfo = nodeInfo[pid.GetName()]
	}

	_, ok = serviceInfo[pid.GetId()]
	if !ok {
		serviceInfo[pid.GetId()] = []*endpoint{}
	}
	serviceInfo[pid.GetId()] = append(serviceInfo[pid.GetId()], ep)

	//log.SysLogger.Debugf("add service: %s, pid: %v", pid.String(), ep)
}

func (em *EndpointManager) removeService(pid *actor.PID) {
	em.mapNodeLock.Lock()
	defer em.mapNodeLock.Unlock()
	if nodeInfo, ok := em.mapNode[em.remote.GetNodeUID()]; ok {
		if serviceInfo, ok := nodeInfo[pid.GetName()]; ok {
			if epList, ok := serviceInfo[pid.GetId()]; ok {
				for idx, ep := range epList {
					if ep.pid.GetUid() == pid.GetUid() {
						if idx == len(epList)-1 {
							serviceInfo[pid.GetId()] = epList[:idx]
						} else {
							serviceInfo[pid.GetId()] = append(epList[:idx], epList[idx+1:]...)
						}

						// 如果是删除本地服务, 从etcd中删除服务
						if pid.GetAddress() == em.remote.GetAddress() {
							ev := event.NewEvent()
							ev.Type = event.SysEventServiceDis
							ev.Data = pid
							em.IProcessor.EventHandler(ev)
						}
					}
				}
				if len(epList) == 0 {
					delete(serviceInfo, pid.GetId())
				}
			}
			if len(nodeInfo[pid.GetName()]) == 0 {
				delete(nodeInfo, pid.GetName())
			}
		}
	}
}

// AddService 添加本地服务到服务发现中
func (em *EndpointManager) AddService(serviceID, serviceName string, rpcHandler interfacedef2.IRpcHandler) *actor.PID {
	pid := actor.NewPID(em.remote.GetNodeUID(), em.remote.GetAddress(), serviceID, serviceName)
	// 生成client
	log.SysLogger.Debugf("add local service: %s, pid: %v", pid.String(), rpcHandler)
	client := NewLClient(pid, em.ICallSet, rpcHandler)
	ep := newEndpoint()
	ep.pid = pid
	ep.client = client

	// 添加服务列表
	em.addService(pid, ep)

	// 私有服务不发布到etcd
	if rpcHandler.IsPrivate() {
		return pid
	}

	log.SysLogger.Debugf("add service to etcd: %s, pid: %v", pid.String(), ep)

	// 将服务信息发布到etcd
	ev := event.NewEvent()
	ev.Type = event.SysEventServiceReg
	ev.Data = pid
	em.IProcessor.EventHandler(ev)

	return pid
}

func (em *EndpointManager) RemoveService(pid *actor.PID) {
	if ep, ok := em.mapPID.Load(pid.GetUid()); ok {
		defer releaseEndpoint(ep.(*endpoint))
		em.removeService(pid)
	}
}

func (em *EndpointManager) GetServiceByPID(pid *actor.PID) *endpoint {
	if ep, ok := em.mapPID.Load(pid.GetUid()); ok {
		return ep.(*endpoint)
	}
	return nil
}

func (em *EndpointManager) GetClientByPID(pid *actor.PID) *Client {
	if ep, ok := em.mapPID.Load(pid.GetUid()); ok {
		cli := ep.(*endpoint).GetClient()
		if cli != nil {
			return cli
		}
	}
	return nil
}

func (em *EndpointManager) GetServiceByID(nodeUID, serviceID, serviceName string, isAll bool) []*Client {
	//log.SysLogger.Debugf("get service: %s-%s-%s", nodeUID, serviceID, serviceName)
	var clients []*Client
	em.mapNodeLock.RLock()
	defer em.mapNodeLock.RUnlock()

	processEndpoints := func(endpoints []*endpoint) {
		//log.SysLogger.Debugf("endpoints len: %d", len(endpoints))
		for _, ep := range endpoints {
			if !actor.IsRetired(ep.pid) {
				clients = append(clients, ep.client)
			}
		}
	}

	getServiceInfo := func(nodeInfo map[string]map[string][]*endpoint) {
		//log.SysLogger.Debugf("nodeInfo len: %d", len(nodeInfo))
		if serviceInfo, ok := nodeInfo[serviceName]; ok {
			if serviceID != "" {
				if eps, ok := serviceInfo[serviceID]; ok {
					processEndpoints(eps)
				}
			} else {
				for _, eps := range serviceInfo {
					processEndpoints(eps)
				}
			}
		}
	}

	if nodeUID != "" {
		if nodeInfo, ok := em.mapNode[nodeUID]; ok {
			getServiceInfo(nodeInfo)
		}
	} else {
		for _, nodeInfo := range em.mapNode {
			getServiceInfo(nodeInfo)
		}
	}

	// 如果有多个,则返回第一个(这里会出现这个情况可能是在更新服务,老的服务还未下线,新的服务已上线)
	if !isAll && len(clients) > 1 {
		return clients[:1]
	}

	return clients
}
