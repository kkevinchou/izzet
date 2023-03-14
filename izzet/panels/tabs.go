package panels

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

func BuildTabs(world World, renderContext RenderContext, menuBarSize imgui.Vec2, ps []*prefabs.Prefab) {
	rect := imgui.Vec2{X: float32(renderContext.Width()), Y: float32(renderContext.Height()) - menuBarSize.Y}
	imgui.SetNextWindowBgAlpha(0.8)
	imgui.SetNextWindowPosV(imgui.Vec2{Y: menuBarSize.Y + rect.Y*0.5}, imgui.ConditionOnce, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: rect.X * 0.15, Y: rect.Y * 0.5}, imgui.ConditionOnce)

	imgui.BeginV("Tabs Window", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse)

	if imgui.BeginTabBar("Main") {
		if imgui.BeginTabItem("World Properties") {
			worldProps(renderContext)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Entity Properties") {
			entityProps(SelectedEntity())
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Prefabs") {
			prefabsUI(ps)
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}

	imgui.End()
}

// some detailed comment here
func createUserSpaceTextureHandle(texture uint32) imgui.TextureID {
	handle := 1<<63 | uint64(texture)
	return imgui.TextureID(handle)
}
