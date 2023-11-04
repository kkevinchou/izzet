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
	settings := app.Settings()

	if imgui.CollapsingHeaderV("General", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("General Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()

		setupRow("Camera Position", func() {
			imgui.LabelText("Camera Position", fmt.Sprintf("{%.1f, %.1f, %.1f}", settings.CameraPosition[0], settings.CameraPosition[1], settings.CameraPosition[2]))
		}, true)

		setupRow("Camera Viewing Direction", func() {
			viewDir := settings.CameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
			imgui.LabelText("Camera Viewing Direction", fmt.Sprintf("{%.1f, %.1f, %.1f}", viewDir[0], viewDir[1], viewDir[2]))
		}, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("Lighting", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Lighting Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Ambient Factor", func() { imgui.SliderFloat("", &settings.AmbientFactor, 0, 1) }, true)
		setupRow("Point Light Bias", func() { imgui.SliderFloat("", &settings.PointLightBias, 0, 1) }, true)
		setupRow("Enable Shadow Mapping", func() { imgui.Checkbox("", &settings.EnableShadowMapping) }, true)
		setupRow("Shadow Far Factor", func() { imgui.SliderFloat("", &settings.ShadowFarFactor, 0, 10) }, true)
		setupRow("Fog Density", func() { imgui.SliderInt("", &settings.FogDensity, 0, 100) }, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("Bloom", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Enable Bloom", func() { imgui.Checkbox("", &settings.Bloom) }, true)
		setupRow("Bloom Intensity", func() { imgui.SliderFloat("", &settings.BloomIntensity, 0, 1) }, true)
		setupRow("Bloom Threshold Passes", func() { imgui.SliderInt("", &settings.BloomThresholdPasses, 0, 3) }, true)
		setupRow("Bloom Threshold", func() { imgui.SliderFloat("", &settings.BloomThreshold, 0, 3) }, true)
		setupRow("Upsampling Scale", func() { imgui.SliderFloat("", &settings.BloomUpsamplingScale, 0, 5.0) }, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("Rendering", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Rendering Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Far", func() { imgui.SliderFloat("", &settings.Far, 0, 100000) }, true)
		setupRow("FovX", func() { imgui.SliderFloat("", &settings.FovX, 0, 170) }, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("NavMesh", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("NavMesh Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("NavMeshHSV", func() {
			if imgui.Checkbox("NavMeshHSV", &settings.NavMeshHSV) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("NavMesh Region Threshold", func() {
			if imgui.InputInt("", &settings.NavMeshRegionIDThreshold) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("NavMesh Distance Field Threshold", func() {
			if imgui.InputInt("", &settings.NavMeshDistanceFieldThreshold) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("HSV Offset", func() {
			if imgui.InputInt("", &settings.HSVOffset) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("Voxel Highlight X", func() {
			if imgui.InputInt("Voxel Highlight X", &settings.VoxelHighlightX) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("Voxel Highlight Z", func() {
			if imgui.InputInt("Voxel Highlight Z", &settings.VoxelHighlightZ) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("Highlight Distance Field", func() {
			imgui.LabelText("voxel highlight distance field", fmt.Sprintf("%f", settings.VoxelHighlightDistanceField))
		}, true)
		setupRow("Highlight Region ID", func() {
			imgui.LabelText("voxel highlight region field", fmt.Sprintf("%d", settings.VoxelHighlightRegionID))
		}, true)
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("Other", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Other Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Enable Spatial Partition", func() { imgui.Checkbox("", &settings.EnableSpatialPartition) }, true)
		setupRow("Render Spatial Partition", func() { imgui.Checkbox("", &settings.RenderSpatialPartition) }, true)
		setupRow("Near Plane Offset", func() { imgui.SliderFloat("", &settings.SPNearPlaneOffset, 0, 1000) }, true)
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
