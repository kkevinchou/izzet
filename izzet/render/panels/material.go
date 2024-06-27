package panels

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/material"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/types"
)

func spaghettios(app renderiface.App, renderContext RenderContext) {
	// settings := app.RuntimeConfig()
	// mr := app.MetricsRegistry()

	if imgui.CollapsingHeaderTreeNodeFlagsV("Materials", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Material Editor", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		// imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)

		material := material.Material{
			PBR: types.PBR{},
		}

		// Frame Profiling
		setupRow("Diffuse", func() {
			imgui.ColorEdit3V("", &material.PBR.Diffuse, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
		}, true)
		setupRow("Invisible", func() {
			imgui.Checkbox("", &material.Invisible)
		}, true)

		setupRow("Diffuse Intensity", func() {
			imgui.SliderFloatV("", &material.PBR.DiffuseIntensity, 1, 20, "%.1f", imgui.SliderFlagsNone)
		}, true)

		setupRow("Roughness", func() { imgui.SliderFloatV("", &material.PBR.Roughness, 0, 1, "%.2f", imgui.SliderFlagsNone) }, true)
		setupRow("Metallic Factor", func() { imgui.SliderFloatV("", &material.PBR.Metallic, 0, 1, "%.2f", imgui.SliderFlagsNone) }, true)

		imgui.EndTable()
		if imgui.Button("Create Material") {
			app.CreateMaterial(material)
		}
	}
}
