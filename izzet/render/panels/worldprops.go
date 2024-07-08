package panels

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

type NavMeshRenderComboOption string

const (
	ComboOptionCompactHeightField NavMeshRenderComboOption = "Compact Height Field"
	ComboOptionVoxel              NavMeshRenderComboOption = "Voxel"
	ComboOptionSimplifiedContour  NavMeshRenderComboOption = "Simplified Contour"
)

var SelectedNavmeshRenderComboOption NavMeshRenderComboOption = ComboOptionCompactHeightField

var (
	navmeshRenderComboOptions []NavMeshRenderComboOption = []NavMeshRenderComboOption{
		ComboOptionCompactHeightField,
		ComboOptionVoxel,
		ComboOptionSimplifiedContour,
	}
)

func worldProps(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("General Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("Camera Position", func() {
			imgui.LabelText("Camera Position", fmt.Sprintf("{%.1f, %.1f, %.1f}", runtimeConfig.CameraPosition[0], runtimeConfig.CameraPosition[1], runtimeConfig.CameraPosition[2]))
		}, true)

		panelutils.SetupRow("Camera Viewing Direction", func() {
			viewDir := runtimeConfig.CameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
			imgui.LabelText("Camera Viewing Direction", fmt.Sprintf("{%.1f, %.1f, %.1f}", viewDir[0], viewDir[1], viewDir[2]))
		}, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Editing", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Editing Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		panelutils.SetupRow("Grid Snapping Size", func() {
			if imgui.InputIntV("", &runtimeConfig.SnapSize, 0, 0, imgui.InputTextFlagsNone) {
				if runtimeConfig.SnapSize < 1 {
					runtimeConfig.SnapSize = 1
				}
			}
		}, true)
		panelutils.SetupRow("Rotation Snapping Size", func() {
			if imgui.InputIntV("", &runtimeConfig.RotationSnapSize, 0, 0, imgui.InputTextFlagsNone) {
				if runtimeConfig.RotationSnapSize < 1 {
					runtimeConfig.RotationSnapSize = 1
				}
			}
		}, true)
		panelutils.SetupRow("Rotation Sensitivity", func() {
			if imgui.InputIntV("", &runtimeConfig.RotationSensitivity, 0, 0, imgui.InputTextFlagsNone) {
				if runtimeConfig.RotationSensitivity < 1 {
					runtimeConfig.RotationSensitivity = 1
				}
			}
		}, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Lighting", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Lighting Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		panelutils.SetupRow("Ambient Factor", func() { imgui.SliderFloat("", &runtimeConfig.AmbientFactor, 0, 1) }, true)
		panelutils.SetupRow("Point Light Bias", func() { imgui.SliderFloat("", &runtimeConfig.PointLightBias, 0, 1) }, true)
		panelutils.SetupRow("Enable Shadow Mapping", func() { imgui.Checkbox("", &runtimeConfig.EnableShadowMapping) }, true)
		panelutils.SetupRow("Shadow Far Factor", func() { imgui.SliderFloat("", &runtimeConfig.ShadowFarFactor, 0, 10) }, true)
		panelutils.SetupRow("Fog Density", func() { imgui.SliderInt("", &runtimeConfig.FogDensity, 0, 100) }, true)
		panelutils.SetupRow("Enable Bloom", func() { imgui.Checkbox("", &runtimeConfig.Bloom) }, true)
		panelutils.SetupRow("Bloom Intensity", func() { imgui.SliderFloat("", &runtimeConfig.BloomIntensity, 0, 1) }, true)
		panelutils.SetupRow("Bloom Threshold Passes", func() { imgui.SliderInt("", &runtimeConfig.BloomThresholdPasses, 0, 3) }, true)
		panelutils.SetupRow("Bloom Threshold", func() { imgui.SliderFloat("", &runtimeConfig.BloomThreshold, 0, 3) }, true)
		panelutils.SetupRow("Bloom Upsampling Scale", func() { imgui.SliderFloat("", &runtimeConfig.BloomUpsamplingScale, 0, 5.0) }, true)
		panelutils.SetupRow("Shadow Map Z Offset", func() { imgui.SliderFloat("", &runtimeConfig.ShadowmapZOffset, 0, 2000) }, true)
		panelutils.SetupRow("SP Near Plane Offset", func() { imgui.SliderFloat("", &runtimeConfig.ShadowSpatialPartitionNearPlane, 0, 2000) }, true)
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderTreeNodeFlagsV("Rendering", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Rendering Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		panelutils.SetupRow("Far", func() { imgui.SliderFloat("", &runtimeConfig.Far, 0, 100000) }, true)
		panelutils.SetupRow("FovX", func() { imgui.SliderFloat("", &runtimeConfig.FovX, 0, 170) }, true)

		panelutils.SetupRow("Debug Color", func() {
			imgui.ColorEdit3V("", &runtimeConfig.Color, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
		}, true)
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Navigation Mesh", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Navigation Mesh Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.SetupRow("Iterations", func() {
			var i int32 = runtimeConfig.NavigationMeshIterations
			if imgui.InputInt("", &i) {
				runtimeConfig.NavigationMeshIterations = i
			}
		}, true)
		panelutils.SetupRow("Walkable Height", func() {
			var i int32 = runtimeConfig.NavigationMeshWalkableHeight
			if imgui.InputInt("", &i) {
				runtimeConfig.NavigationMeshWalkableHeight = i
			}
		}, true)
		panelutils.SetupRow("Climbable Height", func() {
			var i int32 = runtimeConfig.NavigationMeshClimbableHeight
			if imgui.InputInt("", &i) {
				runtimeConfig.NavigationMeshClimbableHeight = i
			}
		}, true)
		panelutils.SetupRow("Min Region Area", func() {
			var i int32 = runtimeConfig.NavigationMeshMinRegionArea
			if imgui.InputInt("", &i) {
				runtimeConfig.NavigationMeshMinRegionArea = i
			}
		}, true)
		panelutils.SetupRow("Max Error", func() {
			var f float32 = float32(runtimeConfig.NavigationmeshMaxError)
			if imgui.InputFloatV("", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
				runtimeConfig.NavigationmeshMaxError = f
			}
		}, true)
		if imgui.Button("Build") {
			iterations := int(runtimeConfig.NavigationMeshIterations)
			walkableHeight := int(runtimeConfig.NavigationMeshWalkableHeight)
			climbableHeight := int(runtimeConfig.NavigationMeshClimbableHeight)
			minRegionArea := int(runtimeConfig.NavigationMeshMinRegionArea)
			maxError := float64(runtimeConfig.NavigationmeshMaxError)
			app.BuildNavMesh(app, iterations, walkableHeight, climbableHeight, minRegionArea, maxError)
		}
		imgui.EndTable()

		imgui.LabelText("##", "Draw")
		if imgui.BeginCombo("##", string(SelectedNavmeshRenderComboOption)) {
			for _, option := range navmeshRenderComboOptions {
				if imgui.SelectableBool(string(option)) {
					SelectedNavmeshRenderComboOption = option
				}
			}
			imgui.EndCombo()
		}
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Other", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Other Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		panelutils.SetupRow("Enable Spatial Partition", func() { imgui.Checkbox("", &runtimeConfig.EnableSpatialPartition) }, true)
		panelutils.SetupRow("Render Spatial Partition", func() { imgui.Checkbox("", &runtimeConfig.RenderSpatialPartition) }, true)
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
