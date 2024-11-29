package config

import (
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/spf13/viper"
	"time"
)

// TODO 这是第一版,后续可能会根据需求改进配置

const (
	Debug   = `debug`
	Release = `release`
)

type conf struct {
	NodeConf     *NodeConf       `binding:"required"` // 节点基础配置
	SystemLogger *log.LoggerConf `binding:"required"` // 系统日志
	ClusterConf  *ClusterConf    `binding:"required"` // 集群配置
	ServiceConf  *ServiceConf    `binding:"required"` // 服务配置
}

type NodeConf struct {
	NodeId           string        `binding:""`         // 节点ID(目前这个没用,节点id是节点启动的时候自动生成的)
	SystemStatus     string        `binding:"required"` // 系统状态(debug/release)
	PVCPath          string        `binding:""`         // 数据持久化目录(默认./data)
	PVPath           string        `binding:""`         // 缓存目录(默认./run)
	ProfilerInterval time.Duration `binding:""`         // 性能分析间隔
	AntsPoolSize     int           `binding:""`         // 线程池大小
}

type ClusterConf struct {
	OpenRemote     bool               `binding:""` // 是否开启远程配置(默认使用本地配置)
	ETCDConf       *ETCDConf          `binding:""` // etcd配置
	RPCServer      *RPCServer         `binding:""` // rpc服务配置
	RemoteType     string             `binding:""` // 远程服务类型(默认rpcx)
	DiscoveryType  string             `binding:""` // 服务发现类型(默认etcd)
	RemoteConfPath string             `binding:""` // 远程配置路径(开启了远程配置才会使用,且必须配置etcd)
	DiscoveryConf  *EtcdDiscoveryConf `binding:""` // 服务发现配置(目前先直接配置,后续会支持多种服务发现方式)
}

type ServiceConf struct {
	OpenRemote      bool                      `binding:""`         // 是否开启远程配置(默认使用本地)
	RemoteConfPath  string                    `binding:""`         // 远程配置路径(开启了远程配置才会使用,且必须配置etcd)
	StartServices   []*ServiceInitConf        `binding:"required"` // 启动服务列表(按照配置的顺序启动!!)
	ServicesConfMap map[string]*ServiceConfig `binding:"required"` // 服务配置
}

type ETCDConf struct {
	Endpoints   []string
	DialTimeout time.Duration // 默认3秒
	UserName    string
	Password    string
}

type RPCServer struct {
	Addr             string // rpc监听地址
	Protoc           string // 协议
	MaxRpcParamLen   int    // 最大rpc参数长度
	CompressBytesLen int    // 消息超过该值将进行压缩
}

type ServiceInitConf struct {
	ServiceName  string // 服务名称
	Type         string // 服务类型
	ServerId     int32  // 服务ID
	TimerSize    int    // 定时器数量
	MailBoxSize  int    // 事件队列数量
	GoroutineNum int32  // 协程数量
	RpcType      string // 远程调用方式(默认使用rpcx)
}

type ServiceConfig struct {
	ServiceName   string             // 服务名称
	ConfName      string             // 配置文件名称
	ConfPath      string             // 配置文件路径
	ConfType      string             // 配置文件类型
	CfgCreator    func() interface{} // 配置获取器(获取真实的配置格式)
	Cfg           interface{}        // 配置结构体
	DefaultSetFun func(*viper.Viper) // 默认配置函数
	OnChangeFun   func()             // 配置变化处理函数
}

type EtcdDiscoveryConf struct {
	Path string
	TTL  time.Duration
}
