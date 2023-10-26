package shared

import "github.com/kkevinchou/izzet/izzet/entities"

type GameWorld interface {
	GetEntityByID(int) *entities.Entity
	Entities() []*entities.Entity
}
