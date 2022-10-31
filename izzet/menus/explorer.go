package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
)

var open bool = true

func BuildExplorer(es map[string]*entities.Entity) {
	imgui.SetNextWindowBgAlpha(0.8)
	imgui.SetNextWindowPosV(imgui.Vec2{}, imgui.ConditionAlways, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: float32(settings.Width) * 0.15, Y: float32(settings.Height)}, imgui.ConditionAlways)

	imgui.BeginV("explorer window", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize)
	imgui.BeginChildV("explorer", imgui.Vec2{}, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	selectedEntity := sceneHierarchy(es)
	entityProps(selectedEntity)

	imgui.EndChild()
	imgui.End()
}
