// Package remote
// @Title  远程rpc服务器
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package remote

import (
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/smallnest/rpcx/server"
)

type Remote struct {
	address  string // 服务监听地址
	nodeId   int32  // 节点唯一ID
	nodeType string // 节点类型
	listener inf.IRpcListener
	svr      *server.Server
}

func NewRemote(nodeId int32, nodeType, address string, listener inf.IRpcListener) *Remote {
	return &Remote{
		address:  address,
		nodeId:   nodeId,
		nodeType: nodeType,
		listener: listener,
	}
}

func (r *Remote) Init() {
	r.svr = server.NewServer()
}

func (r *Remote) Serve() error {
	// 注册rpc监听服务
	if err := r.svr.Register(r.listener, ""); err != nil {
		return err
	}

	go func() {
		if err := r.svr.Serve("tcp", r.address); err != nil {
			log.SysLogger.Errorf("rpc serve stop: %v", err)
		}
	}()

	return nil
}

func (r *Remote) GetAddress() string {
	return r.address
}

func (r *Remote) GetNodeId() int32 {
	return r.nodeId
}

func (r *Remote) GetNodeType() string {
	return r.nodeType
}
