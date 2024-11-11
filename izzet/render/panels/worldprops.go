package panels

import (
	"fmt"
	"strconv"
	"strings"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

type NavMeshRenderComboOption string

const (
	ComboOptionNavMesh                NavMeshRenderComboOption = "Nav Mesh"
	ComboOptionCompactHeightField     NavMeshRenderComboOption = "Compact Height Field"
	ComboOptionDistanceField          NavMeshRenderComboOption = "Distance Field"
	ComboOptionVoxel                  NavMeshRenderComboOption = "Voxel"
	ComboOptionRawContour             NavMeshRenderComboOption = "Raw Contour"
	ComboOptionSimplifiedContour      NavMeshRenderComboOption = "Simplified Contour"
	ComboOptionDetailedMesh           NavMeshRenderComboOption = "Detailed Mesh"
	ComboOptionDetailedMeshAndSamples NavMeshRenderComboOption = "Detailed Mesh + Samples"
	ComboOptionPremergeTriangles      NavMeshRenderComboOption = "Premerge Triangles"
	ComboOptionPolygons               NavMeshRenderComboOption = "Polygons"
	ComboOptionDebug                  NavMeshRenderComboOption = "Debug"
)

var SelectedNavmeshRenderComboOption NavMeshRenderComboOption = ComboOptionNavMesh

var (
	navmeshRenderComboOptions []NavMeshRenderComboOption = []NavMeshRenderComboOption{
		ComboOptionNavMesh,
		ComboOptionCompactHeightField,
		ComboOptionDetailedMesh,
		ComboOptionDetailedMeshAndSamples,
		ComboOptionRawContour,
		ComboOptionSimplifiedContour,
		ComboOptionPolygons,
		ComboOptionPremergeTriangles,
		ComboOptionDebug,
		// ComboOptionDistanceField,
		// ComboOptionVoxel,
	}
)

var RecreateCloudTexture bool

var SelectedCloudTextureIndex int = 0
var SelectedCloudTextureChannelIndex int = 0

func worldProps(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("General Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("Camera Position", func() {
			imgui.LabelText("Camera Position", fmt.Sprintf("{%.1f, %.1f, %.1f}", runtimeConfig.CameraPosition[0], runtimeConfig.CameraPosition[1], runtimeConfig.CameraPosition[2]))
		}, true)

		panelutils.SetupRow("Camera Viewing Direction", func() {
			viewDir := runtimeConfig.CameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
			imgui.LabelText("Camera Viewing Direction", fmt.Sprintf("{%.1f, %.1f, %.1f}", viewDir[0], viewDir[1], viewDir[2]))
		}, true)
		panelutils.SetupRow("Enable Spatial Partition", func() { imgui.Checkbox("", &runtimeConfig.EnableSpatialPartition) }, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Editing", imgui.TreeNodeFlagsNone) {
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
	if imgui.CollapsingHeaderTreeNodeFlagsV("Rendering", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Rendering Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		panelutils.SetupRow("Near", func() { imgui.SliderFloat("", &runtimeConfig.Near, 0.1, 1) }, true)
		panelutils.SetupRow("Far", func() { imgui.SliderFloat("", &runtimeConfig.Far, 0, 1500) }, true)
		panelutils.SetupRow("FovX", func() { imgui.SliderFloat("", &runtimeConfig.FovX, 0, 170) }, true)

		panelutils.SetupRow("Debug Color", func() {
			imgui.ColorEdit3V("", &runtimeConfig.Color, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
		}, true)
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Navigation Mesh", imgui.TreeNodeFlagsNone) {
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
		panelutils.SetupRow("Max Edge Length", func() {
			var f float32 = float32(runtimeConfig.NavigationmeshMaxEdgeLength)
			if imgui.InputFloatV("", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
				runtimeConfig.NavigationmeshMaxEdgeLength = f
			}
		}, true)
		panelutils.SetupRow("Agent Radius", func() {
			var f float32 = float32(runtimeConfig.NavigationMeshAgentRadius)
			if imgui.InputFloatV("", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
				runtimeConfig.NavigationMeshAgentRadius = f
			}
		}, true)
		panelutils.SetupRow("Cell Size", func() {
			f := runtimeConfig.NavigationMeshCellSize
			if imgui.InputFloatV("", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
				runtimeConfig.NavigationMeshCellSize = f
			}
		}, true)
		panelutils.SetupRow("Cell Height", func() {
			f := runtimeConfig.NavigationMeshCellHeight
			if imgui.InputFloatV("", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
				runtimeConfig.NavigationMeshCellHeight = f
			}
		}, true)
		panelutils.SetupRow("Sample Dist", func() {
			var f float32 = float32(runtimeConfig.NavigationmeshSampleDist)
			if imgui.InputFloatV("", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
				runtimeConfig.NavigationmeshSampleDist = f
			}
		}, true)
		panelutils.SetupRow("Filter Ledge Spans", func() {
			imgui.Checkbox("##", &runtimeConfig.NavigationMeshFilterLedgeSpans)
		}, true)
		panelutils.SetupRow("Filter Low Height Spans", func() {
			imgui.Checkbox("##", &runtimeConfig.NavigationMeshFilterLowHeightSpans)
		}, true)
		imgui.EndTable()

		if imgui.InputTextWithHint("##DebugBlob1", "", &runtimeConfig.DebugBlob1, imgui.InputTextFlagsNone, nil) {
			runtimeConfig.DebugBlob1IntMap = map[int]bool{}
			runtimeConfig.DebugBlob1IntSlice = nil
			sIDs := strings.Split(runtimeConfig.DebugBlob1, ",")
			for _, sID := range sIDs {
				id, err := strconv.Atoi(sID)
				if err != nil {
					continue
				}
				runtimeConfig.DebugBlob1IntMap[id] = true
				runtimeConfig.DebugBlob1IntSlice = append(runtimeConfig.DebugBlob1IntSlice, id)
			}
		}
		if imgui.InputTextWithHint("##DebugBlob2", "", &runtimeConfig.DebugBlob2, imgui.InputTextFlagsNone, nil) {
			runtimeConfig.DebugBlob2IntMap = map[int]bool{}
			runtimeConfig.DebugBlob2IntSlice = nil
			sIDs := strings.Split(runtimeConfig.DebugBlob2, ",")
			for _, sID := range sIDs {
				id, err := strconv.Atoi(sID)
				if err != nil {
					continue
				}
				runtimeConfig.DebugBlob2IntMap[id] = true
				runtimeConfig.DebugBlob2IntSlice = append(runtimeConfig.DebugBlob2IntSlice, id)
			}

		}
		if imgui.Button("Build") {
			iterations := int(runtimeConfig.NavigationMeshIterations)
			walkableHeight := int(runtimeConfig.NavigationMeshWalkableHeight)
			climbableHeight := int(runtimeConfig.NavigationMeshClimbableHeight)
			minRegionArea := int(runtimeConfig.NavigationMeshMinRegionArea)
			maxError := float64(runtimeConfig.NavigationmeshMaxError)
			sampleDist := float64(runtimeConfig.NavigationmeshSampleDist)
			app.BuildNavMesh(app, iterations, walkableHeight, climbableHeight, minRegionArea, sampleDist, maxError)
		}

		imgui.BeginTableV("Navigation Mesh Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.SetupRow("Start", func() {
			var i int32 = runtimeConfig.NavigationMeshStart
			if imgui.InputInt("##", &i) {
				runtimeConfig.NavigationMeshStart = i
			}
		}, true)
		panelutils.SetupRow("Goal", func() {
			var i int32 = int32(runtimeConfig.NavigationMeshGoal)
			if imgui.InputInt("##", &i) {
				runtimeConfig.NavigationMeshGoal = i
			}
		}, true)
		imgui.EndTable()

		if imgui.Button("Find Path") {
			app.FindPath(app.RuntimeConfig().NavigationMeshStartPoint, app.RuntimeConfig().NavigationMeshGoalPoint)
		}

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

	if imgui.CollapsingHeaderTreeNodeFlagsV("Noise", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Noise Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("Cloud Texture Index", func() {
			if imgui.BeginCombo("##", fmt.Sprintf("%d", SelectedCloudTextureIndex)) {
				for i := range 2 {
					if imgui.SelectableBool(fmt.Sprintf("%d", i)) {
						SelectedCloudTextureIndex = i
						app.RuntimeConfig().ActiveCloudTextureIndex = int(i)
					}
				}
				imgui.EndCombo()
			}
		}, true)

		panelutils.SetupRow("Cloud Channel Index", func() {
			if imgui.BeginCombo("##", fmt.Sprintf("%d", SelectedCloudTextureChannelIndex)) {
				for i := range 4 {
					if imgui.SelectableBool(fmt.Sprintf("%d", i)) {
						SelectedCloudTextureChannelIndex = i
						app.RuntimeConfig().ActiveCloudTextureChannelIndex = int(i)
					}
				}
				imgui.EndCombo()
			}
		}, true)

		cloudTexture := &app.RuntimeConfig().CloudTextures[app.RuntimeConfig().ActiveCloudTextureIndex]
		activeChannelIndex := &app.RuntimeConfig().ActiveCloudTextureChannelIndex
		panelutils.SetupRow("Noise Z", func() {
			imgui.SliderFloatV("noiseZ", &cloudTexture.Channels[*activeChannelIndex].NoiseZ, 0, 1, "%.3f", imgui.SliderFlagsNone)
		}, true)
		panelutils.SetupRow("Cell Width", func() {
			if imgui.SliderInt("cellWidth", &cloudTexture.Channels[*activeChannelIndex].CellWidth, 1, 30) {
				RecreateCloudTexture = true
			}
		}, true)
		panelutils.SetupRow("Cell Height", func() {
			if imgui.SliderInt("cellHeight", &cloudTexture.Channels[*activeChannelIndex].CellHeight, 1, 30) {
				RecreateCloudTexture = true
			}
		}, true)
		panelutils.SetupRow("Cell Depth", func() {
			if imgui.SliderInt("cellDepth", &cloudTexture.Channels[*activeChannelIndex].CellDepth, 1, 30) {
				RecreateCloudTexture = true
			}
		}, true)
		panelutils.SetupRow("WGroup Width", func() {
			if imgui.SliderInt("workGroupWidth", &cloudTexture.WorkGroupWidth, 1, 512) {
				RecreateCloudTexture = true
			}
		}, true)
		panelutils.SetupRow("WGroup Height", func() {
			if imgui.SliderInt("workGroupHeight", &cloudTexture.WorkGroupHeight, 1, 512) {
				RecreateCloudTexture = true
			}
		}, true)
		panelutils.SetupRow("WGroup Depth", func() {
			if imgui.SliderInt("workGroupDepth", &cloudTexture.WorkGroupDepth, 1, 512) {
				RecreateCloudTexture = true
			}
		}, true)
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
