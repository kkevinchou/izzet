package izzet

import (
	"fmt"
	"sort"

	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/veandco/go-sdl2/sdl"
)

func (g *Izzet) AddEntity(entity *entities.Entity) {
	g.entities[entity.ID] = entity
}

func (g *Izzet) GetPrefabByID(id int) *prefabs.Prefab {
	return g.prefabs[id]
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

func (g *Izzet) AssetManager() *assets.AssetManager {
	return g.assetManager
}

func (g *Izzet) Camera() *camera.Camera {
	return g.camera
}

func (g *Izzet) Window() *sdl.Window {
	return g.window
}

func (g *Izzet) Platform() *input.SDLPlatform {
	return g.platform
}

func (g *Izzet) Serializer() *serialization.Serializer {
	return g.serializer
}

func (g *Izzet) SaveWorld() {
	g.serializer.WriteOut("./scene.txt")
}

func (g *Izzet) LoadWorld() {
	err := g.serializer.ReadIn("./scene.txt")
	if err != nil {
		fmt.Println("failed to load world: ", err)
		return
	}
	es := g.serializer.Entities()
	g.entities = map[int]*entities.Entity{}
	for _, e := range es {
		g.entities[e.ID] = e
	}
}
