package drawer

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/app/apputils"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

const (
	maxContentBrowserHeight float32 = 300
)

var documentTexture *imgui.TextureID

// func BuildDrawer(app renderiface.App, world renderiface.GameWorld, renderContext renderiface.RenderContext, ps []*prefabs.Prefab, x, y float32, height *float32, expanded *bool) {
func BuildDrawer(app renderiface.App, world renderiface.GameWorld, renderContext renderiface.RenderContext, ps []*prefabs.Prefab, x, y float32, height *float32) {
	imgui.SetNextWindowBgAlpha(1)
	r := imgui.ContentRegionAvail()
	imgui.SetNextWindowPosV(imgui.Vec2{x, y}, imgui.ConditionNone, imgui.Vec2{})
	imgui.SetNextWindowSize(imgui.Vec2{X: r.X, Y: *height})

	var open bool = true
	flags := imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoTitleBar
	imgui.BeginV("Drawer", &open, flags)
	if imgui.IsWindowFocused() {
		*height = maxContentBrowserHeight
	} else {
		*height = apputils.CalculateFooterSize(app.RuntimeConfig().UIEnabled)
	}

	if imgui.BeginTabBarV("Drawer Tab Bar", imgui.TabBarFlagsFittingPolicyScroll) {
		contentBrowser(app, world)
		prefabsUI(app, world, ps)
		imgui.EndTabBar()
	}

	imgui.End()
}
