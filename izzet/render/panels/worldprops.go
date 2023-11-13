package panels

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

type ComboOption string

const (
	ComboOptionFinalRender    ComboOption = "FINALRENDER"
	ComboOptionColorPicking   ComboOption = "COLORPICKING"
	ComboOptionHDR            ComboOption = "HDR (bloom only)"
	ComboOptionBloom          ComboOption = "BLOOMTEXTURE (bloom only)"
	ComboOptionShadowDepthMap ComboOption = "SHADOW DEPTH MAP"
	ComboOptionCameraDepthMap ComboOption = "CAMERA DEPTH MAP"
	ComboOptionCubeDepthMap   ComboOption = "CUBE DEPTH MAP"
)

var SelectedComboOption ComboOption = ComboOptionFinalRender
var (
	comboOptions []ComboOption = []ComboOption{
		ComboOptionFinalRender,
		ComboOptionColorPicking,
		ComboOptionHDR,
		ComboOptionBloom,
		ComboOptionShadowDepthMap,
		ComboOptionCameraDepthMap,
	}
)

func worldProps(app renderiface.App, renderContext RenderContext) {
	runtimeConfig := app.RuntimeConfig()

	if imgui.CollapsingHeaderV("General", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("General Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()

		setupRow("Camera Position", func() {
			imgui.LabelText("Camera Position", fmt.Sprintf("{%.1f, %.1f, %.1f}", runtimeConfig.CameraPosition[0], runtimeConfig.CameraPosition[1], runtimeConfig.CameraPosition[2]))
		}, true)

		setupRow("Camera Viewing Direction", func() {
			viewDir := runtimeConfig.CameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
			imgui.LabelText("Camera Viewing Direction", fmt.Sprintf("{%.1f, %.1f, %.1f}", viewDir[0], viewDir[1], viewDir[2]))
		}, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("Lighting", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Lighting Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Ambient Factor", func() { imgui.SliderFloat("", &runtimeConfig.AmbientFactor, 0, 1) }, true)
		setupRow("Point Light Bias", func() { imgui.SliderFloat("", &runtimeConfig.PointLightBias, 0, 1) }, true)
		setupRow("Enable Shadow Mapping", func() { imgui.Checkbox("", &runtimeConfig.EnableShadowMapping) }, true)
		setupRow("Shadow Far Factor", func() { imgui.SliderFloat("", &runtimeConfig.ShadowFarFactor, 0, 10) }, true)
		setupRow("Fog Density", func() { imgui.SliderInt("", &runtimeConfig.FogDensity, 0, 100) }, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("Bloom", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Enable Bloom", func() { imgui.Checkbox("", &runtimeConfig.Bloom) }, true)
		setupRow("Bloom Intensity", func() { imgui.SliderFloat("", &runtimeConfig.BloomIntensity, 0, 1) }, true)
		setupRow("Bloom Threshold Passes", func() { imgui.SliderInt("", &runtimeConfig.BloomThresholdPasses, 0, 3) }, true)
		setupRow("Bloom Threshold", func() { imgui.SliderFloat("", &runtimeConfig.BloomThreshold, 0, 3) }, true)
		setupRow("Upsampling Scale", func() { imgui.SliderFloat("", &runtimeConfig.BloomUpsamplingScale, 0, 5.0) }, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("Rendering", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Rendering Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Far", func() { imgui.SliderFloat("", &runtimeConfig.Far, 0, 100000) }, true)
		setupRow("FovX", func() { imgui.SliderFloat("", &runtimeConfig.FovX, 0, 170) }, true)

		setupRow("Color", func() {
			imgui.ColorEdit3V("", &runtimeConfig.Color, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
		}, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("NavMesh", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("NavMesh Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("NavMeshHSV", func() {
			if imgui.Checkbox("NavMeshHSV", &runtimeConfig.NavMeshHSV) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("NavMesh Region Threshold", func() {
			if imgui.InputInt("", &runtimeConfig.NavMeshRegionIDThreshold) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("NavMesh Distance Field Threshold", func() {
			if imgui.InputInt("", &runtimeConfig.NavMeshDistanceFieldThreshold) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("HSV Offset", func() {
			if imgui.InputInt("", &runtimeConfig.HSVOffset) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("Voxel Highlight X", func() {
			if imgui.InputInt("Voxel Highlight X", &runtimeConfig.VoxelHighlightX) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("Voxel Highlight Z", func() {
			if imgui.InputInt("Voxel Highlight Z", &runtimeConfig.VoxelHighlightZ) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("Highlight Distance Field", func() {
			imgui.LabelText("voxel highlight distance field", fmt.Sprintf("%f", runtimeConfig.VoxelHighlightDistanceField))
		}, true)
		setupRow("Highlight Region ID", func() {
			imgui.LabelText("voxel highlight region field", fmt.Sprintf("%d", runtimeConfig.VoxelHighlightRegionID))
		}, true)
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("Other", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Other Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Enable Spatial Partition", func() { imgui.Checkbox("", &runtimeConfig.EnableSpatialPartition) }, true)
		setupRow("Render Spatial Partition", func() { imgui.Checkbox("", &runtimeConfig.RenderSpatialPartition) }, true)
		setupRow("Near Plane Offset", func() { imgui.SliderFloat("", &runtimeConfig.SPNearPlaneOffset, 0, 1000) }, true)
		imgui.EndTable()
	}
}

func formatNumber(number int) string {
	s := fmt.Sprintf("%d", number)

	var fmtStr string
	runCount := len(s) / 3

	for i := 0; i < runCount; i++ {
		index := len(s) - ((i + 1) * 3)
		fmtStr = "," + s[index:index+3] + fmtStr
	}

	remainder := len(s) - 3*runCount
	if remainder > 0 {
		fmtStr = s[:remainder] + fmtStr
	} else {
		// trim leading separator
		fmtStr = fmtStr[1:]
	}

	return fmtStr
}
