// Package login
// 模块名: 模块名
// 功能描述: 描述
// 作者:  yr  2024/11/16 0016 14:45
// 最后更新:  yr  2024/11/16 0016 14:45
package main

import (
	"github.com/njtc406/chaosengine/engine/core/node"
	_ "github.com/njtc406/chaosengine/engine/sysService/dbservice"
	_ "github.com/njtc406/chaosengine/example/login/internal/service/login"
)

var version = "1.0"
var configPath = "./example/configs/node1"

func main() {
	node.Start(version, configPath)
}
