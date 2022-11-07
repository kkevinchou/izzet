package izzet

import (
	"sort"

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

func (g *Izzet) Entities() []*entities.Entity {
	var ids []int
	for id, _ := range g.entities {
		ids = append(ids, id)
	}

	sort.Ints(ids)

	entities := []*entities.Entity{}
	for _, id := range ids {
		entities = append(entities, g.entities[id])
	}

	return entities
}

func (g *Izzet) Prefabs() []*prefabs.Prefab {
	var ids []int
	for id, _ := range g.prefabs {
		ids = append(ids, id)
	}

	sort.Ints(ids)

	ps := []*prefabs.Prefab{}
	for _, id := range ids {
		ps = append(ps, g.prefabs[id])
	}

	return ps
}
