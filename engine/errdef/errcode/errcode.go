// Package errcode
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/8/30 下午3:27
// @Update  yr  2024/8/30 下午3:27
package errcode

const (
	Ok = 0 // 正常

	RpcErr                                 = 100 // rpc 错误
	UnmarshalErr                           = 101 // 解码错误
	RpcCallTimeout                         = 102 // rpc 调用超时
	RpcCallRequestIDExist                  = 103 // rpc 调用请求ID已存在
	RpcCallResponseErr                     = 104 // rpc 调用返回错误
	ServiceNotExist                        = 201 // 服务不存在
	ServiceNameInvalid                     = 202 // 服务名不合法
	ServiceModuleRepeats                   = 203 // 服务模块重复
	ServiceModuleNotInit                   = 204 // 服务模块未初始化
	ServiceSyncCallbackFunError            = 205 // 服务同步回调函数错误
	ServiceSyncCallbackFunParamsError      = 206 // 服务同步回调函数参数错误
	ServiceSyncCallbackFunParamsCountError = 207 // 服务同步回调函数参数数量错误
	ServiceMethodNotExist                  = 208 // 服务方法不存在
	ServiceMethodParamsError               = 209 // 服务方法参数错误
	ServiceMethodReturnCountError          = 210 // 服务方法返回值数量错误
	ServiceEventChannelFull                = 211 // 服务事件通道已满
	ServiceMethodError                     = 212 // 服务方法错误
	ServerError                            = 301 // 服务端错误
	DBRedisError                           = 401 // 数据库或redis错误
	DBError                                = 402 // 数据库错误
	ParamsInvalid                          = 501 // 参数错误
)
