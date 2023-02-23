package panels

import "github.com/inkyblackness/imgui-go/v4"

type DebugSettings struct {
	MultiplyAlbedo bool

	DirectionalLightX int32
	DirectionalLightY int32
	DirectionalLightZ int32
}

var DBG DebugSettings = DebugSettings{
	DirectionalLightX: -1,
	DirectionalLightY: -1,
	DirectionalLightZ: -1,
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
	imgui.End()

}
