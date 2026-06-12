package panels

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/ui"
)

func hud(app renderiface.App, renderContext RenderContext) {
	if imgui.CollapsingHeaderTreeNodeFlagsV("HUD", imgui.TreeNodeFlagsDefaultOpen) {
		ui.Table("HUD Table", func() {
			ui.CheckboxRow("Show HUD", &app.RuntimeConfig().ShowHUD)
		})
	}
}
