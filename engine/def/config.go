// Package def
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package def

import "github.com/spf13/viper"

type ServiceInitConf struct {
	Name         string // 服务名称
	ServerId     int32  // 服务ID
	TimerSize    int    // 定时器数量
	MailBoxSize  int    // 事件队列数量
	GoroutineNum int32  // 协程数量
}

type ServiceConfig struct {
	ServiceName   string             // 服务名称
	ConfName      string             // 配置文件名称
	ConfPath      string             // 配置文件路径
	ConfType      string             // 配置文件类型
	Cfg           interface{}        // 配置结构体
	DefaultSetFun func(*viper.Viper) // 默认配置函数
}
