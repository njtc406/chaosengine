// Package main
// @Title  title
// @Description  desc
// @Author  yr  2024/12/4
// @Update  yr  2024/12/4
package main

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/core"
	"github.com/njtc406/chaosengine/engine/core/node"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/services"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/example/msg"
	"time"
)

type Service4 struct {
	core.Service
}

func (s *Service4) OnInit() error {
	return nil
}

func (s *Service4) OnStart() error {
	return nil
}

func (s *Service4) OnRelease() {

}

func (s *Service4) RPCTest2() {
	//time.Sleep(time.Second * 3) // 模拟耗时操作
	log.SysLogger.Debugf("call %s func RPCTest2", s.GetName())
}

func (s *Service4) RPCSum(req *msg.Msg_Test_Req) *msg.Msg_Test_Resp {
	time.Sleep(time.Second * 2)
	return &msg.Msg_Test_Resp{
		Ret: req.A + req.B,
	}
}

func (s *Service4) RPCTestWithError(req *msg.Msg_Test_Req) (*msg.Msg_Test_Resp, error) {
	return nil, fmt.Errorf("rpc test")
}

func init() {
	services.SetService("Service4", func() inf.IService {
		return &Service4{}
	})
}

var version = "1.0"

func main() {
	node.Start(version, "./example/configs/node3")
}