package panels

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

var open bool = true

func BuildExplorer(es []*entities.Entity, world World, menuBarSize imgui.Vec2, renderContext RenderContext) {
	rect := imgui.Vec2{X: float32(renderContext.Width()), Y: float32(renderContext.Height()) - menuBarSize.Y}
	imgui.SetNextWindowBgAlpha(0.8)
	imgui.SetNextWindowPosV(imgui.Vec2{Y: menuBarSize.Y}, imgui.ConditionOnce, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: rect.X * 0.15, Y: rect.Y * 0.5}, imgui.ConditionOnce)

	imgui.BeginV("Explorer Window", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse)

	if imgui.BeginTabBar("Scene") {
		if imgui.BeginTabItem("Scene Hierarchy") {
			sceneUI(es, world)
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}

	imgui.End()
}
