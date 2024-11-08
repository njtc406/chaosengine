// Package inf
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package inf

type EventType int

type IEvent interface {
	GetType() EventType
}

type IChannel interface {
	PushEvent(ev IEvent) error
}
