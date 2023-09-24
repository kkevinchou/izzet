package panels

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

// var open bool

func BuildTabsSet(world World, renderContext RenderContext, menuBarSize imgui.Vec2, ps []*prefabs.Prefab) {
	rect := imgui.Vec2{X: float32(renderContext.Width()), Y: float32(renderContext.Height()) - menuBarSize.Y}
	width := rect.X * 0.20
	height := rect.Y * 0.5

	imgui.SetNextWindowBgAlpha(0.8)
	imgui.SetNextWindowPosV(imgui.Vec2{Y: menuBarSize.Y}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: width, Y: height}, imgui.ConditionFirstUseEver)
	imgui.BeginV("Fixed Tab Set", nil, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse)

	if imgui.BeginTabBar("Scene") {
		if imgui.BeginTabItem("Scene Hierarchy") {
			sceneUI(world)
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}

	imgui.End()

	imgui.SetNextWindowBgAlpha(0.8)
	var propertiesWidth float32 = 450
	imgui.SetNextWindowPosV(imgui.Vec2{X: menuBarSize.X - propertiesWidth, Y: menuBarSize.Y}, imgui.ConditionNone, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: propertiesWidth, Y: rect.Y}, imgui.ConditionNone)
	imgui.BeginV("Right Window", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoScrollWithMouse)

	if imgui.BeginTabBarV("Main", imgui.TabBarFlagsFittingPolicyScroll|imgui.TabBarFlagsReorderable) {
		if imgui.BeginTabItem("World") {
			worldProps(world, renderContext)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Details") {
			entityProps(SelectedEntity())
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Prefabs") {
			prefabsUI(world, ps)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Animation") {
			entity := SelectedEntity()
			if entity != nil && entity.AnimationPlayer != nil {
				animationUI(world, entity)
			} else {
				imgui.Text("<select an entity with animations>")
			}
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Stats") {
			stats(world, renderContext)
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}

	imgui.End()
}
