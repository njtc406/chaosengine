// Package actor
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/9/4 下午5:53
// @Update  yr  2024/9/4 下午5:53
package actor

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/def"
	"sync/atomic"
)

func CreateServiceUid(serverId int32, serviceName, serviceId string) string {
	return fmt.Sprintf("%d:%s@%s", serverId, serviceName, serviceId)
}

func NewPID(nodeUID, address string, serverId int32, serviceID, serviceName string, version int64) *PID {
	if serviceID == "" {
		serviceID = nodeUID
	}
	uid := CreateServiceUid(serverId, serviceName, serviceID)
	return &PID{
		Address:    address,
		NodeUid:    nodeUID,
		Uid:        uid,
		Name:       serviceName,
		ServiceUid: serviceID,
		ServerId:   serverId,
		Version:    version,
	}
}

func IsRetired(pid *PID) bool {
	return atomic.LoadInt32(&pid.State) == def.ServiceStatusRetired
}
