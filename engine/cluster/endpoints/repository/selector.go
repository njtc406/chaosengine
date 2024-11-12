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
	"math/rand"
)

func (r *Repository) selectByServiceUid(serviceUid string) inf.IClient {
	v, ok := r.mapPID.Load(serviceUid)
	if ok {
		ep := v.(inf.IClient)
		if ep != nil && !actor.IsRetired(ep.GetPID()) {
			return ep
		}
	}
	return nil
}

func (r *Repository) SelectByPid(sender, receiver *actor.PID) inf.IBus {
	c := r.selectByServiceUid(receiver.GetServiceUid())
	if c != nil {
		b := messagebus.NewMessageBus(sender, c)
		return b
	}
	return nil
}

func (r *Repository) SelectBySvcUid(sender *actor.PID, serviceUid string) inf.IBus {
	c := r.selectByServiceUid(serviceUid)
	if c != nil {
		b := messagebus.NewMessageBus(sender, c)
		return b
	}
	return nil
}

// SelectByNodeUidAndSvcName 根据节点 UID 和服务名称选择服务
func (r *Repository) SelectByNodeUidAndSvcName(sender *actor.PID, nodeUid, serviceName string) inf.IBus {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()
	var returnList inf.MultiBus
	if nodeInfo, ok := r.mapSvcByNidAndSName[nodeUid]; ok {
		if serviceInfo, ok := nodeInfo[serviceName]; ok {
			for serviceUid := range serviceInfo {
				c := r.selectByServiceUid(serviceUid)
				if c != nil && !actor.IsRetired(c.GetPID()) {
					returnList = append(returnList, messagebus.NewMessageBus(sender, c))
				}
			}
		}
	}

	return returnList
}

func (r *Repository) SelectAllByName(sender *actor.PID, serviceName string) inf.IBus {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()
	var returnList inf.MultiBus
	nameMap, ok := r.mapSvcBySNameAndSUid[serviceName]
	if !ok {
		return returnList
	}

	for serviceUid := range nameMap {
		c := r.selectByServiceUid(serviceUid)
		if c != nil && !actor.IsRetired(c.GetPID()) {
			returnList = append(returnList, messagebus.NewMessageBus(sender, c))
		}
	}

	return returnList
}

func (r *Repository) SelectRandomByName(sender *actor.PID, serviceName string) inf.IBus {
	list := r.SelectAllByName(sender, serviceName)

	if len(list.(inf.MultiBus)) == 0 {
		return nil
	}

	if len(list.(inf.MultiBus)) == 1 {
		return list.(inf.MultiBus)[0]
	}

	idx := rand.Intn(len(list.(inf.MultiBus)))
	return list.(inf.MultiBus)[idx]
}

func (r *Repository) SelectNameByRule(sender *actor.PID, serviceName string, rule func(pid *actor.PID) bool) inf.IBus {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()
	var returnList inf.MultiBus
	nameMap, ok := r.mapSvcBySNameAndSUid[serviceName]
	if !ok {
		return returnList
	}

	for serviceUid := range nameMap {
		c := r.selectByServiceUid(serviceUid)
		if c != nil && !actor.IsRetired(c.GetPID()) && rule(c.GetPID()) {
			returnList = append(returnList, messagebus.NewMessageBus(sender, c))
		}
	}

	return returnList
}

func (r *Repository) SelectByRule(sender *actor.PID, rule func(pid *actor.PID) bool) inf.IBus {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()
	var returnList inf.MultiBus
	for _, serviceInfo := range r.mapSvcByNidAndSUid {
		for serviceUid := range serviceInfo {
			c := r.selectByServiceUid(serviceUid)
			if c != nil && !actor.IsRetired(c.GetPID()) && rule(c.GetPID()) {
				returnList = append(returnList, messagebus.NewMessageBus(sender, c))
			}
		}
	}

	return returnList
}
