package panels

import (
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
	BloomUpsamplingRadius     float32
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
	BloomThresholdPasses:      1,
	BloomThreshold:            3,
	BloomUpsamplingRadius:     0.005,
}

func BuildDebug(world World, renderContext RenderContext) {
	if !ShowDebug {
		return
	}

	// drawHUDTextureToQuad(cameraViewerContext, r.shaderManager.GetShaderProgram("depthDebug"), r.shadowMap.depthTexture, 1)

	imgui.SetNextWindowPosV(imgui.Vec2{X: 400, Y: 400}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 100, Y: 100}, imgui.ConditionFirstUseEver)

	imgui.BeginV("Debug", &open, imgui.WindowFlagsNone)

	if imgui.CollapsingHeaderV("Lighting Options", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.SliderFloat("ambient factor", &DBG.AmbientFactor, 0, 1)
		imgui.Dummy(imgui.Vec2{X: 0, Y: 10})
		imgui.SliderFloat("point light bias", &DBG.PointLightBias, 0, 1)
		imgui.InputInt("point light intensity", &DBG.PointLightIntensity)
		imgui.Dummy(imgui.Vec2{X: 0, Y: 10})
		imgui.InputInt("directional light intensity", &DBG.DirectionalLightIntensity)
		imgui.SliderFloat3("directional light dir", &DBG.DirectionalLightDir, -1, 1)

		imgui.Dummy(imgui.Vec2{X: 0, Y: 10})
		imgui.SliderFloat("bloom intensity", &DBG.BloomIntensity, 0, 5)
		imgui.SliderInt("bloom threshold passes", &DBG.BloomThresholdPasses, 0, 3)
		imgui.SliderFloat("bloom threshold", &DBG.BloomThreshold, 0, 3)
		imgui.SliderFloat("upsampling radius", &DBG.BloomUpsamplingRadius, 0, 1.0)

		imgui.Checkbox("bloom", &DBG.Bloom)
		imgui.Checkbox("enable shadow mapping", &DBG.EnableShadowMapping)
	}
	if imgui.CollapsingHeaderV("Other", imgui.TreeNodeFlagsNone) {
		imgui.SliderFloat("roughness", &DBG.Roughness, 0, 1)
		imgui.SliderFloat("metallic", &DBG.Metallic, 0, 1)
		imgui.SliderFloat("exposure", &DBG.Exposure, 0, 1)
		imgui.Checkbox("material override", &DBG.MaterialOverride)
	}

	var imageWidth float32 = 500
	if DBG.DebugTexture != 0 {
		texture := createUserSpaceTextureHandle(DBG.DebugTexture)
		size := imgui.Vec2{X: imageWidth, Y: imageWidth / float32(renderContext.AspectRatio())}
		// invert the Y axis since opengl vs texture coordinate systems differ
		// https://learnopengl.com/Getting-started/Textures
		imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
	}
	imgui.End()
}

// some detailed comment here
func createUserSpaceTextureHandle(texture uint32) imgui.TextureID {
	handle := 1<<63 | uint64(texture)
	return imgui.TextureID(handle)
}
