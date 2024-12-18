// Package gr
// @Title  grpc服务端监听器
// @Description  desc
// @Author  yr  2024/12/18
// @Update  yr  2024/12/18
package gr

import (
	"context"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/cluster/endpoints/remote/handler"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
)

type GrpcListener struct {
	actor.UnimplementedGrpcListenerServer
	cliFactory inf.IRpcSenderFactory
}

func (g *GrpcListener) RPCCall(_ context.Context, req *actor.Message) (*actor.RpcCallResponse, error) {
	log.SysLogger.Debugf("grpc call: %+v", req)
	return &actor.RpcCallResponse{}, handler.RpcMessageHandler(g.cliFactory, req)
}
