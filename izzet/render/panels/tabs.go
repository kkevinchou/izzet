package panels

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var open bool

func BuildTabsSet(app renderiface.App, renderContext RenderContext, ps []*prefabs.Prefab) {
	// rect := imgui.Vec2{X: float32(renderContext.Width()), Y: float32(renderContext.Height()) - menuBarSize.Y}
	// width := rect.X * 0.20
	// height := rect.Y * 0.5

	// imgui.SetNextWindowBgAlpha(0.8)
	// imgui.SetNextWindowPosV(imgui.Vec2{Y: menuBarSize.Y}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	// imgui.SetNextWindowSizeV(imgui.Vec2{X: width, Y: height}, imgui.ConditionFirstUseEver)

	// imgui.BeginV("Fixed Tab Set", nil, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse)
	// imgui.BeginV("Fixed Tab Set", nil, imgui.WindowFlagsNoTitleBar)

	// imgui.BeginChild("Fixed Tab Set")
	// if imgui.BeginTabBar("Scene") {
	// 	if imgui.BeginTabItem("Scene Hierarchy") {
	// 		sceneUI(app, world)
	// 		imgui.EndTabItem()
	// 	}
	// 	imgui.EndTabBar()
	// }
	// imgui.EndChild()

	// imgui.SetNextWindowBgAlpha(0.8)
	// imgui.SetNextWindowPosV(imgui.Vec2{X: menuBarSize.X - propertiesWidth, Y: menuBarSize.Y}, imgui.ConditionNone, imgui.Vec2{})
	// imgui.SetNextWindowSizeV(imgui.Vec2{X: propertiesWidth, Y: rect.Y}, imgui.ConditionNone)
	// imgui.BeginV("Right Window", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoScrollWithMouse)
	imgui.BeginChildStrV("Right Window", imgui.Vec2{}, false, imgui.WindowFlagsNoBringToFrontOnFocus)

	// if imgui.BeginTabBarV("Main", imgui.TabBarFlagsFittingPolicyScroll|imgui.TabBarFlagsReorderable) {
	if imgui.BeginTabBar("Main") {
		if imgui.BeginTabItem("Details") {
			entityProps(app.SelectedEntity(), app)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Scene Graph") {
			sceneGraph(app)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("World") {
			worldProps(app)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Stats") {
			stats(app, renderContext)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Materials") {
			spaghettios(app, renderContext)
			imgui.EndTabItem()
		}
		// if imgui.BeginTabItem("Animation") {
		// 	entity := SelectedEntity()
		// 	if entity != nil && entity.Animation != nil {
		// 		animationUI(app, entity)
		// 	} else {
		// 		imgui.Text("<select an entity with animations>")
		// 	}
		// 	imgui.EndTabItem()
		// }
		imgui.EndTabBar()
	}

	imgui.EndChild()
}
