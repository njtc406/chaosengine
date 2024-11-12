// Package rpc
// @Title  rpc调用器
// @Description  rpc调用器,实现了IRpcInvoker接口
// @Author  yr  2024/11/11
// @Update  yr  2024/11/11
package rpc

import (
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
	"time"
)

// TODO 这里目前有点问题,我去掉了在这里选择节点,那么这里就只剩下了封装信封的逻辑,但是信封封装,又需要一个sender和receiver信息,这里无法获取到
// 本来我是想要按照mysql那种方式,直接一个链式调用,所以这一块应该被封装到其他地方去,放在这里会让整个系统出现一个情况,就是如果我的逻辑不是在一个service中
// 那么他无法使用rpc调用其他服务的资源,当然这种可能不能是call类型的,但是可以是send和cast

func (h *Handler) Call(method string, in, out interface{}) error {

	return nil
}

func (h *Handler) CallWithTimeout(method string, timeout time.Duration, in, out interface{}) error {

	return nil
}

func (h *Handler) AsyncCall(method string, timeout time.Duration, callbacks []actor.CompletionFunc, in, out interface{}) (inf.CancelRpc, error) {

	return inf.EmptyCancelRpc, nil
}

func (h *Handler) Send(method string, in interface{}) error {

	return nil
}

func (h *Handler) Cast(method string, in interface{}) error {

	return nil
}
