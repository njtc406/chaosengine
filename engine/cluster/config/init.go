package config

import (
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/spf13/viper"
)

var discoveryConfMap = map[string]interface{}{
	"etcd": &EtcdConf{},
}

func Register(name string, conf interface{}) {
	discoveryConfMap[name] = conf
}

func Init(configPath string) *Config {
	return parseClusterConfig(configPath)
}

// parseClusterConfig 动态解析配置
func parseClusterConfig(nodeConfigPath string) *Config {
	parser := viper.New()

	// 设置配置文件
	parser.SetConfigType("yaml")         // 配置文件类型
	parser.SetConfigName("cluster")      // 配置文件名称
	parser.AddConfigPath(nodeConfigPath) // 配置文件路径

	conf := new(Config)
	conf.DiscoveryConf = make(map[string]interface{})

	// 读取配置文件
	if err := parser.ReadInConfig(); err != nil {
		log.SysLogger.Panicf("Error reading config file, %s", err)
	}

	// 解析 RPCServer 配置
	if err := parser.UnmarshalKey("RPCServer", &conf.RPCServer); err != nil {
		log.SysLogger.Panicf("Error unmarshaling RPCServer config, %s", err)
	}

	// 解析 DiscoveryUse 配置
	if err := parser.UnmarshalKey("DiscoveryUse", &conf.DiscoveryUse); err != nil {
		log.SysLogger.Panicf("Error unmarshaling DiscoveryUse config, %s", err)
	}

	// 根据 DiscoveryUse 选择具体的配置
	if conf.DiscoveryUse != "" {
		// 如果配置了 DiscoveryUse 为 "etcd"，则解析对应的配置
		if discoveryConfObj, ok := discoveryConfMap[conf.DiscoveryUse]; ok {
			if err := parser.UnmarshalKey("DiscoveryConf."+conf.DiscoveryUse, discoveryConfObj); err != nil {
				log.SysLogger.Panicf("Error unmarshaling DiscoveryConf for %s, %s", conf.DiscoveryUse, err)
			}
			conf.DiscoveryConf[conf.DiscoveryUse] = discoveryConfObj
		} else {
			log.SysLogger.Panicf("Unknown discovery type %s", conf.DiscoveryUse)
		}
	}

	//log.SysLogger.Debugf("config: %+v", conf)

	return conf
}
