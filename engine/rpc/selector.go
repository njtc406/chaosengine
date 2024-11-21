package rpc

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints"
	"github.com/njtc406/chaosengine/engine/inf"
)

// Select 选择服务
func (h *Handler) Select(serverId int32, serviceId, serviceName string) inf.IBus {
	return endpoints.GetEndpointManager().Select(h.GetPID(), serverId, serviceId, serviceName)
}

// SelectSameServer 选择相同服务器标识的服务
func (h *Handler) SelectSameServer(serviceId, serviceName string) inf.IBus {
	pid := h.GetPID()
	return endpoints.GetEndpointManager().Select(pid, pid.GetServerId(), serviceId, serviceName)
}

func (h *Handler) SelectByPid(receiver *actor.PID) inf.IBus {
	return endpoints.GetEndpointManager().SelectByPid(h.GetPID(), receiver)
}

// SelectByRule 根据自定义规则选择服务
func (h *Handler) SelectByRule(rule func(pid *actor.PID) bool) inf.IBus {
	return endpoints.GetEndpointManager().SelectByRule(h.GetPID(), rule)
}

func (h *Handler) SelectByNodeType(nodeType, serviceName string) inf.IBus {
	return endpoints.GetEndpointManager().SelectByNodeType(h.GetPID(), nodeType, serviceName)
}
