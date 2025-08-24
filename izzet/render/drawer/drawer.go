package drawer

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/types"
)

type DrawerTab string

const DrawerTabNone DrawerTab = "NONE"
const DrawerTabContent DrawerTab = "CONTENT"
const DrawerTabMaterials DrawerTab = "MATERIALS"

const (
	drawerTabHeight float32 = 210
	drawerTabWidth  float32 = 800
)

var (
	buttonColorInactive imgui.Vec4 = imgui.Vec4{X: .1, Y: .1, Z: 0.1, W: 1}
	buttonColorActive   imgui.Vec4 = imgui.Vec4{X: .3, Y: .3, Z: 0.3, W: 1}

	last     = DrawerTabNone
	expanded bool
)

var DockID imgui.ID

func BuildDrawerbar(app renderiface.App, renderContext renderiface.RenderContext, width int, materialTextureMap map[types.MaterialHandle]uint32) {
	var open bool = true

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 5, Y: 5})
	var drawerbarFlags imgui.WindowFlags = imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse
	drawerbarFlags |= imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoFocusOnAppearing | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse

	imgui.BeginV("Drawer", nil, drawerbarFlags)

	DockID = imgui.WindowDockID()

	if last == DrawerTabContent {
		imgui.PushStyleColorVec4(imgui.ColButton, buttonColorActive)
	} else {
		imgui.PushStyleColorVec4(imgui.ColButton, buttonColorInactive)
	}
	if imgui.Button("Content Browser") {
		if last != DrawerTabContent {
			expanded = true
			last = DrawerTabContent
		} else {
			expanded = false
			last = DrawerTabNone
		}
	}
	imgui.PopStyleColor()

	imgui.SameLine()

	if last == DrawerTabMaterials {
		imgui.PushStyleColorVec4(imgui.ColButton, buttonColorActive)
	} else {
		imgui.PushStyleColorVec4(imgui.ColButton, buttonColorInactive)
	}
	if imgui.Button("Materials") {
		if last != DrawerTabMaterials {
			expanded = true
			last = DrawerTabMaterials
		} else {
			expanded = false
			last = DrawerTabNone
		}
	}
	imgui.PopStyleColor()

	if expanded {
		var drawerTabFlags imgui.WindowFlags = imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse
		drawerTabFlags |= imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoFocusOnAppearing

		// imgui.SetNextWindowPos(imgui.Vec2{X: drawerbarX, Y: drawerbarY - drawerTabHeight})
		// imgui.SetNextWindowSize(imgui.Vec2{X: drawerTabWidth, Y: drawerTabHeight})
		imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 10, Y: 10})
		imgui.BeginV("DrawerTab", &open, drawerTabFlags)
		imgui.Separator()

		if last == DrawerTabContent {
			contentBrowser(app)
		} else if last == DrawerTabMaterials {
			materialssUI(app, materialTextureMap)
		}
		imgui.End()
		imgui.PopStyleVar()
	}

	imgui.End()
	imgui.PopStyleVar()
}
