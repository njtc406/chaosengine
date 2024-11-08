// Package services
// @Title  服务配置
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package services

import "github.com/njtc406/chaosengine/engine/def"

var mapServiceConf map[string]*def.ServiceConfig

func init() {
	mapServiceConf = make(map[string]*def.ServiceConfig)
}

// RegisterConf 注册服务配置
func RegisterConf(cfgs ...*def.ServiceConfig) {
	for _, cfg := range cfgs {
		mapServiceConf[cfg.ServiceName] = cfg
	}
}

// GetServiceConfMap 获取服务配置
func GetServiceConfMap() map[string]*def.ServiceConfig {
	return mapServiceConf
}

// GetConfByName 根据服务名获取服务配置
func GetConfByName(serviceName string) *def.ServiceConfig {
	return mapServiceConf[serviceName]
}
