package izzet

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/world"
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
	g.serializer.WriteToFile(g.world, fmt.Sprintf("./%s.json", name))
}

func (g *Izzet) LoadWorld(name string) {
	if name == "" {
		return
	}

	filename := fmt.Sprintf("./%s.json", name)
	world, err := g.serializer.ReadFromFile(filename)
	if err != nil {
		fmt.Println("failed to load world", filename, err)
		panic(err)
	}

	g.editHistory.Clear()
	g.world.SpatialPartition().Clear()

	var maxID int
	for _, e := range world.Entities() {
		if e.ID > maxID {
			maxID = e.ID
		}
		g.entities[e.ID] = e
	}

	if len(g.entities) > 0 {
		entities.SetNextID(maxID + 1)
	}

	panels.SelectEntity(nil)
	g.SetWorld(world)
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

func (g *Izzet) SetWorld(world *world.GameWorld) {
	g.world = world
	g.renderer.SetWorld(world)
}

func (g *Izzet) StartLiveWorld() {
	if g.AppMode() != app.AppModeEditor {
		return
	}
	g.appMode = app.AppModePlay
	g.editorWorld = g.world

	var buffer bytes.Buffer
	err := g.serializer.Write(g.world, &buffer)
	if err != nil {
		panic(err)
	}

	liveWorld, err := g.serializer.Read(&buffer)
	if err != nil {
		panic(err)
	}

	// TODO: more global state that needs to be cleaned up still, mostly around entities that are selected
	panels.SelectEntity(nil)
	g.SetWorld(liveWorld)
}

func (g *Izzet) StopLiveWorld() {
	if g.AppMode() != app.AppModePlay {
		return
	}
	g.appMode = app.AppModeEditor
	// TODO: more global state that needs to be cleaned up still, mostly around entities that are selected
	panels.SelectEntity(nil)
	g.SetWorld(g.editorWorld)
}

func (g *Izzet) AppMode() app.AppMode {
	return g.appMode
}
