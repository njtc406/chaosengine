// Package client
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package client

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/def"
	"github.com/njtc406/chaosengine/engine/inf"
)

type Handler struct {
	inf.IChannel
	inf.IRpcHandler
	internalFuncMap map[string]*def.MethodInfo
	rpcFuncMap      map[string]*def.MethodInfo
}

func NewHandler(handler inf.IRpcHandler, rpcChannel inf.IChannel) *Handler {
	return &Handler{
		IChannel:        rpcChannel,
		IRpcHandler:     handler,
		internalFuncMap: make(map[string]*def.MethodInfo),
		rpcFuncMap:      make(map[string]*def.MethodInfo),
	}
}

func (h *Handler) Init() *Handler {

	return h
}

func (h *Handler) GetName() string {
	return h.IRpcHandler.GetName()
}

func (h *Handler) GetPID() *actor.PID {
	return h.IRpcHandler.GetPID()
}

func (h *Handler) GetRpcHandler() inf.IRpcHandler {
	return h.IRpcHandler
}

func (h *Handler) HandleRequest(envelope *actor.MsgEnvelope) {

}

func (h *Handler) HandleResponse(envelope *actor.MsgEnvelope) {

}

func (h *Handler) IsPrivate() bool {
	return len(h.rpcFuncMap) == 0
}
