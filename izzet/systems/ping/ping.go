package ping

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/knetwork"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/systems/base"
)

type World interface {
	GetPlayer() *player.Player
}

type PingSystem struct {
	*base.BaseSystem
	world   World
	enabled bool
}

func NewPingSystem(world World) *PingSystem {
	return &PingSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
		enabled:    true,
	}
}

func (s *PingSystem) Update(delta time.Duration) {
	if !s.enabled {
		return
	}

	player := s.world.GetPlayer()

	pingMessage := &knetwork.PingMessage{
		SendTime: time.Now(),
	}

	err := player.Client.SendMessage(knetwork.MessageTypePing, pingMessage)
	if err != nil {
		fmt.Printf("error sending ping message: %s\nshutting down ping system\n", err)
		s.enabled = false
	}
}

func (s *PingSystem) Name() string {
	return "PingSystem"
}
