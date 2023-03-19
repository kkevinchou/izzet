package panels

import (
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

type RenderContext interface {
	Width() int
	Height() int
	AspectRatio() float64
}

// var HierarchySelection int
var selectedEntity *entities.Entity
var ShowDebug bool = true

type World interface {
	AddEntity(entity *entities.Entity)
	GetPrefabByID(id int) *prefabs.Prefab
	GetEntityByID(id int) *entities.Entity
	Entities() []*entities.Entity
}

func SelectEntity(entity *entities.Entity) {
	selectedEntity = entity
}

func SelectedEntity() *entities.Entity {
	return selectedEntity
}
