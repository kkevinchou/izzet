package izzet

import (
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

func (g *Izzet) AddEntity(entity *entities.Entity) {
	g.entities[entity.ID] = entity
}

func (g *Izzet) GetPrefabByID(id int) *prefabs.Prefab {
	return g.prefabs[id]
}
