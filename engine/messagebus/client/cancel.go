// Package client
// @Title  title
// @Description  desc
// @Author  yr  2024/11/7
// @Update  yr  2024/11/7
package client

import (
	"github.com/njtc406/chaosengine/engine/inf"
)

type RpcCancel struct {
	Cli     inf.IClient
	CallSeq uint64
}

func (rc *RpcCancel) CancelRpc() {
	rc.Cli.Remove(rc.CallSeq)
}

func NewRpcCancel(cli inf.IClient, seq uint64) inf.CancelRpc {
	cancel := &RpcCancel{Cli: cli, CallSeq: seq}
	return cancel.CancelRpc
}
