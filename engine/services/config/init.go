package config

import (
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosutil/validate"
	"github.com/spf13/viper"
)

func Init(configPath string) map[string]*def.ServiceInitConf {
	return parseClusterConfig(configPath)
}

// parseNodeConfig 解析node配置文件
func parseClusterConfig(nodeConfigPath string) map[string]*def.ServiceInitConf {
	parser := viper.New()

	// 设置配置文件
	parser.SetConfigType(`yaml`)         // 配置文件类型
	parser.SetConfigName(`service`)      // 配置文件名称
	parser.AddConfigPath(nodeConfigPath) // 配置文件路径

	conf := new(config)
	parseSystemConfig(parser, conf)

	var ret = make(map[string]*def.ServiceInitConf)
	for _, v := range conf.Services {
		ret[v.Name] = v
	}

	return ret
}

func parseSystemConfig(parser *viper.Viper, c interface{}) {
	if err := parser.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := parser.Unmarshal(c); err != nil {
		panic(err)
	} else if err = validate.Struct(c); err != nil {
		panic(validate.TransError(err, validate.ZH))
	}
}
