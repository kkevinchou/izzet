package izzet

import (
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/veandco/go-sdl2/sdl"
)

func (g *Izzet) AddEntity(entity *entities.Entity) {
	g.entities[entity.ID] = entity
}

func (g *Izzet) GetPrefabByID(id int) *prefabs.Prefab {
	return g.prefabs[id]
}

func (g *Izzet) Window() *sdl.Window {
	return g.window
}
