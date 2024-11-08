// Package event
// @Title  事件类型
// @Description  事件类型
// @Author  yr  2024/7/19 下午3:40
// @Update  yr  2024/7/19 下午3:40
package event

type EventType int

const (

	// 基础事件 -1000以上 系统事件 -1 到 -999  用户事件 1 - 99
	SysEventWebsocket = -5

	SysEventRpc        = -1001 // 远程调用事件
	SysEventReply      = -1002 // 远程调用回复事件
	SysEventClientMsg  = -1003 // 客户端消息事件
	SysEventETCDPut    = -1004 // etcd 存储事件
	SysEventETCDDel    = -1005 // etcd 删除事件
	SysEventServiceReg = -1006 // 服务注册事件
	SysEventServiceDis = -1007 // 服务注销事件

	SysEventNodeConn = -1010 // 节点连接事件
	SysEventNatsConn = -1011 // nats 连接事件

	MaxType = -1000
)
