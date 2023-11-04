package world

import (
	"sort"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

func (g *GameWorld) AddEntity(entity *entities.Entity) {
	if _, ok := g.entities[entity.GetID()]; ok {
		return
	}
	g.entities[entity.ID] = entity
}

func (g *GameWorld) DeleteEntity(entity *entities.Entity) {
	if entity == nil {
		return
	}

	for _, child := range entity.Children {
		entities.RemoveParent(child)
		g.DeleteEntity(child)
	}

	entities.RemoveParent(entity)
	g.spatialPartition.DeleteEntity(entity.ID)
	delete(g.entities, entity.ID)
}

func (g *GameWorld) GetEntityByID(id int) *entities.Entity {
	return g.entities[id]
}

func (g *GameWorld) Entities() []*entities.Entity {
	if g.sortFrame != g.CommandFrame() {
		g.sortFrame = g.CommandFrame()

		var ids []int
		for id, _ := range g.entities {
			ids = append(ids, id)
		}

		sort.Ints(ids)

		entities := []*entities.Entity{}
		for _, id := range ids {
			entities = append(entities, g.entities[id])
		}
		g.sortedEntities = entities
	}

	return g.sortedEntities
}

// func (g *GameWorld) Camera() *camera.Camera {
// 	return g.camera
// }

func (g *GameWorld) CommandFrame() int {
	return g.commandFrameCount
}

func (g *GameWorld) IncrementCommandFrameCount() {
	g.commandFrameCount++
}

func (g *GameWorld) Lights() []*entities.Entity {
	allEntities := g.Entities()
	result := []*entities.Entity{}
	for _, e := range allEntities {
		if e.LightInfo != nil {
			result = append(result, e)
		}
	}
	return result
}

// game world
func (g *GameWorld) SpatialPartition() *spatialpartition.SpatialPartition {
	return g.spatialPartition
}

// func (g *GameWorld) NavMesh() *navmesh.NavigationMesh {
// 	return g.navigationMesh
// }

// func (g *GameWorld) ResetNavMeshVAO() {
// 	render.ResetNavMeshVAO = true
// }

func (g *GameWorld) SetFrameInput(input input.Input) {
	g.frameInput = input
}

func (g *GameWorld) GetFrameInput() input.Input {
	return g.frameInput
}

func (g *GameWorld) GetEvents() []events.Event {
	return g.events
}

func (g *GameWorld) QueueEvent(event events.Event) {
	g.events = append(g.events, event)
}

func (g *GameWorld) ClearEventQueue() {
	g.events = nil
}

// DIRTY HACK: the camera orientation isn't really an input, but is contextual
// information for inputs. I don't know a good place to put this yet so I'm
// hijacking input.Input
func (g *GameWorld) SetInputCameraRotation(rotation mgl64.Quat) {
	g.frameInput.CameraRotation = rotation
}
