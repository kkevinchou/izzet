package izzet

import (
	"fmt"
	"sort"

	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

func (g *Izzet) AddEntity(entity *entities.Entity) {
	g.entities[entity.ID] = entity
}

func (g *Izzet) DeleteEntity(entity *entities.Entity) {
	if entity == nil {
		return
	}
	delete(g.entities, entity.ID)
	g.RemoveParent(entity)
}

func (g *Izzet) GetPrefabByID(id int) *prefabs.Prefab {
	return g.prefabs[id]
}

func (g *Izzet) GetEntityByID(id int) *entities.Entity {
	return g.entities[id]
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
		panic(fmt.Sprintf("failed to load world: %s", err))
	}

	var maxID int
	es := g.serializer.Entities()
	g.entities = map[int]*entities.Entity{}
	for _, e := range es {
		if e.ID > maxID {
			maxID = e.ID
		}
		g.entities[e.ID] = e
	}

	if len(g.entities) > 0 {
		entities.SetNextID(maxID + 1)
	}

	panels.SelectEntity(nil)
	g.editHistory.Clear()
}

func (g *Izzet) AppendEdit(edit edithistory.Edit) {
	g.editHistory.Append(edit)
}

func (g *Izzet) Redo() {
	g.editHistory.Redo()
}

func (g *Izzet) Undo() {
	g.editHistory.Undo()
}

func (g *Izzet) BuildRelation(parent *entities.Entity, child *entities.Entity) {
	g.RemoveParent(child)
	parent.Children[child.ID] = child
	child.Parent = parent
}

func (g *Izzet) RemoveParent(child *entities.Entity) {
	if parent := child.Parent; parent != nil {
		delete(parent.Children, child.ID)
		child.Parent = nil
	}
}

func (g *Izzet) CommandFrame() int {
	return g.commandFrameCount
}

func (g *Izzet) Lights() []*entities.Entity {
	allEntities := g.Entities()
	result := []*entities.Entity{}
	for _, e := range allEntities {
		if e.LightInfo != nil {
			result = append(result, e)
		}
	}
	return result
}

func (g *Izzet) SpatialPartition() *spatialpartition.SpatialPartition {
	return g.spatialPartition
}
