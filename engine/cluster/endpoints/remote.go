// Package endpoints
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package endpoints

import (
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/smallnest/rpcx/server"
)

type Remote struct {
	address string // 服务监听地址
	nodeUID string // 节点唯一ID
	svr     *server.Server
}

func NewRemote(nodeUID, address string) *Remote {
	return &Remote{
		address: address,
		nodeUID: nodeUID,
	}
}

func (r *Remote) Init() {
	r.svr = server.NewServer()
}

func (r *Remote) Serve() error {
	// 注册rpc监听服务
	if err := r.svr.Register(new(RPCListener), ""); err != nil {
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

func (r *Remote) GetNodeUID() string {
	return r.nodeUID
}
