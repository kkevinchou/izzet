package panels

import (
	"github.com/inkyblackness/imgui-go/v4"
)

type DebugSettings struct {
	DirectionalLightX         int32
	DirectionalLightY         int32
	DirectionalLightZ         int32
	Roughness                 float32
	Metallic                  float32
	PointLightIntensity       int32
	DirectionalLightIntensity int32
	PointLightBias            float32
	MaterialOverride          bool
	EnableShadowMapping       bool
	DebugTexture              uint32 // 64 bits as we need extra bits to specify a the type of texture to IMGUI
}

var DBG DebugSettings = DebugSettings{
	DirectionalLightX:         0,
	DirectionalLightY:         -1,
	DirectionalLightZ:         -1,
	Roughness:                 0.55,
	Metallic:                  1.0,
	PointLightIntensity:       100000,
	DirectionalLightIntensity: 10,
	PointLightBias:            1,
	MaterialOverride:          false,
	EnableShadowMapping:       false,
}

func BuildDebug(world World, renderContext RenderContext) {
	if !ShowDebug {
		return
	}

	// drawHUDTextureToQuad(cameraViewerContext, r.shaderManager.GetShaderProgram("depthDebug"), r.shadowMap.depthTexture, 1)

	imgui.SetNextWindowPosV(imgui.Vec2{X: 400, Y: 400}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 100, Y: 100}, imgui.ConditionFirstUseEver)

	imgui.BeginV("Debug", &open, imgui.WindowFlagsNone)
	// imgui.Checkbox("multiply albedo", &DBG.MultiplyAlbedo)
	imgui.InputInt("directional light X", &DBG.DirectionalLightX)
	imgui.InputInt("directional light Y", &DBG.DirectionalLightY)
	imgui.InputInt("directional light Z", &DBG.DirectionalLightZ)
	imgui.InputInt("point light intensity", &DBG.PointLightIntensity)
	imgui.InputInt("directional light intensity", &DBG.DirectionalLightIntensity)
	imgui.SliderFloat("point light bias", &DBG.PointLightBias, 0, 1)
	imgui.SliderFloat("roughness", &DBG.Roughness, 0, 1)
	imgui.SliderFloat("metallic", &DBG.Metallic, 0, 1)
	imgui.Checkbox("material override", &DBG.MaterialOverride)
	imgui.Checkbox("enable shadow mapping", &DBG.EnableShadowMapping)

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
