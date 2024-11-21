// Package inf
// @Title  title
// @Description  desc
// @Author  yr  2024/11/14
// @Update  yr  2024/11/14
package inf

import (
	"context"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/dto"
)

type IRpcListener interface {
	RPCCall(ctx context.Context, req *actor.Message, res *dto.RPCResponse) error
}
