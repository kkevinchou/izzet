package clientsystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type PingSystem struct {
	app         App
	accumulator int64
}

func NewPingSystem(app App) *PingSystem {
	return &PingSystem{app: app}
}

func (s *PingSystem) Name() string {
	return "PingSystem"
}

func (s *PingSystem) Update(delta time.Duration, world systems.GameWorld) {
	s.accumulator += delta.Milliseconds()
	if s.accumulator <= 2000 {
		return
	}
	s.accumulator -= 500
	pingMessage := network.PingMessage{UnixTime: time.Now().UnixNano()}
	s.app.Client().Send(pingMessage, s.app.CommandFrame())
}
