// Package inf
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package inf

import "github.com/njtc406/chaosengine/engine/def"

type IEvent interface {
	GetType() def.EventType
}

type IChannel interface {
	PushEvent(ev IEvent) error
}
