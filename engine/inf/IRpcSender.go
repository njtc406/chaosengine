// Package inf
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/7/29 下午4:47
// @Update  yr  2024/7/29 下午4:47
package inf

type IRpcSender interface {
	IRpcHandler
	Close()
	// SendRequest 发送请求消息
	SendRequest(envelope IEnvelope) error
	SendRequestAndRelease(envelope IEnvelope) error
	SendResponse(envelope IEnvelope) error
}
