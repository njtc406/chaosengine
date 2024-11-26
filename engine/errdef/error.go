package errdef

import (
	"errors"
)

// 定义系统错误

var (
	ModuleNotInitialized    = errors.New("module not initialized")
	ModuleHadRegistered     = errors.New("module had registered")
	EventChannelIsFull      = errors.New("event channel is full")
	RPCCallTimeout          = errors.New("RPC call timeout")
	ServiceNotFound         = errors.New("service not found")
	ServiceIsRunning        = errors.New("service is running")
	RPCCallFailed           = errors.New("RPC call failed")
	ParamNotMatch           = errors.New("param not match")
	InputParamCantUseStruct = errors.New("input param can't use struct, must be ptr")
	OutputParamNotMatch     = errors.New("output param not match")
	MethodNotFound          = errors.New("method not found")
	RPCHadClosed            = errors.New("RPC had closed")
	MsgSerializeFailed      = errors.New("message serialize failed")
	TokenExpired            = errors.New("token expired")
	TokenInvalid            = errors.New("token invalid")
	JsonMarshalFailed       = errors.New("json marshal failed")
	JsonUnmarshalFailed     = errors.New("json unmarshal failed")
	HttpCreateRequestFailed = errors.New("http create request failed")
	HttpRequestFailed       = errors.New("http request failed")
	HttpReadResponseFailed  = errors.New("http read response failed")
	ServiceIsUnavailable    = errors.New("service is unavailable")
	DiscoveryConfNotFound   = errors.New("discovery conf not found")
)
