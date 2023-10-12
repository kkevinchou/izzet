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
)

func (g *Izzet) GetPrefabByID(id int) *prefabs.Prefab {
	return g.prefabs[id]
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

func (g *Izzet) SaveWorld(name string) {
	g.serializer.WriteOut(fmt.Sprintf("./%s.json", name))
}

func (g *Izzet) LoadWorld(name string) {
	err := g.serializer.ReadIn(fmt.Sprintf("./%s.json", name))
	if err != nil {
		fmt.Println("failed to load world", name, err)
		return
	}

	g.sortFrame = -1
	g.sortedEntities = []*entities.Entity{}

	g.editHistory.Clear()
	g.spatialPartition.Clear()

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

}

// game world
func (g *Izzet) AppendEdit(edit edithistory.Edit) {
	g.editHistory.Append(edit)
}

// game world
func (g *Izzet) Redo() {
	g.editHistory.Redo()
}

// game world
func (g *Izzet) Undo() {
	g.editHistory.Undo()
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
