package panels

import (
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/veandco/go-sdl2/sdl"
)

type World interface {
	AddEntity(entity *entities.Entity)
	GetPrefabByID(id int) *prefabs.Prefab
	Window() *sdl.Window
}
