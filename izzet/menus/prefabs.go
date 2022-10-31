package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func BuildPrefabs(es map[string]*entities.Entity) {
	var heightRatio float32 = 0.15
	imgui.SetNextWindowBgAlpha(0.8)
	imgui.SetNextWindowPosV(imgui.Vec2{X: float32(settings.Width) * 0.15, Y: float32(settings.Height) * (1 - heightRatio)}, imgui.ConditionAlways, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: float32(settings.Width)*(1-0.15) + 1, Y: float32(settings.Height) * heightRatio}, imgui.ConditionAlways)

	imgui.BeginV("prefab window", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize)
	imgui.BeginChildV("prefab", imgui.Vec2{}, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 5, Y: 5})
	prefabs()
	imgui.PopStyleVar()

	imgui.EndChild()
	imgui.End()
}

func prefabs() {
	regionSize := imgui.ContentRegionAvail()
	windowSize := imgui.Vec2{X: regionSize.X, Y: regionSize.Y}
	imgui.BeginChildV("prefab", windowSize, true, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)
	imgui.Text("hello")
	imgui.EndChild()
}
