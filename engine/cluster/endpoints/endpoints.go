// Package endpoints
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/8/29 下午6:24
// @Update  yr  2024/8/29 下午6:24
package endpoints

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints/client"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints/remote"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints/repository"
	"github.com/njtc406/chaosengine/engine/config"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/event"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"google.golang.org/protobuf/encoding/protojson"
)

var endMgr = &EndpointManager{}

type EndpointManager struct {
	inf.IProcessor
	inf.IHandler

	remote     inf.IRemoteServer      // 通讯器(这里之后可以根据类型扩展为多个)
	stopped    bool                   // 是否已停止
	repository *repository.Repository // 服务存储仓库
}

func GetEndpointManager() *EndpointManager {
	return endMgr
}

func (em *EndpointManager) Init(eventProcessor inf.IProcessor) *EndpointManager {
	em.remote = remote.GetRemote(def.DefaultRpcTypeRpcx)
	log.SysLogger.Debugf("cluster rpc server config: %+v", config.Conf.ClusterConf.RPCServer)
	em.remote.Init(config.Conf.ClusterConf.RPCServer, em)

	em.IProcessor = eventProcessor

	// 事件管理
	em.IProcessor = eventProcessor
	em.IHandler = event.NewHandler()
	em.IHandler.Init(em.IProcessor)

	em.repository = repository.NewRepository()

	return em
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
	em.remote.Close()
}

// updateServiceInfo 更新远程服务信息事件
func (em *EndpointManager) updateServiceInfo(e inf.IEvent) {
	//log.SysLogger.Debugf("endpoints receive update service event: %+v", e)
	ev := e.(*event.Event)
	kv := ev.Data.(*mvccpb.KeyValue)
	if kv.Value != nil {
		var pid actor.PID
		if err := protojson.Unmarshal(kv.Value, &pid); err != nil {
			log.SysLogger.Errorf("unmarshal pid error: %v", err)
			return
		}
		if pid.GetNodeUid() == em.remote.GetNodeUid() {
			log.SysLogger.Debugf("local service: %s, pid: %+v", pid.String(), &pid)
			// 本地服务,忽略
			return
		}

		senderCreator := client.NewSender(pid.GetRpcType())
		if senderCreator == nil {
			log.SysLogger.Errorf("create sender error: %s", pid.String())
			return
		}

		em.repository.Add(senderCreator(&pid, nil))
	}
}

// removeServiceInfo 删除远程服务信息事件
func (em *EndpointManager) removeServiceInfo(e inf.IEvent) {
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
func (em *EndpointManager) AddService(serverId int32, serviceId, serviceType, serviceName string, version int64, rpcType string, rpcHandler inf.IRpcHandler) *actor.PID {
	pid := actor.NewPID(em.remote.GetAddress(), serverId, serviceId, serviceType, serviceName, version, rpcType)
	log.SysLogger.Debugf("add local service: %s, pid: %v", pid.String(), rpcHandler)
	senderCreator := client.NewSender(def.DefaultRpcTypeLocal)
	if senderCreator == nil {
		log.SysLogger.Errorf("create sender error: %s", pid.String())
		return nil
	}

	em.repository.Add(senderCreator(pid, rpcHandler))

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

func (em *EndpointManager) GetClient(serviceUid string) inf.IRpcSender {
	return em.repository.SelectByServiceUid(serviceUid)
}

func (em *EndpointManager) Select(sender *actor.PID, serverId int32, serviceId, serviceName string) inf.IBus {
	return em.repository.Select(sender, serverId, serviceId, serviceName)
}

func (em *EndpointManager) SelectByPid(sender, receiver *actor.PID) inf.IBus {
	return em.repository.SelectByPid(sender, receiver)
}

// SelectByRule 根据自定义规则选择服务
func (em *EndpointManager) SelectByRule(sender *actor.PID, rule func(pid *actor.PID) bool) inf.IBus {
	return em.repository.SelectByRule(sender, rule)
}

func (em *EndpointManager) SelectByServiceType(sender *actor.PID, nodeType, serviceName string) inf.IBus {
	return em.repository.SelectByServiceType(sender, nodeType, serviceName)
}
