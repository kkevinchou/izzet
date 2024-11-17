package panels

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func postProcessing(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()

	imgui.BeginTableV("General Table", 2, tableFlags, imgui.Vec2{}, 0)
	panelutils.InitColumns()
	panelutils.SetupRow("Kuwahara Filter", func() { imgui.Checkbox("", &runtimeConfig.KuwaharaFilter) }, true)
	imgui.EndTable()
}
