package drawer

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/app/apputils"
	"github.com/kkevinchou/izzet/app/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

const (
	maxContentBrowserHeight float32 = 300
)

var documentTexture *imgui.TextureID

var (
	drawerExpanded bool
)

// func BuildDrawer(app renderiface.App, renderContext renderiface.RenderContext, ps []*prefabs.Prefab, x, y float32, height *float32, expanded *bool) {
func BuildDrawer(app renderiface.App, renderContext renderiface.RenderContext, ps []*prefabs.Prefab) {
	_, windowHeight := app.WindowSize()
	var height = maxContentBrowserHeight
	if !drawerExpanded {
		height = apputils.CalculateFooterSize(app.RuntimeConfig().UIEnabled)
	}

	imgui.SetNextWindowBgAlpha(1)
	r := imgui.ContentRegionAvail()
	imgui.SetNextWindowPosV(imgui.Vec2{X: 0, Y: float32(windowHeight) - height}, imgui.CondNone, imgui.Vec2{})
	imgui.SetNextWindowSize(imgui.Vec2{X: r.X, Y: height})

	var open bool = true
	var flags imgui.WindowFlags = imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoFocusOnAppearing
	if !drawerExpanded {
		flags |= imgui.WindowFlagsNoScrollbar
	}
	imgui.BeginV("Drawer", &open, flags)
	windowFocused := imgui.IsWindowFocused()

	var menuOpen bool
	if imgui.BeginTabBarV("Drawer Tab Bar", imgui.TabBarFlagsFittingPolicyScroll) {
		if contentBrowser(app) {
			menuOpen = true
		}
		if prefabsUI(app, ps) {
			menuOpen = true
		}
		imgui.EndTabBar()
	}

	drawerExpanded = windowFocused || menuOpen

	imgui.End()
}
