package drawer

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

type ShelfType string

const ShelfNone ShelfType = "NONE"
const ShelfContent ShelfType = "CONTENT"
const ShelfMaterials ShelfType = "MATERIALS"

var last = ShelfContent
var expanded bool

func BuildFooter(app renderiface.App, renderContext renderiface.RenderContext, materialTextureMap map[types.MaterialHandle]uint32) {
	_, windowHeight := app.WindowSize()

	imgui.SetNextWindowBgAlpha(1)
	r := imgui.ContentRegionAvail()
	imgui.SetNextWindowPosV(imgui.Vec2{X: 0, Y: float32(windowHeight) - settings.FooterSize}, imgui.CondNone, imgui.Vec2{})
	imgui.SetNextWindowSize(imgui.Vec2{X: r.X, Y: settings.FooterSize})

	var open bool = true
	var footerFlags imgui.WindowFlags = imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse
	footerFlags |= imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoFocusOnAppearing | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse

	imgui.BeginV("Footer", &open, footerFlags)

	if imgui.BeginTabBarV("Footer Tab Bar", imgui.TabBarFlagsFittingPolicyScroll) {
		if imgui.BeginTabItem("Content Browser") {
			if last != ShelfContent {
				expanded = true
			} else if imgui.IsItemClicked() {
				expanded = !expanded
			}
			last = ShelfContent
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Materials") {
			if last != ShelfMaterials {
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
		_, height := app.WindowSize()
		var shelfFlags imgui.WindowFlags = imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse
		shelfFlags |= imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoFocusOnAppearing | imgui.WindowFlagsMenuBar

		imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: float32(height) - settings.FooterSize - settings.ShelfHeight - 2})
		imgui.SetNextWindowSize(imgui.Vec2{X: settings.ShelfWidth, Y: settings.ShelfHeight})
		imgui.BeginV("Shelf", &open, shelfFlags)
		if imgui.BeginMenuBar() {
			imgui.EndMenuBar()
		}

		if last == ShelfContent {
			contentBrowser(app)
		} else if last == ShelfMaterials {
			materialssUI(app, materialTextureMap)
		}
		imgui.End()
	}

	imgui.End()
}
