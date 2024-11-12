// Package inf
// @Title  服务选择器
// @Description  根据条件选择服务
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package inf

import "github.com/njtc406/chaosengine/engine/actor"

type ISelector interface {
	// SelectByPid 根据 PID 选择服务
	SelectByPid(sender, receiver *actor.PID) IBus

	// SelectBySvcUid 根据服务 UID 选择服务
	SelectBySvcUid(sender *actor.PID, serviceUid string) IBus // TODO 这里的选择器还稍微有点问题,使用uid无法完全定位服务,是现在的uid规则导致的

	// SelectByNodeUidAndSvcName 根据节点 UID 和服务名称选择服务
	SelectByNodeUidAndSvcName(sender *actor.PID, nodeUID string, serviceName string) IBus

	// SelectAllByName 根据服务名称选择所有服务
	SelectAllByName(sender *actor.PID, serviceName string) IBus

	// SelectRandomByName 随机选择某个服务名称的服务
	SelectRandomByName(sender *actor.PID, serviceName string) IBus

	SelectNameByRule(sender *actor.PID, serviceName string, rule func(pid *actor.PID) bool) IBus

	// SelectByRule 根据自定义规则选择服务
	SelectByRule(sender *actor.PID, rule func(pid *actor.PID) bool) IBus

	// SelectAll 选择所有服务(暂时没用)
	//SelectAll() IRpcInvoker
}
