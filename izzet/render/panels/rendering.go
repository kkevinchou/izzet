package panels

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/ui"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func Rendering(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsDefaultOpen) {
		ui.Table("Rendering Table", func() {
			ui.SliderFloatRow("Near", &runtimeConfig.Near, 0.1, 1)
			ui.SliderFloatRow("Far", &runtimeConfig.Far, 0, 1500)
			ui.SliderFloatRow("FovX", &runtimeConfig.FovX, 0, 170)
			ui.CheckboxRow("Batch Render", &runtimeConfig.BatchRenderingEnabled)
			ui.CheckboxRow("Antialiasing", &runtimeConfig.EnableAntialiasing)
			ui.CheckboxRow("SSAO", &runtimeConfig.EnableSSAO)
			ui.CheckboxRow("Bloom", &runtimeConfig.Bloom)
			ui.CheckboxRow("Shadow Mapping", &runtimeConfig.EnableShadowMapping)
			ui.Row("Draw Nav Mesh", func() {
				if imgui.BeginCombo("##value", string(SelectedNavmeshRenderComboOption)) {
					for _, option := range navmeshRenderComboOptions {
						if imgui.SelectableBool(string(option)) {
							SelectedNavmeshRenderComboOption = option
						}
					}
					imgui.EndCombo()
				}
			})
		})
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Lighting", imgui.TreeNodeFlagsNone) {
		ui.Table("Lighting Table", func() {
			ui.SectionRow("Lights")
			ui.SliderFloatRow("Ambient Factor", &runtimeConfig.AmbientFactor, 0, 100)
			ui.SliderFloatRow("Specular Factor", &runtimeConfig.SpecularFactor, 0, 1)
			ui.SliderFloatRow("Point Light Bias", &runtimeConfig.PointLightBias, 0, 1)
			ui.SectionRow("Shadow Cascades")
			ui.SliderFloatRow("Shadow Map Min Bias", &runtimeConfig.ShadowMapMinBias, 0, 100)
			ui.SliderFloatRow("Shadow Map Angle Bias Rate", &runtimeConfig.ShadowMapAngleBiasRate, 0, 100)
			ui.SliderFloatRow("Shadow Near Distance", &runtimeConfig.ShadowNearDistance, 0.01, 1)
			ui.SliderFloatRow("Shadow Far Distance", &runtimeConfig.ShadowFarDistance, 0, 1000)
			ui.SliderFloatRow("Shadow Cascade Blend", &runtimeConfig.ShadowCascadeBlendFactor, 0, 1)
			ui.SliderFloatRow("Shadow Cascade Crossfade", &runtimeConfig.ShadowCascadeCrossfadeBoundary, 0, 1)
			ui.SliderIntRow("Shadow Max Cascade", &runtimeConfig.ShadowMapMaxCascadeIndex, 0, int32(settings.NumShadowMapCascades-1))
			ui.SectionRow("Bloom")
			ui.SliderFloatRow("Bloom Intensity", &runtimeConfig.BloomIntensity, 0, 1)
			ui.SliderIntRow("Bloom Threshold Passes", &runtimeConfig.BloomThresholdPasses, 0, 3)
			ui.SliderFloatRow("Bloom Threshold", &runtimeConfig.BloomThreshold, 0, 3)
			ui.SliderFloatRow("Bloom Upsampling Scale", &runtimeConfig.BloomUpsamplingScale, 0, 5.0)
			ui.SliderFloatRow("Shadow Map Z Offset", &runtimeConfig.ShadowmapZOffset, 0, 2000)
			ui.SectionRow("SSAO")
			ui.SliderFloatRow("Radius", &runtimeConfig.SSAORadius, 0, 10)
			ui.SliderFloatRow("Bias", &runtimeConfig.SSAOBias, 0, 1)
		})
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Skybox", imgui.TreeNodeFlagsNone) {
		ui.Table("Skybox Table", func() {
			ui.Row("Skybox Top Color", func() {
				imgui.ColorEdit3V("##value", &runtimeConfig.SkyboxTopColor, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
			})
			ui.Row("Skybox Bottom Color", func() {
				imgui.ColorEdit3V("##value", &runtimeConfig.SkyboxBottomColor, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
			})
			ui.SliderFloatRow("Skybox Mix Value", &runtimeConfig.SkyboxMixValue, 0, 1)
		})
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Misc", imgui.TreeNodeFlagsNone) {
		ui.Table("Misc Table", func() {
			ui.CheckboxRow("Kuwahara Filter", &runtimeConfig.KuwaharaFilter)
			ui.SliderIntRow("Fog Density", &runtimeConfig.FogDensity, 0, 500)
		})
	}

}
