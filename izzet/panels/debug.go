package panels

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
)

type DebugSettings struct {
	DirectionalLightDir       [3]float32
	Roughness                 float32
	Metallic                  float32
	PointLightIntensity       int32
	DirectionalLightIntensity int32
	PointLightBias            float32
	MaterialOverride          bool
	EnableShadowMapping       bool
	DebugTexture              uint32 // 64 bits as we need extra bits to specify a the type of texture to IMGUI
	BloomIntensity            float32
	Exposure                  float32
	AmbientFactor             float32
	Bloom                     bool
	BloomThresholdPasses      int32
	BloomThreshold            float32
	BloomUpsamplingScale      float32
	Color                     [3]float32
	ColorIntensity            float32

	RenderSpatialPartition bool

	RenderTime float64
}

var DBG DebugSettings = DebugSettings{
	DirectionalLightDir:       [3]float32{0, -1, -1},
	Roughness:                 0.55,
	Metallic:                  1.0,
	PointLightIntensity:       100,
	DirectionalLightIntensity: 10,
	PointLightBias:            1,
	MaterialOverride:          false,
	EnableShadowMapping:       true,
	BloomIntensity:            0.04,
	Exposure:                  1.0,
	AmbientFactor:             0.001,
	Bloom:                     true,
	BloomThresholdPasses:      0,
	BloomThreshold:            0.8,
	BloomUpsamplingScale:      1.0,
	Color:                     [3]float32{1, 1, 1},
	ColorIntensity:            1.0,
	RenderSpatialPartition:    false,
}

func BuildDebug(world World, renderContext RenderContext) {
	if !ShowDebug {
		return
	}

	// drawHUDTextureToQuad(cameraViewerContext, r.shaderManager.GetShaderProgram("depthDebug"), r.shadowMap.depthTexture, 1)

	imgui.SetNextWindowPosV(imgui.Vec2{X: 400, Y: 400}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 500, Y: 300}, imgui.ConditionFirstUseEver)

	imgui.BeginV("Debug", &open, imgui.WindowFlagsNone)

	if imgui.CollapsingHeaderV("Lighting", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Lights", 2, imgui.TableFlagsBordersInnerV, imgui.Vec2{}, 0)
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
		imgui.BeginTableV("Bloom Table", 2, imgui.TableFlagsBordersInnerV, imgui.Vec2{}, 0)
		setupRow("Enable Bloom", func() { imgui.Checkbox("", &DBG.Bloom) })
		setupRow("Bloom Intensity", func() { imgui.SliderFloat("", &DBG.BloomIntensity, 0, 1) })
		setupRow("Bloom Threshold Passes", func() { imgui.SliderInt("", &DBG.BloomThresholdPasses, 0, 3) })
		setupRow("Bloom Threshold", func() { imgui.SliderFloat("", &DBG.BloomThreshold, 0, 3) })
		setupRow("Upsampling Scale", func() { imgui.SliderFloat("", &DBG.BloomUpsamplingScale, 0, 5.0) })
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("Other", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Bloom Table", 2, imgui.TableFlagsBordersInnerV, imgui.Vec2{}, 0)
		setupRow("Roughness", func() { imgui.SliderFloat("", &DBG.Roughness, 0, 1) })
		setupRow("Metallic", func() { imgui.SliderFloat("", &DBG.Metallic, 0, 1) })
		setupRow("Exposure", func() { imgui.SliderFloat("", &DBG.Exposure, 0, 1) })
		setupRow("Material Override", func() { imgui.Checkbox("", &DBG.MaterialOverride) })
		setupRow("Render Spatial Partition", func() { imgui.Checkbox("", &DBG.RenderSpatialPartition) })
		imgui.EndTable()
	}

	if imgui.CollapsingHeaderV("RenderStats", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("Bloom Table", 2, imgui.TableFlagsBordersInnerV, imgui.Vec2{}, 0)
		setupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%f", DBG.RenderTime)) })
		imgui.EndTable()
		var imageWidth float32 = 500
		if DBG.DebugTexture != 0 {
			texture := createUserSpaceTextureHandle(DBG.DebugTexture)
			size := imgui.Vec2{X: imageWidth, Y: imageWidth / float32(renderContext.AspectRatio())}
			// invert the Y axis since opengl vs texture coordinate systems differ
			// https://learnopengl.com/Getting-started/Textures
			imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
		}
	}

	imgui.End()
}

func setupRow(label string, item func()) {
	imgui.TableNextRow()
	imgui.TableNextColumn()
	imgui.Text(label)
	imgui.TableNextColumn()
	imgui.PushItemWidth(300)
	imgui.PushID(label)
	item()
	imgui.PopID()
	imgui.PopItemWidth()
}

// some detailed comment here
func createUserSpaceTextureHandle(texture uint32) imgui.TextureID {
	handle := 1<<63 | uint64(texture)
	return imgui.TextureID(handle)
}
