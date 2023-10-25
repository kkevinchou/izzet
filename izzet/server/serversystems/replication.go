package serversystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type App interface {
	GetPlayers() map[int]network.Player
}

type ReplicationSystem struct {
	app App
}

func New(app App) *ReplicationSystem {
	return &ReplicationSystem{app: app}
}

func (s *ReplicationSystem) Update(delta time.Duration, world systems.GameWorld) {
	// entityIDs := []int{}
	// for i, entity := range world.Entities() {
	// 	if i > 3 {
	// 		break
	// 	}
	// 	entityIDs = append(entityIDs, entity.GetID())
	// }

	// for _, player := range s.app.GetPlayers() {
	// 	encoder := json.NewEncoder(player.Connection)
	// 	err := encoder.Encode(entityIDs)
	// 	if err != nil {
	// 		fmt.Println("error with writing message: %w", err)
	// 	}
	// }
}
