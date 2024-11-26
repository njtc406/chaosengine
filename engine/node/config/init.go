package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosutil/validate"
	"github.com/spf13/viper"
	"os"
	"path"
	"strings"
	"time"
)

// TODO 增加一个从远程读取配置文件的功能

var (
	Conf    = new(conf)
	mapConf = make(map[string]*ConfInfo)
)

type ConfInfo struct {
	ServiceName      string                  // 服务名称
	ConfName         string                  // 配置文件名称
	ConfPath         string                  // 配置文件路径(这个路径是基于node.Start传入的路径)
	ConfType         string                  // 配置文件类型
	Cfg              interface{}             // 配置结构体
	DefaultSetFun    func(*viper.Viper)      // 默认配置函数
	OnChangeFun      func(in fsnotify.Event) // 配置变化处理函数
	OpenRemote       bool                    // 是否使用远程配置
	RemoteEndPoints  []string                // 远程配置文件地址
	RemotePrefixPath string                  // 远程配置监听目录前缀
}

func Init(nodeConfigPath string) {
	parseNodeConfig(nodeConfigPath)
	parseServiceConf(nodeConfigPath)
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

	parser.SetDefault(`AntsPoolSize`, 10000)

	parseSystemConfig(parser, Conf)
}

// parseServiceConf 解析服务配置文件
func parseServiceConf(confPath string) {
	for _, v := range mapConf {
		if v.Cfg == nil {
			continue
		}
		parser := viper.New()
		// 设置配置文件
		parser.SetConfigType(v.ConfType)                      // 配置文件类型
		parser.SetConfigName(v.ConfName)                      // 配置文件名称
		parser.AddConfigPath(path.Join(confPath, v.ConfPath)) // 配置文件路径
		executeDefaultSet(parser)
		if !parseSystemConfig(parser, v.Cfg) {
			if v.OpenRemote {
				// 从远程获取配置文件
			} else {
				log.SysLogger.Panic("配置文件不存在:" + v.ConfPath + v.ConfName)
			}
		}
		// TODO 监听配置文件变化
		if v.OnChangeFun != nil {
			listenConfChange(parser, v.OnChangeFun)
		}
	}
}

func listenConfChange(parser *viper.Viper, onChangeFun func(in fsnotify.Event)) {
	parser.WatchConfig()
	parser.OnConfigChange(onChangeFun)
}

func executeDefaultSet(parser *viper.Viper) {
	for _, v := range mapConf {
		if v.DefaultSetFun != nil {
			v.DefaultSetFun(parser)
		}
	}
}

func parseSystemConfig(parser *viper.Viper, c interface{}) bool {
	if err := parser.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return false
		}
	}
	if err := parser.Unmarshal(c); err != nil {
		panic(err)
	} else if err = validate.Struct(Conf); err != nil {
		panic(validate.TransError(err, validate.ZH))
	}

	return true
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

		if c.ServiceName == "" {
			panic(`service name is empty`)
		}

		if c.ConfName == "" {
			panic(`conf name is empty`)
		}

		if c.ConfType == "" {
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
