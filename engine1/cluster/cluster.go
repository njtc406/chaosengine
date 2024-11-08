// Package cluster
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/7/29 下午6:25
// @Update  yr  2024/7/29 下午6:25
package cluster

import (
	"github.com/njtc406/chaosengine/engine1/cluster/config"
	"github.com/njtc406/chaosengine/engine1/cluster/discovery"
	"github.com/njtc406/chaosengine/engine1/cluster/endpoints"
	"github.com/njtc406/chaosengine/engine1/errdef"
	"github.com/njtc406/chaosengine/engine1/event"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosengine/engine1/utils/asynclib"
	"time"
)

var cluster Cluster

type Cluster struct {
	closed bool
	// 集群配置
	conf *config.Config

	// 服务发现
	discovery *discovery.Discovery

	// 节点列表
	endpoints *endpoints.EndpointManager

	// 事件
	eventProcessor event.IProcessor
	eventChannel   chan event.IEvent

	callSet endpoints.CallSet // 调用集
}

func GetCluster() *Cluster {
	return &cluster
}

func (c *Cluster) initConfig(confPath string) {
	c.conf = config.Init(confPath)
}

func (c *Cluster) Init(nodeUID string, confPath string) {
	// 加载集群配置
	c.initConfig(confPath)

	c.eventChannel = make(chan event.IEvent, 1024)
	c.eventProcessor = event.NewProcessor()
	c.eventProcessor.Init(c)

	c.callSet.Init()
	c.endpoints = endpoints.GetEndpointManager()
	c.endpoints.Init(nodeUID, c.conf.RPCServer.Addr, c.eventProcessor, &c.callSet)

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
	c.closed = true
	c.endpoints.Stop()
	c.discovery.Close()
}

func (c *Cluster) PushEvent(ev event.IEvent) error {
	if c.closed {
		return nil
	}
	if len(c.eventChannel) == cap(c.eventChannel) {
		return errdef.EventChannelIsFull
	}
	c.eventChannel <- ev
	return nil
}

func (c *Cluster) run() {
	for !c.closed {
		select {
		case ev := <-c.eventChannel:
			if ev != nil {
				switch ev.GetEventType() {
				case event.SysEventETCDPut, event.SysEventETCDDel:
					e := ev.(*event.Event)
					c.eventProcessor.EventHandler(e)
					event.ReleaseEvent(e)
				}
			}
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
