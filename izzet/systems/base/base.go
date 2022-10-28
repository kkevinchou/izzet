package base

import "github.com/kkevinchou/izzet/izzet/events"

type BaseSystem struct {
}

func NewBaseSystem() *BaseSystem {
	return &BaseSystem{}
}

func (b *BaseSystem) Observe(event events.Event) {
}
