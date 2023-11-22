package drawer

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

const (
	maxContentBrowserHeight float32 = 300
)

var documentTexture *imgui.TextureID

func BuildDrawer(app renderiface.App, world renderiface.GameWorld, renderContext renderiface.RenderContext, ps []*prefabs.Prefab, x, y float32, height *float32, expanded *bool) {
	imgui.SetNextWindowBgAlpha(1)
	r := imgui.ContentRegionAvail()
	imgui.SetNextWindowPosV(imgui.Vec2{x, y}, imgui.ConditionNone, imgui.Vec2{})
	imgui.SetNextWindowSize(imgui.Vec2{X: r.X, Y: *height})

	var open bool = true
	flags := imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoTitleBar
	if !*expanded {
		flags |= imgui.WindowFlagsNoScrollbar
	}
	imgui.BeginV("Drawer", &open, flags)

	if imgui.BeginTabBarV("Drawer Tab Bar", imgui.TabBarFlagsFittingPolicyScroll) {
		clicked := false

		if contentBrowser(app, world) {
			clicked = true
		}
		prefabsUI(app, world, ps)
		// if prefabsUI(app, world, ps) {
		// 	clicked = true
		// }

		if clicked {
			*expanded = true
			*height = maxContentBrowserHeight
		}
		imgui.EndTabBar()
	}

	if imgui.IsWindowHovered() && imgui.IsMouseClicked(0) {
		*expanded = true
		*height = maxContentBrowserHeight
	}

	imgui.End()
}
