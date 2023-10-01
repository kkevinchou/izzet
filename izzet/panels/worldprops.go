package panels

import (
	"fmt"

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

	tableColumn0Width float32          = 200
	tableColumn1Width float32          = 250
	tableFlags        imgui.TableFlags = imgui.TableFlagsBordersInnerV
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

func worldProps(world World, renderContext RenderContext) {
	if imgui.CollapsingHeaderV("Lighting", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Lights", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Ambient Factor", func() { imgui.SliderFloat("", &DBG.AmbientFactor, 0, 1) })
		setupRow("Point Light Bias", func() { imgui.SliderFloat("", &DBG.PointLightBias, 0, 1) })
		setupRow("Point Light Intensity", func() { imgui.InputInt("", &DBG.PointLightIntensity) })
		setupRow("Directional Light Intensity", func() { imgui.InputInt("", &DBG.DirectionalLightIntensity) })
		// setupRow("Color", func() { imgui.ColorEdit3V("", &DBG.Color, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel) })
		// setupRow("Color Intensity", func() { imgui.SliderFloat("", &DBG.ColorIntensity, 0, 50) })
		setupRow("Enable Shadow Mapping", func() { imgui.Checkbox("", &DBG.EnableShadowMapping) })
		setupRow("Shadow Far Factor", func() { imgui.SliderFloat("", &DBG.ShadowFarFactor, 0, 10) })
		setupRow("Fog Density", func() { imgui.SliderInt("", &DBG.FogDensity, 0, 100) })
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("Bloom", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Enable Bloom", func() { imgui.Checkbox("", &DBG.Bloom) })
		setupRow("Bloom Intensity", func() { imgui.SliderFloat("", &DBG.BloomIntensity, 0, 1) })
		setupRow("Bloom Threshold Passes", func() { imgui.SliderInt("", &DBG.BloomThresholdPasses, 0, 3) })
		setupRow("Bloom Threshold", func() { imgui.SliderFloat("", &DBG.BloomThreshold, 0, 3) })
		setupRow("Upsampling Scale", func() { imgui.SliderFloat("", &DBG.BloomUpsamplingScale, 0, 5.0) })
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("Rendering", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Rendering Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Far", func() { imgui.SliderFloat("", &DBG.Far, 0, 100000) })
		setupRow("FovX", func() { imgui.SliderFloat("", &DBG.FovX, 0, 170) })
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("NavMesh", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("NavMesh Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("NavMeshHSV", func() {
			if imgui.Checkbox("NavMeshHSV", &DBG.NavMeshHSV) {
				world.ResetNavMeshVAO()
			}
		})
		setupRow("NavMesh Region Threshold", func() {
			if imgui.InputInt("", &DBG.NavMeshRegionIDThreshold) {
				world.ResetNavMeshVAO()
			}
		})
		setupRow("NavMesh Distance Field Threshold", func() {
			if imgui.InputInt("", &DBG.NavMeshDistanceFieldThreshold) {
				world.ResetNavMeshVAO()
			}
		})
		setupRow("HSV Offset", func() {
			if imgui.InputInt("", &DBG.HSVOffset) {
				world.ResetNavMeshVAO()
			}
		})
		setupRow("Voxel Highlight X", func() {
			if imgui.InputInt("Voxel Highlight X", &DBG.VoxelHighlightX) {
				world.ResetNavMeshVAO()
			}
		})
		setupRow("Voxel Highlight Z", func() {
			if imgui.InputInt("Voxel Highlight Z", &DBG.VoxelHighlightZ) {
				world.ResetNavMeshVAO()
			}
		})
		setupRow("Highlight Distance Field", func() {
			imgui.LabelText("voxel highlight distance field", fmt.Sprintf("%f", DBG.VoxelHighlightDistanceField))
		})
		setupRow("Highlight Region ID", func() {
			imgui.LabelText("voxel highlight region field", fmt.Sprintf("%d", DBG.VoxelHighlightRegionID))
		})
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
		setupRow("Enable Spatial Partition", func() { imgui.Checkbox("", &DBG.EnableSpatialPartition) })
		setupRow("Render Spatial Partition", func() { imgui.Checkbox("", &DBG.RenderSpatialPartition) })
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

// createUserSpaceTextureHandle creates a handle to a user space texture
// that the imgui renderer is able to render
func createUserSpaceTextureHandle(texture uint32) imgui.TextureID {
	handle := 1<<63 | uint64(texture)
	return imgui.TextureID(handle)
}

func setupRow(label string, item func()) {
	imgui.TableNextRow()
	imgui.TableNextColumn()
	imgui.Text(label)
	imgui.TableNextColumn()
	imgui.PushID(label)
	imgui.PushItemWidth(-1)
	item()
	imgui.PopItemWidth()
	imgui.PopID()
}

func initColumns() {
	imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)
	imgui.TableSetupColumnV("1", imgui.TableColumnFlagsWidthFixed, tableColumn1Width, 0)
}
