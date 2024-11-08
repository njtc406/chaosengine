// Package inf
// @Title  服务选择器
// @Description  根据条件选择服务
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package inf

import "github.com/njtc406/chaosengine/engine/actor"

type ISelector interface {
	// SelectBySvcUid 根据服务 UID 选择服务
	SelectBySvcUid(serviceUid string) IRpcInvoker // TODO 这里的选择器还稍微有点问题,使用uid无法完全定位服务,是现在的uid规则导致的

	// SelectByNodeUidAndSvcName 根据节点 UID 和服务名称选择服务
	SelectByNodeUidAndSvcName(nodeUID string, serviceName string) IRpcInvoker

	// SelectAllByName 根据服务名称选择所有服务
	SelectAllByName(serviceName string) IRpcInvoker

	// SelectRandomByName 随机选择某个服务名称的服务
	SelectRandomByName(serviceName string) IRpcInvoker

	SelectNameByRule(serviceName string, rule func(pid *actor.PID) bool) IRpcInvoker

	// SelectByRule 根据自定义规则选择服务
	SelectByRule(rule func(pid *actor.PID) bool) IRpcInvoker

	// SelectAll 选择所有服务(暂时没用)
	//SelectAll() IRpcInvoker
}
