package world

import (
	"sort"

	"github.com/kkevinchou/izzet/app/entities"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

func (g *GameWorld) AddEntity(entity *entities.Entity) {
	if _, ok := g.entities[entity.GetID()]; ok {
		return
	}
	g.entities[entity.ID] = entity
}

func (g *GameWorld) DestroyEntity(entityID int) {
	if _, ok := g.entities[entityID]; !ok {
		return
	}
	delete(g.entities, entityID)
}

func (g *GameWorld) DeleteEntity(entityID int) {
	entity := g.GetEntityByID(entityID)
	if entity == nil {
		return
	}

	for _, child := range entity.Children {
		entities.RemoveParent(child)
		g.DeleteEntity(child.GetID())
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

func (g *GameWorld) SpatialPartition() *spatialpartition.SpatialPartition {
	return g.spatialPartition
}
