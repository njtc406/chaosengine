// Package repository
// @Title  服务选择器
// @Description  根据条件选择服务
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package repository

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/messagebus"
)

func (r *Repository) SelectByServiceUid(serviceUid string) inf.IRpcSender {
	v, ok := r.mapPID.Load(serviceUid)
	if ok {
		ep := v.(inf.IRpcSender)
		if ep != nil && !actor.IsRetired(ep.GetPID()) {
			return ep
		}
	}
	return nil
}

func (r *Repository) SelectByPid(sender, receiver *actor.PID) inf.IBus {
	s := r.SelectByServiceUid(sender.GetServiceUid())
	c := r.SelectByServiceUid(receiver.GetServiceUid())
	if c != nil && !actor.IsRetired(c.GetPID()) {
		b := messagebus.NewMessageBus(s, c)
		return b
	}
	return nil
}

func (r *Repository) SelectBySvcUid(sender *actor.PID, serviceUid string) inf.IBus {
	s := r.SelectByServiceUid(sender.GetServiceUid())
	c := r.SelectByServiceUid(serviceUid)

	if c != nil && !actor.IsRetired(c.GetPID()) {
		b := messagebus.NewMessageBus(s, c)
		return b
	}
	return nil
}

func (r *Repository) SelectByRule(sender *actor.PID, rule func(pid *actor.PID) bool) inf.IBus {
	s := r.SelectByServiceUid(sender.GetServiceUid())
	var returnList messagebus.MultiBus
	r.mapPID.Range(func(key, value any) bool {
		if rule(value.(inf.IRpcSender).GetPID()) {
			returnList = append(returnList, messagebus.NewMessageBus(s, value.(inf.IRpcSender)))
		}
		return true
	})

	return returnList
}

func (r *Repository) Select(sender *actor.PID, serverId int32, serviceId, serviceName string) inf.IBus {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()
	serviceUid := actor.CreateServiceUid(serverId, serviceName, serviceId)
	return r.SelectBySvcUid(sender, serviceUid)
}

func (r *Repository) SelectByNodeType(sender *actor.PID, nodeType, serviceName string) inf.IBus {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()

	var returnList messagebus.MultiBus
	nameUidMap, ok := r.mapSvcByNtpAndSName[nodeType]
	if !ok {
		return returnList
	}

	serviceList, ok := nameUidMap[serviceName]
	if !ok {
		return returnList
	}

	s := r.SelectByServiceUid(sender.GetServiceUid())

	for serviceUid := range serviceList {
		c := r.SelectByServiceUid(serviceUid)
		if c != nil && !actor.IsRetired(c.GetPID()) {
			returnList = append(returnList, messagebus.NewMessageBus(s, c))
		}
	}

	return returnList
}
