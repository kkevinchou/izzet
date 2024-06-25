package clientsystems

import (
	"fmt"
	"math"
	"time"

	"github.com/kkevinchou/izzet/app/entities"
	"github.com/kkevinchou/izzet/app/systems"
)

type PositionSyncSystem struct {
	app App
}

func NewPositionSyncSystem(app App) *PositionSyncSystem {
	return &PositionSyncSystem{app: app}
}

func (s *PositionSyncSystem) Name() string {
	return "PositionSyncSystem"
}

// tion easeInOutExpo(x: number): number {
// return x === 0
//   ? 0
//   : x === 1
//   ? 1
//   : x < 0.5 ? Math.pow(2, 20 * x - 10) / 2
//   : (2 - Math.pow(2, -20 * x + 10)) / 2;
// }

// x between 0 and 1
func easeInOutExpo(x float64) float64 {
	if x == 0 || x == 1 {
		return x
	} else if x > 1 {
		return 1
	}

	if x < 0.5 {
		return math.Pow(2, 20*x-10) / 2
	}

	return (2 - math.Pow(2, -20*x+10)) / 2
}

func (s *PositionSyncSystem) Update(delta time.Duration, world systems.GameWorld) {
	player := s.app.GetPlayerEntity()
	if player.PositionSync.Active {
		x := float64(time.Since(player.PositionSync.StartTime).Milliseconds()) / 1000
		fmt.Println("PRE", x)
		x = easeInOutExpo(x)
		fmt.Println("POST", x)
		if x == 1 {
			player.PositionSync.Active = false
			entities.SetLocalPosition(player, player.PositionSync.Goal)
			return
		}

		startPos := entities.GetLocalPosition(player)
		endPos := player.PositionSync.Goal
		delta := endPos.Sub(startPos).Mul(x)
		entities.SetLocalPosition(player, startPos.Add(delta))
	}
}
