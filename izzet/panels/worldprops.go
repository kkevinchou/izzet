package panels

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
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

func worldProps(app App, renderContext RenderContext) {
	if imgui.CollapsingHeaderV("General", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("General Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()

		setupRow("Camera Position", func() {
			imgui.LabelText("Camera Position", fmt.Sprintf("{%.1f, %.1f, %.1f}", DBG.CameraPosition[0], DBG.CameraPosition[1], DBG.CameraPosition[2]))
		}, true)

		setupRow("Camera Viewing Direction", func() {
			viewDir := DBG.CameraOrientation.Rotate(mgl64.Vec3{0, 0, -1})
			imgui.LabelText("Camera Viewing Direction", fmt.Sprintf("{%.1f, %.1f, %.1f}", viewDir[0], viewDir[1], viewDir[2]))
		}, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("Lighting", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Lighting Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Ambient Factor", func() { imgui.SliderFloat("", &DBG.AmbientFactor, 0, 1) }, true)
		setupRow("Point Light Bias", func() { imgui.SliderFloat("", &DBG.PointLightBias, 0, 1) }, true)
		// setupRow("Color", func() { imgui.ColorEdit3V("", &DBG.Color, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel) })
		// setupRow("Color Intensity", func() { imgui.SliderFloat("", &DBG.ColorIntensity, 0, 50) })
		setupRow("Enable Shadow Mapping", func() { imgui.Checkbox("", &DBG.EnableShadowMapping) }, true)
		setupRow("Shadow Far Factor", func() { imgui.SliderFloat("", &DBG.ShadowFarFactor, 0, 10) }, true)
		setupRow("Fog Density", func() { imgui.SliderInt("", &DBG.FogDensity, 0, 100) }, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("Bloom", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Enable Bloom", func() { imgui.Checkbox("", &DBG.Bloom) }, true)
		setupRow("Bloom Intensity", func() { imgui.SliderFloat("", &DBG.BloomIntensity, 0, 1) }, true)
		setupRow("Bloom Threshold Passes", func() { imgui.SliderInt("", &DBG.BloomThresholdPasses, 0, 3) }, true)
		setupRow("Bloom Threshold", func() { imgui.SliderFloat("", &DBG.BloomThreshold, 0, 3) }, true)
		setupRow("Upsampling Scale", func() { imgui.SliderFloat("", &DBG.BloomUpsamplingScale, 0, 5.0) }, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("Rendering", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Rendering Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Far", func() { imgui.SliderFloat("", &DBG.Far, 0, 100000) }, true)
		setupRow("FovX", func() { imgui.SliderFloat("", &DBG.FovX, 0, 170) }, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("NavMesh", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("NavMesh Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("NavMeshHSV", func() {
			if imgui.Checkbox("NavMeshHSV", &DBG.NavMeshHSV) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("NavMesh Region Threshold", func() {
			if imgui.InputInt("", &DBG.NavMeshRegionIDThreshold) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("NavMesh Distance Field Threshold", func() {
			if imgui.InputInt("", &DBG.NavMeshDistanceFieldThreshold) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("HSV Offset", func() {
			if imgui.InputInt("", &DBG.HSVOffset) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("Voxel Highlight X", func() {
			if imgui.InputInt("Voxel Highlight X", &DBG.VoxelHighlightX) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("Voxel Highlight Z", func() {
			if imgui.InputInt("Voxel Highlight Z", &DBG.VoxelHighlightZ) {
				app.ResetNavMeshVAO()
			}
		}, true)
		setupRow("Highlight Distance Field", func() {
			imgui.LabelText("voxel highlight distance field", fmt.Sprintf("%f", DBG.VoxelHighlightDistanceField))
		}, true)
		setupRow("Highlight Region ID", func() {
			imgui.LabelText("voxel highlight region field", fmt.Sprintf("%d", DBG.VoxelHighlightRegionID))
		}, true)
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("Other", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Other Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		// setupRow("Roughness", func() { imgui.SliderFloat("", &DBG.Roughness, 0, 1) })
		// setupRow("Roughness", func() { imgui.SliderFloat("", &DBG.Roughness, 0, 1) })
		// setupRow("Metallic", func() { imgui.SliderFloat("", &DBG.Metallic, 0, 1) })
		// setupRow("Exposure", func() { imgui.SliderFloat("", &DBG.Exposure, 0, 1) })
		// setupRow("Material Override", func() { imgui.Checkbox("", &DBG.MaterialOverride) })
		setupRow("Enable Spatial Partition", func() { imgui.Checkbox("", &DBG.EnableSpatialPartition) }, true)
		setupRow("Render Spatial Partition", func() { imgui.Checkbox("", &DBG.RenderSpatialPartition) }, true)
		setupRow("Near Plane Offset", func() { imgui.SliderFloat("", &DBG.SPNearPlaneOffset, 0, 1000) }, true)
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
