package drawer

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
)

const (
	maxContentBrowserHeight float32 = 300
)

var documentTexture *imgui.TextureID

var (
	drawerExpanded bool
	currentDrawer  Drawer = DrawerNone
)

type Drawer string

const DrawerNone Drawer = "NONE"
const DrawerContent Drawer = "CONTENT"
const DrawerPrefabs Drawer = "PREFABS"
const DrawerMaterials Drawer = "MATERIALS"

var last = DrawerContent
var expanded bool

func BuildFooter(app renderiface.App, renderContext renderiface.RenderContext, ps []*prefabs.Prefab) {
	_, windowHeight := app.WindowSize()

	imgui.SetNextWindowBgAlpha(1)
	r := imgui.ContentRegionAvail()
	imgui.SetNextWindowPosV(imgui.Vec2{X: 0, Y: float32(windowHeight) - settings.FooterSize}, imgui.CondNone, imgui.Vec2{})
	imgui.SetNextWindowSize(imgui.Vec2{X: r.X, Y: settings.FooterSize})

	var open bool = true
	var flags imgui.WindowFlags = imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse
	flags |= imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoFocusOnAppearing | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse

	if !drawerExpanded {
		flags |= imgui.WindowFlagsNoScrollbar
	}
	imgui.BeginV("Drawer", &open, flags)
	windowFocused := imgui.IsWindowFocused()

	if imgui.BeginTabBarV("Drawer Tab Bar", imgui.TabBarFlagsFittingPolicyScroll) {
		if imgui.BeginTabItem("Content Browser") {
			if last != DrawerContent {
				currentDrawer = DrawerContent
				expanded = true
			} else if imgui.IsItemClicked() {
				expanded = !expanded
			}
			last = DrawerContent
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Prefabs") {
			if last != DrawerPrefabs {
				currentDrawer = DrawerPrefabs
				expanded = true
			} else if imgui.IsItemClicked() {
				expanded = !expanded
			}
			last = DrawerPrefabs
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Materials") {
			if last != DrawerMaterials {
				currentDrawer = DrawerMaterials
				expanded = true
			} else if imgui.IsItemClicked() {
				expanded = !expanded
			}
			last = DrawerMaterials
			imgui.EndTabItem()
		}

		imgui.EndTabBar()
	}

	if expanded {
		imgui.BeginV("ExpandedDrawer", &open, imgui.WindowFlagsNone)
		if last == DrawerContent {
			contentBrowser(app)
		} else if last == DrawerPrefabs {
			prefabsUI(app, ps)
		} else if last == DrawerMaterials {
			materialssUI(app)
		}
		imgui.End()
	}

	drawerExpanded = windowFocused

	imgui.End()
}
