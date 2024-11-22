// Package cluster
// @Title  集群模块
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package cluster

import (
	"github.com/njtc406/chaosengine/engine/cluster/config"
	"github.com/njtc406/chaosengine/engine/cluster/discovery"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/event"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/asynclib"
	"github.com/njtc406/chaosengine/engine/utils/log"
)

var cluster Cluster

func GetCluster() *Cluster {
	return &cluster
}

type Cluster struct {
	closed chan struct{}
	// 集群配置
	conf *config.Config

	// 服务发现
	discovery *discovery.Discovery

	// 节点列表
	endpoints *endpoints.EndpointManager

	// 事件
	eventProcessor inf.IProcessor
	eventChannel   chan inf.IEvent
}

func (c *Cluster) initConfig(confPath string) {
	c.conf = config.Init(confPath)
}

func (c *Cluster) Init(nodeId int32, confPath string) {
	// 加载集群配置
	c.initConfig(confPath)

	c.closed = make(chan struct{})
	c.eventChannel = make(chan inf.IEvent, 1024)
	c.eventProcessor = event.NewProcessor()
	c.eventProcessor.Init(c)

	c.endpoints = endpoints.GetEndpointManager()
	c.endpoints.Init(nodeId, c.conf.RPCServer.Addr, c.eventProcessor)

	c.discovery = discovery.NewDiscovery()
	if err := c.discovery.Init(c.conf.ETCDConf, c.eventProcessor); err != nil {
		log.SysLogger.Fatalf("init discovery error:%v", err)
	}
}

func (c *Cluster) Start() {
	c.discovery.Start()
	c.endpoints.Start()
	asynclib.Go(func() {
		c.run()
	})
}

func (c *Cluster) Close() {
	close(c.closed)
	c.endpoints.Stop()
	c.discovery.Close()
}

func (c *Cluster) PushEvent(ev inf.IEvent) error {
	select {
	case c.eventChannel <- ev:
	default:
		return errdef.EventChannelIsFull
	}
	return nil
}

func (c *Cluster) run() {
	for {
		select {
		case ev := <-c.eventChannel:
			if ev != nil {
				switch ev.GetType() {
				case event.SysEventETCDPut, event.SysEventETCDDel:
					e := ev.(*event.Event)
					c.eventProcessor.EventHandler(e)
					event.ReleaseEvent(e)
				}
			}
		case <-c.closed:
			log.SysLogger.Info("cluster closed")
			return
		}
	}
}
