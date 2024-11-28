package panels

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func rendering(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Rendering Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		panelutils.SetupRow("Near", func() { imgui.SliderFloat("", &runtimeConfig.Near, 0.1, 1) }, true)
		panelutils.SetupRow("Far", func() { imgui.SliderFloat("", &runtimeConfig.Far, 0, 1500) }, true)
		panelutils.SetupRow("FovX", func() { imgui.SliderFloat("", &runtimeConfig.FovX, 0, 170) }, true)

		panelutils.SetupRow("Debug Color", func() {
			imgui.ColorEdit3V("", &runtimeConfig.Color, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
		}, true)
		panelutils.SetupRow("Batch Render", func() {
			imgui.Checkbox("", &runtimeConfig.BatchRenderingEnabled)
		}, true)
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("SSAO", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("SSAO", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		panelutils.SetupRow("Radius", func() { imgui.SliderFloat("", &runtimeConfig.SSAORadius, 0, 1) }, true)
		panelutils.SetupRow("Bias", func() { imgui.SliderFloat("", &runtimeConfig.SSAOBias, 0, 1) }, true)
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Lighting", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Lighting Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		panelutils.SetupRow("Ambient Factor", func() { imgui.SliderFloat("", &runtimeConfig.AmbientFactor, 0, 1) }, true)
		panelutils.SetupRow("Point Light Bias", func() { imgui.SliderFloat("", &runtimeConfig.PointLightBias, 0, 1) }, true)
		panelutils.SetupRow("Enable Shadow Mapping", func() { imgui.Checkbox("", &runtimeConfig.EnableShadowMapping) }, true)
		panelutils.SetupRow("Shadow Far Distance", func() { imgui.SliderFloat("", &runtimeConfig.ShadowFarDistance, 0, 1000) }, true)
		panelutils.SetupRow("Fog Density", func() { imgui.SliderInt("", &runtimeConfig.FogDensity, 0, 100) }, true)
		panelutils.SetupRow("Enable Bloom", func() { imgui.Checkbox("", &runtimeConfig.Bloom) }, true)
		panelutils.SetupRow("Bloom Intensity", func() { imgui.SliderFloat("", &runtimeConfig.BloomIntensity, 0, 1) }, true)
		panelutils.SetupRow("Bloom Threshold Passes", func() { imgui.SliderInt("", &runtimeConfig.BloomThresholdPasses, 0, 3) }, true)
		panelutils.SetupRow("Bloom Threshold", func() { imgui.SliderFloat("", &runtimeConfig.BloomThreshold, 0, 3) }, true)
		panelutils.SetupRow("Bloom Upsampling Scale", func() { imgui.SliderFloat("", &runtimeConfig.BloomUpsamplingScale, 0, 5.0) }, true)
		panelutils.SetupRow("Shadow Map Z Offset", func() { imgui.SliderFloat("", &runtimeConfig.ShadowmapZOffset, 0, 2000) }, true)
		panelutils.SetupRow("SP Near Plane Offset", func() { imgui.SliderFloat("", &runtimeConfig.ShadowSpatialPartitionNearPlane, 0, 2000) }, true)
		panelutils.SetupRow("Skybox Top Color", func() {
			imgui.ColorEdit3V("", &runtimeConfig.SkyboxTopColor, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
		}, true)
		panelutils.SetupRow("Skybox Bottom Color", func() {
			imgui.ColorEdit3V("", &runtimeConfig.SkyboxBottomColor, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
		}, true)
		panelutils.SetupRow("Skybox Mix Value", func() { imgui.SliderFloat("##", &runtimeConfig.SkyboxMixValue, 0, 1) }, true)
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Post Processing", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Post Processing Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		panelutils.SetupRow("Kuwahara Filter", func() { imgui.Checkbox("", &runtimeConfig.KuwaharaFilter) }, true)
		imgui.EndTable()
	}

}
