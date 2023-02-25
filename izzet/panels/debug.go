package panels

import "github.com/inkyblackness/imgui-go/v4"

type DebugSettings struct {
	DirectionalLightX int32
	DirectionalLightY int32
	DirectionalLightZ int32
	Roughness         float32
	Metallic          float32
}

var DBG DebugSettings = DebugSettings{
	DirectionalLightX: 0,
	DirectionalLightY: 0,
	DirectionalLightZ: 0,
	Roughness:         0.8,
	Metallic:          0.8,
}

func BuildDebug(world World) {
	if !ShowDebug {
		return
	}

	imgui.SetNextWindowPosV(imgui.Vec2{X: 400, Y: 400}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 100, Y: 100}, imgui.ConditionFirstUseEver)

	imgui.BeginV("Debug", &open, imgui.WindowFlagsNone)
	// imgui.Checkbox("multiply albedo", &DBG.MultiplyAlbedo)
	imgui.InputInt("directional light X", &DBG.DirectionalLightX)
	imgui.InputInt("directional light Y", &DBG.DirectionalLightY)
	imgui.InputInt("directional light Z", &DBG.DirectionalLightZ)
	imgui.SliderFloat("roughness", &DBG.Roughness, 0, 1)
	imgui.SliderFloat("metallic", &DBG.Metallic, 0, 1)
	imgui.End()

}
