// Package node
// 模块名: 程序入口
// 功能描述: 用于提供程序入口
// 作者:  yr  2024/1/10 0010 23:43
// 最后更新:  yr  2024/1/10 0010 23:43
package node

import (
	"github.com/njtc406/chaosengine/engine/cluster"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/monitor"
	"github.com/njtc406/chaosengine/engine/node/config"
	"github.com/njtc406/chaosengine/engine/profiler"
	"github.com/njtc406/chaosengine/engine/services"
	serviceConf "github.com/njtc406/chaosengine/engine/services/config"
	"github.com/njtc406/chaosengine/engine/utils/asynclib"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/pid"
	"github.com/njtc406/chaosengine/engine/utils/timer"
	"github.com/njtc406/chaosengine/engine/utils/version"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	exitCh           = make(chan os.Signal)
	ID               int32
	Type             string
	baseSetupService []inf.IService
	preSetupService  []inf.IService
	hooks            []func()
)

func init() {
	// 注册退出信号
	signal.Notify(exitCh, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
}

func GetNodeUID() string {
	return config.Conf.NodeConf.GetNodeUid()
}

func GetNodeID() int32 {
	return config.Conf.NodeConf.ID
}

func SetStartHook(f ...func()) {
	hooks = append(hooks, f...)
}

func Start(v string, confPath string) {
	// 打印版本信息
	version.EchoVersion(v)
	// TODO: 这里后面如果加入集群,那么需要从集群中获取节点配置
	// 初始化节点配置
	config.Init(confPath)

	// 初始化日志
	log.Init(config.Conf.GetSystemLoggerFileName(), config.Conf.SystemLogger, config.IsDebug())

	// 启动线程池
	asynclib.InitAntsPool(config.Conf.AntsPoolSize)

	// 初始化等待队列,并启动监听
	monitor.GetRpcMonitor().Init().Start()

	// 初始化集群设置
	cluster.GetCluster().Init(config.Conf.NodeConf.ID, config.Conf.NodeConf.Type, confPath)
	// 启动集群管理器
	cluster.GetCluster().Start()

	// 记录pid
	pid.RecordPID(config.Conf.CachePath, ID, Type)
	defer pid.DeletePID(config.Conf.CachePath, ID, Type)

	// 启动timer
	timer.StartTimer(10*time.Millisecond, 1000000)

	// 执行钩子
	for _, f := range hooks {
		f()
	}

	// 加载服务配置
	servicesConfig := serviceConf.Init(confPath)

	// 执行节点初始化
	initNode(servicesConfig)

	// 启动服务
	services.Start()

	running := true
	pProfilerTicker := new(time.Ticker)
	defer pProfilerTicker.Stop()
	if config.Conf.ProfilerInterval > 0 {
		pProfilerTicker = time.NewTicker(config.Conf.ProfilerInterval)
	}

	for running {
		select {
		case sig := <-exitCh:
			log.SysLogger.Infof("-------------->>received the signal: %v", sig)
			running = false
		case <-pProfilerTicker.C:
			profiler.Report()
		}
	}
	log.SysLogger.Info("==================>>begin stop modules<<==================")
	services.StopAll()
	cluster.GetCluster().Close()
	monitor.GetRpcMonitor().Stop()
	log.SysLogger.Info("server stopped, program exited...")
	log.Close()
}

func initNode(serviceConfig map[string]*def.ServiceInitConf) {
	log.SysLogger.Debugf("serviceConfig: %+v", serviceConfig)
	// 先加载基础服务
	for _, s := range baseSetupService {
		s.Init(s, serviceConfig[s.GetName()], config.GetConf(s.GetName()))
		services.Setup(s)
	}

	// 顺序加载服务
	for _, s := range preSetupService {
		s.Init(s, serviceConfig[s.GetName()], config.GetConf(s.GetName()))
		services.Setup(s)
	}

	services.Init()
}

func Setup(s ...inf.IService) {
	for _, sv := range s {
		sv.OnSetup(sv)
		preSetupService = append(preSetupService, sv)
	}
}

// SetupBase 设置基础服务
func SetupBase(s ...inf.IService) {
	for _, sv := range s {
		sv.OnSetup(sv)
		baseSetupService = append(baseSetupService, sv)
	}
}
