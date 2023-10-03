package izzet

import (
	"fmt"
	"sort"

	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

func (g *Izzet) AddEntity(entity *entities.Entity) {
	g.entities[entity.ID] = entity
	if entity.BoundingBox() != nil {
		g.spatialPartition.IndexEntities([]spatialpartition.Entity{entity})
	}
}

func (g *Izzet) DeleteEntity(entity *entities.Entity) {
	if entity == nil {
		return
	}

	for _, child := range entity.Children {
		entities.RemoveParent(child)
		g.DeleteEntity(child)
	}

	entities.RemoveParent(entity)
	delete(g.entities, entity.ID)
}

func (g *Izzet) GetPrefabByID(id int) *prefabs.Prefab {
	return g.prefabs[id]
}

func (g *Izzet) GetEntityByID(id int) *entities.Entity {
	return g.entities[id]
}

var sortFrame int
var sortedEntities []*entities.Entity

func (g *Izzet) Entities() []*entities.Entity {
	if sortFrame != g.CommandFrame() {
		sortFrame = g.CommandFrame()

		var ids []int
		for id, _ := range g.entities {
			ids = append(ids, id)
		}

		sort.Ints(ids)

		entities := []*entities.Entity{}
		for _, id := range ids {
			entities = append(entities, g.entities[id])
		}
		sortedEntities = entities
	}

	return sortedEntities
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

func (g *Izzet) ModelLibrary() *modellibrary.ModelLibrary {
	return g.modelLibrary
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

func (g *Izzet) NavMesh() *navmesh.NavigationMesh {
	return g.navigationMesh
}

func (g *Izzet) ResetNavMeshVAO() {
	render.ResetNavMeshVAO = true
}

func (g *Izzet) SetShowImguiDemo(value bool) {
	g.showImguiDemo = value
}

func (g *Izzet) ShowImguiDemo() bool {
	return g.showImguiDemo
}

func (g *Izzet) MetricsRegistry() *metrics.MetricsRegistry {
	return g.metricsRegistry
}
