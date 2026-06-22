package drawer

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
)

type DrawerTab string

const DrawerTabNone DrawerTab = "NONE"
const DrawerTabContent DrawerTab = "CONTENT"
const DrawerTabMaterials DrawerTab = "MATERIALS"
const DrawerTabPrefabs DrawerTab = "PREFABS"

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

func BuildDrawerbar(app renderiface.App, renderContext renderiface.RenderContext, width int, materialTextureMap map[assets.MaterialHandle]uint32) {
	var drawerbarX float32 = settings.WindowPadding[0]
	var drawerbarY float32 = imgui.MainViewport().Pos().Y + imgui.MainViewport().Size().Y - settings.DrawerbarSize

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 5, Y: 5})

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

	imgui.SameLine()

	if last == DrawerTabPrefabs {
		imgui.PushStyleColorVec4(imgui.ColButton, buttonColorActive)
	} else {
		imgui.PushStyleColorVec4(imgui.ColButton, buttonColorInactive)
	}
	if imgui.Button("Prefabs") {
		if last != DrawerTabPrefabs {
			expanded = true
			last = DrawerTabPrefabs
		} else {
			expanded = false
			last = DrawerTabNone
		}
	}
	imgui.PopStyleColor()

	if expanded {
		var open bool = true
		var drawerTabFlags imgui.WindowFlags = imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse
		drawerTabFlags |= imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoFocusOnAppearing

		imgui.SetNextWindowPos(imgui.Vec2{X: drawerbarX, Y: drawerbarY - drawerTabHeight})
		imgui.SetNextWindowSize(imgui.Vec2{X: drawerTabWidth, Y: drawerTabHeight})
		imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 10, Y: 10})
		imgui.BeginV("DrawerTab", &open, drawerTabFlags)
		imgui.Separator()

		if last == DrawerTabContent {
			contentBrowser(app)
		} else if last == DrawerTabMaterials {
			materialssUI(app, materialTextureMap)
		} else if last == DrawerTabPrefabs {
			prefabsUI(app)
		}
		imgui.End()
		imgui.PopStyleVar()
	}

	imgui.PopStyleVar()
}
