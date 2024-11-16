// Package main
// @Title  title
// @Description  desc
// @Author  yr  2024/11/15
// @Update  yr  2024/11/15
package main

import "github.com/njtc406/chaosengine/engine/node"

var version = "1.0"

const configPath = "./configs/game"

func main() {
	node.Start(version, configPath)
}
