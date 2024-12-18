// Package remote
// @Title  title
// @Description  desc
// @Author  yr  2024/12/18
// @Update  yr  2024/12/18
package remote

import (
	"github.com/njtc406/chaosengine/engine/cluster/endpoints/remote/pool"
	"github.com/njtc406/chaosengine/engine/config"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/asynclib"
	"github.com/njtc406/chaosengine/engine/utils/log"
)

func NewRemote() *Remote {
	return &Remote{}
}

type Remote struct {
	conf *config.RPCServer
	svr  inf.IRemoteServer
}

func (r *Remote) Init(conf *config.RPCServer, cliFactory inf.IRpcSenderFactory) *Remote {
	log.SysLogger.Debugf(">>>>>>>>>>>>>>>>cluster rpc server config: %+v", conf)
	r.conf = conf
	r.svr = pool.GetServer(conf.Type)
	if r.svr == nil {
		log.SysLogger.Panicf("rpc server type %s not support", conf.Type)
		return nil
	}
	r.svr.Init(cliFactory)
	return r
}

func (r *Remote) Serve() {
	_ = asynclib.Go(func() {
		if err := r.svr.Serve(r.conf); err != nil {
			log.SysLogger.Warnf("rpc serve stop: %v", err)
		}
	})
}

func (r *Remote) Close() {
	r.svr.Close()
}

func (r *Remote) GetAddress() string {
	return r.conf.Addr
}
