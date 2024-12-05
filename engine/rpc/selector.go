package rpc

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/messagebus"
	"github.com/njtc406/chaosengine/engine/utils/log"
)

// Select 选择服务
func (h *Handler) Select(serverId int32, serviceId, serviceName string) inf.IBus {
	return endpoints.GetEndpointManager().Select(h.GetPid(), serverId, serviceId, serviceName)
}

// SelectSameServer 选择相同服务器标识的服务
func (h *Handler) SelectSameServer(serviceId, serviceName string) inf.IBus {
	pid := h.GetPid()
	return endpoints.GetEndpointManager().Select(pid, pid.GetServerId(), serviceId, serviceName)
}

func (h *Handler) SelectByPid(receiver *actor.PID) inf.IBus {
	return endpoints.GetEndpointManager().SelectByPid(h.GetPid(), receiver)
}

// SelectByRule 根据自定义规则选择服务
func (h *Handler) SelectByRule(rule func(pid *actor.PID) bool) inf.IBus {
	return endpoints.GetEndpointManager().SelectByRule(h.GetPid(), rule)
}

// SelectByServiceType 根据类型选择服务
func (h *Handler) SelectByServiceType(serverId int32, serviceType, serviceName string, filters ...func(pid *actor.PID) bool) inf.IBus {
	list := endpoints.GetEndpointManager().SelectByServiceType(h.GetPid(), serverId, serviceType, serviceName)

	var returnList messagebus.MultiBus
	if len(filters) == 0 {
		return list
	}
	for _, filter := range filters {
		for _, bus := range list.(messagebus.MultiBus) {
			if filter(bus.(inf.IRpcSender).GetPid()) {
				returnList = append(returnList, bus)
			}
		}
	}
	return returnList
}

func (h *Handler) SelectByFilterAndChoice(filter func(pid *actor.PID) bool, choice func(pids []*actor.PID) []*actor.PID) inf.IBus {
	return endpoints.GetEndpointManager().SelectByFilterAndChoice(h.GetPid(), filter, choice)
}

func (h *Handler) SelectSameServerByServiceType(serviceType, serviceName string, filters ...func(pid *actor.PID) bool) inf.IBus {
	list := endpoints.GetEndpointManager().SelectByServiceType(h.GetPid(), h.GetPid().GetServerId(), serviceType, serviceName)
	log.SysLogger.Debugf("list len: %+v", list)
	var returnList messagebus.MultiBus
	if len(filters) == 0 {
		return list
	}
	for _, filter := range filters {
		for _, bus := range list.(messagebus.MultiBus) {
			if filter(bus.(inf.IRpcSender).GetPid()) {
				returnList = append(returnList, bus)
			}
		}
	}
	return returnList
}
