package panels

import (
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/veandco/go-sdl2/sdl"
)

// var HierarchySelection int
var selectedEntity *entities.Entity
var ShowDebug bool = true

type World interface {
	AddEntity(entity *entities.Entity)
	GetPrefabByID(id int) *prefabs.Prefab
	GetEntityByID(id int) *entities.Entity
	BuildRelation(parent *entities.Entity, child *entities.Entity)
	RemoveParent(child *entities.Entity)
	Entities() []*entities.Entity
	Window() *sdl.Window
}

func SelectEntity(entity *entities.Entity) bool {
	newSelection := entity != selectedEntity
	selectedEntity = entity
	return newSelection
}

func SelectedEntity() *entities.Entity {
	return selectedEntity
}
