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
	shelfExpanded bool
	currentSehfl  ShelfType = ShelfNone
)

type ShelfType string

const ShelfNone ShelfType = "NONE"
const ShelfContent ShelfType = "CONTENT"
const ShelfPrefabs ShelfType = "PREFABS"
const ShelfMaterials ShelfType = "MATERIALS"

var last = ShelfContent
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

	if !shelfExpanded {
		flags |= imgui.WindowFlagsNoScrollbar
	}
	imgui.BeginV("Footer", &open, flags)
	windowFocused := imgui.IsWindowFocused()

	if imgui.BeginTabBarV("Footer Tab Bar", imgui.TabBarFlagsFittingPolicyScroll) {
		if imgui.BeginTabItem("Content Browser") {
			if last != ShelfContent {
				currentSehfl = ShelfContent
				expanded = true
			} else if imgui.IsItemClicked() {
				expanded = !expanded
			}
			last = ShelfContent
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Prefabs") {
			if last != ShelfPrefabs {
				currentSehfl = ShelfPrefabs
				expanded = true
			} else if imgui.IsItemClicked() {
				expanded = !expanded
			}
			last = ShelfPrefabs
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Materials") {
			if last != ShelfMaterials {
				currentSehfl = ShelfMaterials
				expanded = true
			} else if imgui.IsItemClicked() {
				expanded = !expanded
			}
			last = ShelfMaterials
			imgui.EndTabItem()
		}

		imgui.EndTabBar()
	}

	if expanded {
		//  := imgui.MainViewport().Center()
		// fmt.Println(imgui.MainViewport().Center().Y)
		width, height := app.WindowSize()
		_ = width
		_ = height

		imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: float32(height) - settings.FooterSize - settings.ShelfHeight - 2})
		imgui.SetNextWindowSize(imgui.Vec2{X: settings.ShelfWidth, Y: settings.ShelfHeight})
		imgui.BeginV("Shelf", &open, flags)
		if last == ShelfContent {
			contentBrowser(app)
		} else if last == ShelfPrefabs {
			prefabsUI(app, ps)
		} else if last == ShelfMaterials {
			materialssUI(app)
		}
		imgui.End()
	}

	shelfExpanded = windowFocused

	imgui.End()
}
