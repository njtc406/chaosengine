// Package remote
// @Title  远程rpc服务器
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/9/3 下午4:26
// @Update  yr  2024/9/3 下午4:26
package remote

import (
	"github.com/google/uuid"
	"github.com/njtc406/chaosengine/engine/config"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/smallnest/rpcx/server"
)

var remoteMap = map[string]inf.IRemoteServer{
	def.DefaultRpcTypeRpcx: NewRemote(),
	// TODO 支持grpc
}

func GetRemote(tp string) inf.IRemoteServer {
	return remoteMap[tp]
}

type DefaultRemote struct {
	nodeUid  string // 节点唯一ID
	listener inf.IRpcListener
	conf     *config.RPCServer
	svr      *server.Server
}

func NewRemote() *DefaultRemote {
	return &DefaultRemote{}
}

func (r *DefaultRemote) Init(conf *config.RPCServer, cliFactory inf.IRpcSenderFactory) {
	r.conf = conf
	if r.nodeUid == "" {
		r.nodeUid = uuid.NewString()
	}
	r.listener = NewRPCListener(cliFactory)
	r.svr = server.NewServer()
}

func (r *DefaultRemote) Serve() error {
	// 注册rpc监听服务
	if err := r.svr.RegisterName("RpcListener", r.listener, ""); err != nil {
		return err
	}

	go func() {
		if err := r.svr.Serve(r.conf.Protoc, r.conf.Addr); err != nil {
			log.SysLogger.Warnf("rpc serve stop: %v", err)
		}
	}()

	return nil
}

func (r *DefaultRemote) Close() {
	if err := r.svr.Close(); err != nil {
		log.SysLogger.Errorf("rpc server close error: %v", err)
	}
}

func (r *DefaultRemote) GetAddress() string {
	return r.conf.Addr
}

func (r *DefaultRemote) GetNodeUid() string {
	return r.nodeUid
}
