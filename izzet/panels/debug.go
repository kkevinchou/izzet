package panels

import "github.com/inkyblackness/imgui-go/v4"

type DebugSettings struct {
	MultiplyAlbedo bool
}

var DBG DebugSettings

func BuildDebug(world World) {
	if !ShowDebug {
		return
	}

	imgui.SetNextWindowPosV(imgui.Vec2{X: 400, Y: 400}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 100, Y: 100}, imgui.ConditionFirstUseEver)

	imgui.BeginV("Debug", &open, imgui.WindowFlagsNone)
	imgui.Checkbox("multiply albedo", &DBG.MultiplyAlbedo)
	imgui.End()

}
