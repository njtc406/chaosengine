// Package repository
// @Title  服务存储器
// @Description  用于存放所有服务的注册信息,包括本地和远程的服务信息
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package repository

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
	"sync"
)

type Repository struct {
	mapPID      *sync.Map // 服务 [serviceUid]inf.IClient
	mapNodeLock sync.RWMutex
	// 快速查询表
	mapSvcByNidAndSUid   map[string]map[string]struct{}            // [nodeUid]map[serviceUid]struct{}
	mapSvcBySNameAndSUid map[string]map[string]struct{}            // [serviceName]map[serviceUid]struct{}
	mapSvcByNidAndSName  map[string]map[string]map[string]struct{} // [nodeUid]map[serviceName]map[serviceUid]struct{}
}

func NewRepository() *Repository {
	return &Repository{
		mapPID:               new(sync.Map),
		mapSvcByNidAndSUid:   make(map[string]map[string]struct{}),
		mapSvcBySNameAndSUid: make(map[string]map[string]struct{}),
		mapSvcByNidAndSName:  make(map[string]map[string]map[string]struct{}),
	}
}

func (r *Repository) Add(client inf.IClient) {
	r.mapPID.Store(client.GetPID().GetServiceUid(), client)
	r.mapNodeLock.Lock()
	defer r.mapNodeLock.Unlock()

	pid := client.GetPID()
	nodeUid := pid.GetNodeUid()
	serviceName := pid.GetName()
	serviceUid := pid.GetServiceUid()

	nodeUidMap, ok := r.mapSvcByNidAndSUid[nodeUid]
	if !ok {
		r.mapSvcByNidAndSUid[nodeUid] = make(map[string]struct{})
		nodeUidMap = r.mapSvcByNidAndSUid[nodeUid]
	}
	_, ok = nodeUidMap[serviceUid]
	if !ok {
		nodeUidMap[serviceUid] = struct{}{}
	}

	nameMap, ok := r.mapSvcBySNameAndSUid[serviceName]
	if !ok {
		r.mapSvcBySNameAndSUid[serviceName] = make(map[string]struct{})
		nameMap = r.mapSvcBySNameAndSUid[serviceName]
	}

	_, ok = nameMap[serviceUid]
	if !ok {
		nameMap[serviceUid] = struct{}{}
	}

	nodeNameUidMap, ok := r.mapSvcByNidAndSName[nodeUid]
	if !ok {
		r.mapSvcByNidAndSName[nodeUid] = make(map[string]map[string]struct{})
		nodeNameUidMap = r.mapSvcByNidAndSName[nodeUid]
	}

	nameUidMap, ok := nodeNameUidMap[serviceName]
	if !ok {
		nodeNameUidMap[serviceName] = make(map[string]struct{})
		nameUidMap = nodeNameUidMap[serviceName]
	}

	_, ok = nameUidMap[serviceUid]
	if !ok {
		nameUidMap[serviceUid] = struct{}{}
	}
}

func (r *Repository) Remove(pid *actor.PID) {
	client, ok := r.mapPID.LoadAndDelete(pid.GetServiceUid())
	if !ok {
		return
	}

	client.(inf.IClient).Close()

	r.mapNodeLock.Lock()
	defer r.mapNodeLock.Unlock()
	serviceName := pid.GetName()
	serviceUid := pid.GetServiceUid()
	nodeUid := pid.GetNodeUid()

	nodeUidMap, ok := r.mapSvcByNidAndSUid[nodeUid]
	if ok {
		delete(nodeUidMap, serviceUid)
		if len(nodeUidMap) == 0 {
			delete(r.mapSvcByNidAndSUid, nodeUid)
		}
	} else {
		return
	}

	nameMap, ok := r.mapSvcBySNameAndSUid[serviceName]
	if ok {
		delete(nameMap, serviceUid)
		if len(nameMap) == 0 {
			delete(r.mapSvcBySNameAndSUid, serviceName)
		}
	} else {
		return
	}

	nodeNameUidMap, ok := r.mapSvcByNidAndSName[nodeUid]
	if ok {
		nameUidMap, ok := nodeNameUidMap[serviceName]
		if ok {
			delete(nameUidMap, serviceUid)
			if len(nameUidMap) == 0 {
				delete(nodeNameUidMap, serviceName)
			}
		}
		if len(nodeNameUidMap) == 0 {
			delete(r.mapSvcByNidAndSName, nodeUid)
		}
	}
}
