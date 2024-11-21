package config

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"path"
	"time"
)

const (
	Debug   = `debug`
	Release = `release`
)

type conf struct {
	SystemStatus     string          `binding:"oneof=debug release"` // 系统状态(debug/release)
	CachePath        string          `binding:"required"`            // 系统缓存目录(默认./run)
	LogPath          string          `binding:"omitempty"`           // 日志存储目录(默认./run/logs)
	SystemLogger     *log.LoggerConf `binding:"required"`            // 系统日志
	NodeConf         *node           `binding:"omitempty"`           // 节点配置(允许为空, 可以从环境变量中获取)
	ProfilerInterval time.Duration   `binding:""`                    // 性能分析间隔
	AntsPoolSize     int             `binding:""`                    // 线程池大小
}

func (s *conf) GetSystemLoggerFileName() string {
	if s.SystemLogger.Name == "" {
		return path.Join(s.CachePath, s.LogPath, "/error.log")
	}

	return path.Join(s.CachePath, s.LogPath, s.SystemLogger.Name)
}

type node struct {
	ID   int32  `json:"id"`   // 节点id
	Type string `json:"type"` // 节点类型
}

func (n *node) GetNodeUid() string {
	return fmt.Sprintf("%s_%d", n.Type, n.ID)
}
