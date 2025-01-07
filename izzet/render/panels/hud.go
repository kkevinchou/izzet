package panels

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func hud(app renderiface.App, renderContext RenderContext) {
	if imgui.CollapsingHeaderTreeNodeFlagsV("HUD", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("HUD Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("Show HUD", func() { imgui.Checkbox("", &app.RuntimeConfig().ShowHUD) }, true)

		imgui.EndTable()
	}
}
