package panels

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/ui"
)

type NavMeshRenderComboOption string

const (
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

var SelectedNavmeshRenderComboOption NavMeshRenderComboOption = ComboOptionDetailedMesh

var (
	navmeshRenderComboOptions []NavMeshRenderComboOption = []NavMeshRenderComboOption{
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

func WorldProps(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsDefaultOpen) {
		ui.Table("General Table", func() {
			ui.LabelRow("Camera Position", fmt.Sprintf("{%.1f, %.1f, %.1f}", runtimeConfig.CameraPosition[0], runtimeConfig.CameraPosition[1], runtimeConfig.CameraPosition[2]))
			viewDir := runtimeConfig.CameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
			ui.LabelRow("Camera Viewing Direction", fmt.Sprintf("{%.1f, %.1f, %.1f}", viewDir[0], viewDir[1], viewDir[2]))
		})
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Editing", imgui.TreeNodeFlagsNone) {
		ui.Table("Editing Table", func() {
			ui.Row("Grid Snapping Size", func() {
				var value float32 = float32(app.RuntimeConfig().SnapSize)
				if imgui.SliderFloatV("noiseZ", &value, 0.1, 2, "%.3f", imgui.SliderFlagsNone) {
					app.RuntimeConfig().SnapSize = float64(value)
				}
			})
			ui.Row("Rotation Snapping Size", func() {
				if imgui.InputIntV("##value", &runtimeConfig.RotationSnapSize, 0, 0, imgui.InputTextFlagsNone) {
					if runtimeConfig.RotationSnapSize < 1 {
						runtimeConfig.RotationSnapSize = 1
					}
				}
			})
			ui.Row("Rotation Sensitivity", func() {
				if imgui.InputIntV("##value", &runtimeConfig.RotationSensitivity, 0, 0, imgui.InputTextFlagsNone) {
					if runtimeConfig.RotationSensitivity < 1 {
						runtimeConfig.RotationSensitivity = 1
					}
				}
			})
		})
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Navigation Mesh", imgui.TreeNodeFlagsNone) {
		ui.Table("Navigation Mesh Table", func() {
			ui.Row("Iterations", func() {
				var i int32 = runtimeConfig.NavigationMeshIterations
				if imgui.InputInt("##value", &i) {
					runtimeConfig.NavigationMeshIterations = i
				}
			})
			ui.Row("Walkable Height (World)", func() {
				f := runtimeConfig.NavigationMeshWalkableHeight
				if imgui.InputFloatV("##value", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
					runtimeConfig.NavigationMeshWalkableHeight = f
				}
			})
			ui.Row("Climbable Height (World)", func() {
				f := runtimeConfig.NavigationMeshClimbableHeight
				if imgui.InputFloatV("##value", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
					runtimeConfig.NavigationMeshClimbableHeight = f
				}
			})
			ui.Row("Agent Radius (World)", func() {
				var f float32 = float32(runtimeConfig.NavigationMeshAgentRadius)
				if imgui.InputFloatV("##value", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
					runtimeConfig.NavigationMeshAgentRadius = f
				}
			})
			ui.Row("Min Region Area", func() {
				var i int32 = runtimeConfig.NavigationMeshMinRegionArea
				if imgui.InputInt("##value", &i) {
					runtimeConfig.NavigationMeshMinRegionArea = i
				}
			})
			ui.Row("Max Error", func() {
				var f float32 = float32(runtimeConfig.NavigationmeshMaxError)
				if imgui.InputFloatV("##value", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
					runtimeConfig.NavigationmeshMaxError = f
				}
			})
			ui.Row("Max Edge Length", func() {
				var f float32 = float32(runtimeConfig.NavigationmeshMaxEdgeLength)
				if imgui.InputFloatV("##value", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
					runtimeConfig.NavigationmeshMaxEdgeLength = f
				}
			})
			ui.Row("Cell Size", func() {
				f := runtimeConfig.NavigationMeshCellSize
				if imgui.InputFloatV("##value", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
					runtimeConfig.NavigationMeshCellSize = f
				}
			})
			ui.Row("Cell Height", func() {
				f := runtimeConfig.NavigationMeshCellHeight
				if imgui.InputFloatV("##value", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
					runtimeConfig.NavigationMeshCellHeight = f
				}
			})
			ui.Row("Sample Dist", func() {
				var f float32 = float32(runtimeConfig.NavigationmeshSampleDist)
				if imgui.InputFloatV("##value", &f, 0.1, 0.1, "%.1f", imgui.InputTextFlagsNone) {
					runtimeConfig.NavigationmeshSampleDist = f
				}
			})
			ui.CheckboxRow("Filter Ledge Spans", &runtimeConfig.NavigationMeshFilterLedgeSpans)
			ui.CheckboxRow("Filter Low Height Spans", &runtimeConfig.NavigationMeshFilterLowHeightSpans)
		})

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
			walkableHeight := runtimeConfig.NavigationMeshWalkableHeight
			climbableHeight := runtimeConfig.NavigationMeshClimbableHeight
			minRegionArea := int(runtimeConfig.NavigationMeshMinRegionArea)
			maxError := float64(runtimeConfig.NavigationmeshMaxError)
			sampleDist := float64(runtimeConfig.NavigationmeshSampleDist)
			app.BuildNavMesh(app, iterations, walkableHeight, climbableHeight, minRegionArea, sampleDist, maxError)
		}

		ui.Table("Navigation Mesh Table", func() {
			ui.Row("Start", func() {
				var i int32 = runtimeConfig.NavigationMeshStart
				if imgui.InputInt("##value", &i) {
					runtimeConfig.NavigationMeshStart = i
				}
			})
			ui.Row("Goal", func() {
				var i int32 = int32(runtimeConfig.NavigationMeshGoal)
				if imgui.InputInt("##value", &i) {
					runtimeConfig.NavigationMeshGoal = i
				}
			})
		})

		if imgui.Button("Find Path") {
			app.FindPath(app.RuntimeConfig().NavigationMeshStartPoint, app.RuntimeConfig().NavigationMeshGoalPoint)
		}
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Noise", imgui.TreeNodeFlagsNone) {
		ui.Table("Noise Table", func() {
			ui.Row("Cloud Texture Index", func() {
				if imgui.BeginCombo("##value", fmt.Sprintf("%d", SelectedCloudTextureIndex)) {
					for i := range 2 {
						if imgui.SelectableBool(fmt.Sprintf("%d", i)) {
							SelectedCloudTextureIndex = i
							app.RuntimeConfig().ActiveCloudTextureIndex = int(i)
						}
					}
					imgui.EndCombo()
				}
			})

			ui.Row("Cloud Channel Index", func() {
				if imgui.BeginCombo("##value", fmt.Sprintf("%d", SelectedCloudTextureChannelIndex)) {
					for i := range 4 {
						if imgui.SelectableBool(fmt.Sprintf("%d", i)) {
							SelectedCloudTextureChannelIndex = i
							app.RuntimeConfig().ActiveCloudTextureChannelIndex = int(i)
						}
					}
					imgui.EndCombo()
				}
			})

			cloudTexture := &app.RuntimeConfig().CloudTextures[app.RuntimeConfig().ActiveCloudTextureIndex]
			activeChannelIndex := &app.RuntimeConfig().ActiveCloudTextureChannelIndex
			ui.Row("Noise Z", func() {
				imgui.SliderFloatV("##value", &cloudTexture.Channels[*activeChannelIndex].NoiseZ, 0, 1, "%.3f", imgui.SliderFlagsNone)
			})
			ui.Row("Cell Width", func() {
				if imgui.SliderInt("##value", &cloudTexture.Channels[*activeChannelIndex].CellWidth, 1, 30) {
					RecreateCloudTexture = true
				}
			})
			ui.Row("Cell Height", func() {
				if imgui.SliderInt("##value", &cloudTexture.Channels[*activeChannelIndex].CellHeight, 1, 30) {
					RecreateCloudTexture = true
				}
			})
			ui.Row("Cell Depth", func() {
				if imgui.SliderInt("##value", &cloudTexture.Channels[*activeChannelIndex].CellDepth, 1, 30) {
					RecreateCloudTexture = true
				}
			})
			ui.Row("WGroup Width", func() {
				if imgui.SliderInt("##value", &cloudTexture.WorkGroupWidth, 1, 512) {
					RecreateCloudTexture = true
				}
			})
			ui.Row("WGroup Height", func() {
				if imgui.SliderInt("##value", &cloudTexture.WorkGroupHeight, 1, 512) {
					RecreateCloudTexture = true
				}
			})
			ui.Row("WGroup Depth", func() {
				if imgui.SliderInt("##value", &cloudTexture.WorkGroupDepth, 1, 512) {
					RecreateCloudTexture = true
				}
			})
		})
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

func inputVec3(v *mgl64.Vec3) {
	x, y, z := float32(v.X()), float32(v.Y()), float32(v.Z())

	imgui.PushItemWidth(imgui.ContentRegionAvail().X / 3.0)
	if imgui.InputFloatV("##x", &x, 0, 0, "%.2f", imgui.InputTextFlagsNone) {
		v[0] = float64(x)
	}
	imgui.SameLine()
	if imgui.InputFloatV("##y", &y, 0, 0, "%.2f", imgui.InputTextFlagsNone) {
		v[1] = float64(y)
	}
	imgui.SameLine()
	if imgui.InputFloatV("##z", &z, 0, 0, "%.2f", imgui.InputTextFlagsNone) {
		v[2] = float64(z)
	}
	imgui.PopItemWidth()
}
