package panels

import (
	"github.com/inkyblackness/imgui-go/v4"
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
	BuildRelation(parent *entities.Entity, child *entities.Entity)
	RemoveParent(child *entities.Entity)
	Entities() []*entities.Entity
}

func SelectEntity(entity *entities.Entity) bool {
	newSelection := entity != selectedEntity
	selectedEntity = entity
	return newSelection
}

func SelectedEntity() *entities.Entity {
	return selectedEntity
}

func setupRow(label string, item func()) {
	imgui.TableNextRow()
	imgui.TableNextColumn()
	imgui.Text(label)
	imgui.TableNextColumn()
	imgui.PushItemWidth(300)
	imgui.PushID(label)
	item()
	imgui.PopID()
	imgui.PopItemWidth()
}
