// Package repository
// @Title  服务选择器
// @Description  根据条件选择服务
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package repository

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/errdef"
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
		b := messagebus.NewMessageBus(s, c, nil)
		return b
	}
	return messagebus.NewMessageBus(s, c, errdef.ServiceNotFound)
}

func (r *Repository) SelectBySvcUid(sender *actor.PID, serviceUid string) inf.IBus {
	s := r.SelectByServiceUid(sender.GetServiceUid())
	c := r.SelectByServiceUid(serviceUid)

	if c != nil && !actor.IsRetired(c.GetPID()) {
		b := messagebus.NewMessageBus(s, c, nil)
		return b
	}
	return messagebus.NewMessageBus(s, c, errdef.ServiceNotFound)
}

func (r *Repository) SelectByRule(sender *actor.PID, rule func(pid *actor.PID) bool) inf.IBus {
	s := r.SelectByServiceUid(sender.GetServiceUid())
	var returnList messagebus.MultiBus
	r.mapPID.Range(func(key, value any) bool {
		if rule(value.(inf.IRpcSender).GetPID()) {
			returnList = append(returnList, messagebus.NewMessageBus(s, value.(inf.IRpcSender), nil))
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

func (r *Repository) SelectByServiceType(sender *actor.PID, serverId int32, nodeType, serviceName string) inf.IBus {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()

	var list messagebus.MultiBus
	nameUidMap, ok := r.mapSvcBySTpAndSName[nodeType]
	if !ok {
		return list
	}

	var serviceList []string
	if serviceName == "" {
		for _, uidMap := range nameUidMap {
			for serviceUid := range uidMap {
				serviceList = append(serviceList, serviceUid)
			}
		}
	} else {
		uidMap, ok := nameUidMap[serviceName]
		if !ok {
			return list
		}
		for serviceUid := range uidMap {
			serviceList = append(serviceList, serviceUid)
		}
	}

	s := r.SelectByServiceUid(sender.GetServiceUid())

	for _, serviceUid := range serviceList {
		c := r.SelectByServiceUid(serviceUid)
		if c != nil && !actor.IsRetired(c.GetPID()) && (serverId == 0 || c.GetPID().GetServerId() == serverId) {
			list = append(list, messagebus.NewMessageBus(s, c, nil))
		}
	}

	return list
}

func (r *Repository) SelectByFilterAndChoice(sender *actor.PID, filter func(pid *actor.PID) bool, choice func(pids []*actor.PID) []*actor.PID) inf.IBus {
	s := r.SelectByServiceUid(sender.GetServiceUid())
	var tmpList []*actor.PID
	r.mapPID.Range(func(key, value any) bool {
		if filter(value.(inf.IRpcSender).GetPID()) {
			tmpList = append(tmpList, value.(inf.IRpcSender).GetPID())
		}
		return true
	})

	list := choice(tmpList)
	var returnList messagebus.MultiBus
	for _, pid := range list {
		c := r.SelectByServiceUid(pid.GetServiceUid())
		if c != nil && !actor.IsRetired(c.GetPID()) {
			returnList = append(returnList, messagebus.NewMessageBus(s, c, nil))
		}
	}

	return returnList
}
