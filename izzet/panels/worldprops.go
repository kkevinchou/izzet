package panels

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
)

type ComboOption string

const (
	ComboOptionFinalRender  ComboOption = "FINALRENDER"
	ComboOptionColorPicking ComboOption = "COLORPICKING"
	ComboOptionHDR          ComboOption = "HDR (bloom only)"
	ComboOptionBloom        ComboOption = "BLOOMTEXTURE (bloom only)"
	ComboOptionDepthMap     ComboOption = "DEPTH MAP"

	tableColumn0Width float32          = 250
	tableColumn1Width float32          = 400
	tableFlags        imgui.TableFlags = imgui.TableFlagsBordersInnerV
)

var SelectedComboOption ComboOption = ComboOptionFinalRender
var (
	comboOptions []ComboOption = []ComboOption{
		ComboOptionFinalRender,
		ComboOptionColorPicking,
		ComboOptionHDR,
		ComboOptionBloom,
		ComboOptionDepthMap,
	}
)

func worldProps(renderContext RenderContext) {
	if imgui.CollapsingHeaderV("Lighting", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Lights", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Ambient Factor", func() { imgui.SliderFloat("", &DBG.AmbientFactor, 0, 1) })
		setupRow("Point Light Bias", func() { imgui.SliderFloat("", &DBG.PointLightBias, 0, 1) })
		setupRow("Point Light Intensity", func() { imgui.InputInt("", &DBG.PointLightIntensity) })
		setupRow("Directional Light Intensity", func() { imgui.InputInt("", &DBG.DirectionalLightIntensity) })
		setupRow("Directional Light DIrection", func() { imgui.SliderFloat3("", &DBG.DirectionalLightDir, -1, 1) })
		setupRow("Color", func() { imgui.ColorEdit3V("", &DBG.Color, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel) })
		setupRow("Color Intensity", func() { imgui.SliderFloat("", &DBG.ColorIntensity, 0, 50) })
		setupRow("Enable Shadow Mapping", func() { imgui.Checkbox("", &DBG.EnableShadowMapping) })
		imgui.EndTable()
	}
	if imgui.CollapsingHeaderV("Bloom", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Enable Bloom", func() { imgui.Checkbox("", &DBG.Bloom) })
		setupRow("Bloom Intensity", func() { imgui.SliderFloat("", &DBG.BloomIntensity, 0, 1) })
		setupRow("Bloom Threshold Passes", func() { imgui.SliderInt("", &DBG.BloomThresholdPasses, 0, 3) })
		setupRow("Bloom Threshold", func() { imgui.SliderFloat("", &DBG.BloomThreshold, 0, 3) })
		setupRow("Upsampling Scale", func() { imgui.SliderFloat("", &DBG.BloomUpsamplingScale, 0, 5.0) })
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("Other", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		initColumns()
		setupRow("Roughness", func() { imgui.SliderFloat("", &DBG.Roughness, 0, 1) })
		setupRow("Metallic", func() { imgui.SliderFloat("", &DBG.Metallic, 0, 1) })
		setupRow("Exposure", func() { imgui.SliderFloat("", &DBG.Exposure, 0, 1) })
		setupRow("Material Override", func() { imgui.Checkbox("", &DBG.MaterialOverride) })
		setupRow("Render Spatial Partition", func() { imgui.Checkbox("", &DBG.RenderSpatialPartition) })
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("RenderStats", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)
		setupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%f", DBG.RenderTime)) })
		setupRow("Texture", func() {
			imgui.PushItemWidth(tableColumn1Width)
			if imgui.BeginCombo("", string(SelectedComboOption)) {
				for _, option := range comboOptions {
					if imgui.Selectable(string(option)) {
						SelectedComboOption = option
					}
				}
				imgui.EndCombo()
			}
			imgui.PopItemWidth()
		})
		setupRow("Texture Viewer", func() {
			if DBG.DebugTexture != 0 {
				var imageWidth float32 = 500
				texture := createUserSpaceTextureHandle(DBG.DebugTexture)
				size := imgui.Vec2{X: imageWidth, Y: imageWidth / float32(renderContext.AspectRatio())}
				// invert the Y axis since opengl vs texture coordinate systems differ
				// https://learnopengl.com/Getting-started/Textures
				imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
			}
		})
		imgui.EndTable()
	}
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
