// Package inf
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/7/29 下午4:47
// @Update  yr  2024/7/29 下午4:47
package inf

import "github.com/njtc406/chaosengine/engine/actor"

// TODO 在sender加入一个create方法,根据不同的pid参数来创建不同的sender,这样可以直接扩展出不同方式的sender

type IRpcSender interface {
	IRpcHandler
	Close()
	// SendRequest 发送请求消息
	SendRequest(envelope IEnvelope) error
	SendRequestAndRelease(envelope IEnvelope) error
	SendResponse(envelope IEnvelope) error
}

type IRpcSenderFactory interface {
	GetClient(pid *actor.PID) IRpcSender
}
