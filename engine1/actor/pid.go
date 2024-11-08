// Package actor
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/9/4 下午5:53
// @Update  yr  2024/9/4 下午5:53
package actor

func NewPID(nodeUID, address, serviceID, serviceName string) *PID {
	if serviceID == "" {
		serviceID = nodeUID
	}
	uid := serviceID + "@" + serviceName
	return &PID{
		NodeUID: nodeUID,
		Address: address,
		Uid:     uid,
		Name:    serviceName,
		Id:      serviceID,
		State:   1,
	}
}

func IsRetired(pid *PID) bool {
	return pid.GetState() == 0
}
