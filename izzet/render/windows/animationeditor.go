package windows

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func renderAnimationEditorWindow(app renderiface.App) {
	if !app.RuntimeConfig().ShowAnimationEditor {
		return
	}

	center := imgui.MainViewport().Center()
	imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})

	if imgui.BeginV("Animation Editor", &app.RuntimeConfig().ShowAnimationEditor, imgui.WindowFlagsNone) {
	}
	imgui.End()
}
