package panels

import "github.com/inkyblackness/imgui-go/v4"

type DebugSettings struct {
	DirectionalLightX         int32
	DirectionalLightY         int32
	DirectionalLightZ         int32
	Roughness                 float32
	Metallic                  float32
	PointLightIntensity       int32
	DirectionalLightIntensity int32
	PointLightBias            float32
}

var DBG DebugSettings = DebugSettings{
	DirectionalLightX:         0,
	DirectionalLightY:         -1,
	DirectionalLightZ:         -1,
	Roughness:                 0.55,
	Metallic:                  1.0,
	PointLightIntensity:       100000,
	DirectionalLightIntensity: 20,
	PointLightBias:            1,
}

func BuildDebug(world World, depthTexture uint32, aspectRatio float64) {
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
	imgui.SliderFloat("point light bias", &DBG.PointLightBias, 0, 10)
	imgui.SliderFloat("roughness", &DBG.Roughness, 0, 1)
	imgui.SliderFloat("metallic", &DBG.Metallic, 0, 1)

	var imageWidth float32 = 500
	imgui.Image(imgui.TextureID(depthTexture), imgui.Vec2{X: imageWidth, Y: imageWidth / float32(aspectRatio)})
	imgui.End()

}
