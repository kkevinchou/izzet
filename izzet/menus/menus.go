package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

func BuildExplorer(es map[string]*entities.Entity) {
	parentWindowSize := imgui.WindowSize()
	windowSize := imgui.Vec2{X: parentWindowSize.X * 0.2, Y: parentWindowSize.Y}

	imgui.BeginChildV("explorer", windowSize, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)
	selectedEntity := sceneHierarchy(es)
	entityProps(selectedEntity)
	imgui.EndChild()
}
