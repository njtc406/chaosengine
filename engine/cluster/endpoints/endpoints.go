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
	inf.IEventProcessor
	inf.IEventHandler

	remote     inf.IRemoteServer      // 通讯器(这里之后可以根据类型扩展为多个)
	stopped    bool                   // 是否已停止
	repository *repository.Repository // 服务存储仓库
}

func GetEndpointManager() *EndpointManager {
	return endMgr
}

func (em *EndpointManager) Init(eventProcessor inf.IEventProcessor) *EndpointManager {
	em.remote = remote.GetRemote(def.DefaultRpcTypeRpcx)
	log.SysLogger.Debugf("cluster rpc server config: %+v", config.Conf.ClusterConf.RPCServer)
	em.remote.Init(config.Conf.ClusterConf.RPCServer, em)

	em.IEventProcessor = eventProcessor

	// 事件管理
	em.IEventProcessor = eventProcessor
	em.IEventHandler = event.NewHandler()
	em.IEventHandler.Init(em.IEventProcessor)

	em.repository = repository.NewRepository()

	return em
}

func (em *EndpointManager) Start() {
	// 启动rpc监听服务器
	if err := em.remote.Serve(); err != nil {
		log.SysLogger.Fatalf("start rpc server error: %v", err)
	}
	// 新增、修改服务事件
	em.IEventProcessor.RegEventReceiverFunc(event.SysEventETCDPut, em.IEventHandler, em.updateServiceInfo)
	// 删除服务事件
	em.IEventProcessor.RegEventReceiverFunc(event.SysEventETCDDel, em.IEventHandler, em.removeServiceInfo)
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
			log.SysLogger.Debugf("endpointmgr ignore -> remote: %s local: %s  pid:%s", pid.GetNodeUid(), em.remote.GetNodeUid(), pid.String())
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
	pid := actor.NewPID(em.remote.GetAddress(), em.remote.GetNodeUid(), serverId, serviceId, serviceType, serviceName, version, rpcType)
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
	em.IEventProcessor.EventHandler(ev)

	return pid
}

func (em *EndpointManager) RemoveService(pid *actor.PID) {
	em.repository.Remove(pid)
}

func (em *EndpointManager) GetClient(pid *actor.PID) inf.IRpcSender {
	cli := em.repository.SelectByServiceUid(pid.GetServiceUid())
	if cli == nil {
		// 只有一种情况下可能是空的,就是调用者是私有服务,那么此时就单独创建一个
		// TODO 这里考虑将私有服务的client加入到一个临时列表中,防止每次都创建,然后加入一个过期机制,多久没有调用就删除
		return client.NewSender(pid.GetRpcType())(pid, nil)
	}
	return cli
}
