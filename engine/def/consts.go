// Package def
// @Title  常量定义
// @Description  desc
// @Author  yr  2024/11/6
// @Update  yr  2024/11/6
package def

import "time"

const (
	DefaultRpcConnNum           = 1
	DefaultRpcLenMsgLen         = 4
	DefaultRpcMinMsgLen         = 2
	DefaultMaxCheckCallRpcCount = 1000
	DefaultMaxPendingWriteNum   = 1000000

	DefaultConnectInterval             = 2 * time.Second
	DefaultCheckRpcCallTimeoutInterval = 1 * time.Second
	DefaultRpcTimeout                  = 3 * time.Second
)

const (
	ServiceStatusNormal int32 = iota
	ServiceStatusRetired
)
