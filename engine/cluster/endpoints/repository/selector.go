// Package repository
// @Title  服务选择器
// @Description  根据条件选择服务
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package repository

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints/internal/client"
	"math/rand"
)

func (r *Repository) SelectBySvcUid(serviceUid string) *client.Client {
	v, ok := r.mapPID.Load(serviceUid)
	if ok {
		ep := v.(*client.Client)
		if ep != nil && !actor.IsRetired(ep.GetPID()) {
			return ep
		}
	}
	return nil
}

// SelectByNodeUidAndSvcName 根据节点 UID 和服务名称选择服务
func (r *Repository) SelectByNodeUidAndSvcName(nodeUid, serviceName string) []*client.Client {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()
	var returnList []*client.Client
	if nodeInfo, ok := r.mapSvcByNidAndSName[nodeUid]; ok {
		if serviceInfo, ok := nodeInfo[serviceName]; ok {
			for serviceUid := range serviceInfo {
				ep := r.SelectBySvcUid(serviceUid)
				if ep != nil && !actor.IsRetired(ep.GetPID()) {
					returnList = append(returnList, ep)
				}
			}
		}
	}

	return returnList
}

func (r *Repository) SelectAllByName(serviceName string) []*client.Client {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()
	var returnList []*client.Client
	nameMap, ok := r.mapSvcBySNameAndSUid[serviceName]
	if !ok {
		return returnList
	}

	for serviceUid := range nameMap {
		ep := r.SelectBySvcUid(serviceUid)
		if ep != nil && !actor.IsRetired(ep.GetPID()) {
			returnList = append(returnList, ep)
		}
	}

	return returnList
}

func (r *Repository) SelectRandomByName(serviceName string) *client.Client {
	list := r.SelectAllByName(serviceName)
	if len(list) == 0 {
		return nil
	}
	if len(list) == 1 {
		return list[0]
	}

	idx := rand.Intn(len(list))
	return list[idx]
}

func (r *Repository) SelectNameByRule(serviceName string, rule func(pid *actor.PID) bool) []*client.Client {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()
	var returnList []*client.Client
	nameMap, ok := r.mapSvcBySNameAndSUid[serviceName]
	if !ok {
		return returnList
	}

	for serviceUid := range nameMap {
		ep := r.SelectBySvcUid(serviceUid)
		if ep != nil && !actor.IsRetired(ep.GetPID()) && rule(ep.GetPID()) {
			returnList = append(returnList, ep)
		}
	}

	return returnList
}

func (r *Repository) SelectByRule(rule func(pid *actor.PID) bool) []*client.Client {
	r.mapNodeLock.RLock()
	defer r.mapNodeLock.RUnlock()
	var returnList []*client.Client
	for _, serviceInfo := range r.mapSvcByNidAndSUid {
		for serviceUid := range serviceInfo {
			ep := r.SelectBySvcUid(serviceUid)
			if ep != nil && !actor.IsRetired(ep.GetPID()) && rule(ep.GetPID()) {
				returnList = append(returnList, ep)
			}
		}
	}

	return returnList
}
