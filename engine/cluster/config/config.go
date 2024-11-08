// Package config
// @Title  集群配置
// @Description  集群配置
// @Author  yr  2024/8/30 下午6:54
// @Update  yr  2024/8/30 下午6:54
package config

import (
	"time"
)

type Config struct {
	RPCServer *RPCServer `binding:"required"` // rpc服务配置
	ETCDConf  *ETCDConf
}

type RPCServer struct {
	Addr             string // rpc监听地址
	Protoc           string // 协议
	MaxRpcParamLen   int    // 最大rpc参数长度
	CompressBytesLen int    // 消息超过该值将进行压缩
}

type ETCDConf struct {
	EtcdEndPoints []string
	DialTimeout   time.Duration
	DiscoveryPath string
	UserName      string
	Password      string
	TTL           time.Duration // 租约过期时间
}

func (ec *ETCDConf) GetFullPath(uid string) (path string) {
	return ec.DiscoveryPath + "/" + uid
}
