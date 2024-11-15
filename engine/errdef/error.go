package errdef

import (
	"errors"
)

// 定义系统错误

var (
	ModuleNotInitialized = errors.New("module not initialized")
	ModuleHadRegistered  = errors.New("module had registered")
	EventChannelIsFull   = errors.New("event channel is full")
	RPCCallTimeout       = errors.New("RPC call timeout")
	ServiceNotFound      = errors.New("service not found")
	RPCCallFailed        = errors.New("RPC call failed")
	ParamNotMatch        = errors.New("param not match")
	OutputParamNotMatch  = errors.New("output param not match")
	MethodNotFound       = errors.New("method not found")
	RPCHadClosed         = errors.New("RPC had closed")
	MsgSerializeFailed   = errors.New("message serialize failed")
)
