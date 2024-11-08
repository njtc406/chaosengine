// Package config
// @Title
// @Description  config
// @Author sly 2024/9/12
// @Created sly 2024/9/12
package config

import "github.com/njtc406/chaosengine/engine1/network"

type WSService struct {
	ServerConf *network.WSServer `binding:"required"` // mysql配置
}
