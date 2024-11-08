package config

import (
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosutil/validate"
	"github.com/spf13/viper"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	Conf    = new(conf)
	mapConf = make(map[string]*ConfInfo)
)

type ConfInfo struct {
	ServiceName   string             // 服务名称
	ConfName      string             // 配置文件名称
	ConfPath      string             // 配置文件路径
	ConfType      string             // 配置文件类型
	Cfg           interface{}        // 配置结构体
	DefaultSetFun func(*viper.Viper) // 默认配置函数
}

func Init(nodeConfigPath string) {
	parseNodeConfig(nodeConfigPath)
	parseServiceConf()
	initDir()
}

func initDir() {
	// 创建缓存目录
	if err := os.MkdirAll(path.Join(Conf.CachePath, Conf.LogPath), 0644); err != nil {
		panic(err)
	}
}

// parseNodeConfig 解析node配置文件
func parseNodeConfig(nodeConfigPath string) {
	parser := viper.New()

	// 设置配置文件
	parser.SetConfigType(`yaml`)         // 配置文件类型
	parser.SetConfigName(`node`)         // 配置文件名称
	parser.AddConfigPath(nodeConfigPath) // 配置文件路径

	// 默认配置
	parser.SetDefault(`SystemStatus`, Debug)
	parser.SetDefault(`CachePath`, `./run`)
	parser.SetDefault(`LogPath`, `logs`)

	// 日志默认配置
	parser.SetDefault(`SystemLogger`, &log.LoggerConf{
		Name:         "error.log",
		Level:        "error",
		Caller:       true,
		FullCaller:   false,
		Color:        false,
		MaxAge:       time.Hour * 24 * 15,
		RotationTime: time.Hour * 24,
	})

	//parser.SetDefault(`SystemLogger.Module`, `error.log`)
	//parser.SetDefault(`SystemLogger.Level`, `error`)
	//parser.SetDefault(`SystemLogger.Caller`, true)
	//parser.SetDefault(`SystemLogger.FullCaller`, false)
	//parser.SetDefault(`SystemLogger.Color`, false)
	//parser.SetDefault(`SystemLogger.MaxAge`, time.Hour*24*15)    // 文件保留时间,默认15天
	//parser.SetDefault(`SystemLogger.RotationTime`, time.Hour*24) // 文件切割时间,默认1天

	parseSystemConfig(parser, Conf)

	// 修正配置
	fixConf()
}

func fixConf() {
	// 如果在环境变量中设置了NODE_ID,则使用环境变量的值
	nodeIdStr := os.Getenv("NODE_ID")
	if 0 < len(nodeIdStr) {
		n, err := strconv.Atoi(nodeIdStr)
		if err != nil {
			log.SysLogger.Panic("节点id必须是一个数字")
		}
		Conf.NodeConf.ID = int32(n)
	}

	if 0 == Conf.NodeConf.ID {
		log.SysLogger.Panic("无法获取节点id,请在配置文件中配置或者在环境变量中设置NODE_ID")
	}

	nodeType := os.Getenv("NODE_TYPE")
	if 0 < len(nodeType) {
		Conf.NodeConf.Type = nodeType
	}
	if 0 == len(Conf.NodeConf.Type) {
		log.SysLogger.Panic("无法获取节点类型,请在环境变量中设置NODE_TYPE")
	}
}

// parseServiceConf 解析服务配置文件
func parseServiceConf() {
	for _, v := range mapConf {
		if v.Cfg == nil {
			continue
		}
		parser := viper.New()
		// 设置配置文件
		parser.SetConfigType(v.ConfType) // 配置文件类型
		parser.SetConfigName(v.ConfName) // 配置文件名称
		parser.AddConfigPath(v.ConfPath) // 配置文件路径
		executeDefaultSet(parser)
		parseSystemConfig(parser, v.Cfg)
	}
}

func executeDefaultSet(parser *viper.Viper) {
	for _, v := range mapConf {
		if v.DefaultSetFun != nil {
			v.DefaultSetFun(parser)
		}
	}
}

func parseSystemConfig(parser *viper.Viper, c interface{}) {
	if err := parser.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := parser.Unmarshal(c); err != nil {
		panic(err)
	} else if err = validate.Struct(Conf); err != nil {
		panic(validate.TransError(err, validate.ZH))
	}
}

// IsDebug 系统状态
func IsDebug() bool {
	return Conf.SystemStatus == Debug
}

func SetStatus(status string) {
	stat := strings.ToLower(status)
	if stat != Debug && stat != Release {
		return
	}

	Conf.SystemStatus = stat
}

func RegisterConf(cfgs ...*ConfInfo) {
	for _, c := range cfgs {
		if c == nil {
			panic(`register conf is nil`)
		}

		if c.ServiceName == `` {
			panic(`service name is empty`)
		}

		if c.ConfName == `` {
			panic(`conf name is empty`)
		}

		if c.ConfPath == `` {
			panic(`conf path is empty`)
		}

		if c.ConfType == `` {
			panic(`conf type is empty`)
		}

		if c.Cfg == nil {
			panic(`cfg struct is nil`)
		}

		mapConf[c.ServiceName] = c
	}
}

func GetConf(serviceName string) interface{} {
	if cf, ok := mapConf[serviceName]; !ok {
		return nil
	} else {
		return cf.Cfg
	}
}
