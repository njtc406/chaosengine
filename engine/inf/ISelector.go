// Package inf
// @Title  服务选择器
// @Description  根据条件选择服务
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package inf

import "github.com/njtc406/chaosengine/engine/actor"

type ISelector interface {
	// Select 选择服务
	Select(sender *actor.PID, serverId int32, serviceId, serviceName string) IBus

	// SelectByRule 根据自定义规则选择服务
	SelectByRule(sender *actor.PID, rule func(pid *actor.PID) bool) IBus

	SelectByPid(sender, receiver *actor.PID) IBus

	SelectByNodeType(sender *actor.PID, nodeType, serviceName string) IBus

	// SelectOneByType 根据选择类型选择服务(暂时不做)
	//SelectOneByType(sender *actor.PID, selectType int32, rule func(pid *actor.PID)) IBus
}
