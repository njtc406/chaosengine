package errdef

import "errors"

// 定义系统错误

var (
	ModuleNotInitialized = errors.New("module not initialized")
	ModuleHadRegistered  = errors.New("module had registered")
	EventChannelIsFull   = errors.New("event channel is full")
)
