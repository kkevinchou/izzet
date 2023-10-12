package world

import (
	"sort"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

func (g *GameWorld) AddEntity(entity *entities.Entity) {
	g.entities[entity.ID] = entity
	if entity.BoundingBox() != collider.EmptyBoundingBox {
		g.spatialPartition.IndexEntities([]spatialpartition.Entity{entity})
	}
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
