// Package inf
// @Title  title
// @Description  desc
// @Author  yr  2024/11/26
// @Update  yr  2024/11/26
package inf

import "github.com/njtc406/chaosengine/engine/config"

type IRemoteServer interface {
	Init(conf *config.RPCServer, cliFactory IRpcSenderFactory)
	Serve() error
	Close()
	GetAddress() string
	GetNodeUid() string
}
