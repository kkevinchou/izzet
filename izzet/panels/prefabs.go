package panels

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

func BuildPrefabs(ps []*prefabs.Prefab, world World) {
	var heightRatio float32 = 0.15
	_ = heightRatio
	imgui.SetNextWindowBgAlpha(0.8)

	x, y := world.Window().GetSize()
	rect := imgui.Vec2{X: float32(x), Y: float32(y)}
	imgui.SetNextWindowPosV(imgui.Vec2{X: float32(rect.X) * 0.15, Y: float32(rect.Y) * (1 - heightRatio)}, imgui.ConditionAlways, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: float32(rect.X)*(1-0.15) + 1, Y: float32(rect.Y) * (heightRatio)}, imgui.ConditionAlways)

	imgui.BeginV("prefab window", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize)
	imgui.BeginChildV("prefab", imgui.Vec2{}, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 5, Y: 5})
	prefabsUI(ps)
	imgui.PopStyleVar()

	imgui.EndChild()
	imgui.End()
}

func prefabsUI(ps []*prefabs.Prefab) {
	regionSize := imgui.ContentRegionAvail()
	windowSize := imgui.Vec2{X: regionSize.X, Y: regionSize.Y}
	imgui.BeginChildV("prefab", windowSize, true, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: .95, Y: .91, Z: 0.81, W: 1})
	imgui.Text("Prefabs")
	imgui.PopStyleColor()

	for _, prefab := range ps {
		nodeFlags := imgui.TreeNodeFlagsNone //| imgui.TreeNodeFlagsLeaf

		if imgui.TreeNodeV(prefab.Name, nodeFlags) {
			// this call allows drag/drop from an expanded tree node
			beginPrefabDragDrop(prefab.ID)
			if imgui.TreeNodeV("meshes", imgui.TreeNodeFlagsNone) {
				for _, mr := range prefab.ModelRefs {
					if imgui.TreeNodeV(mr.Name, imgui.TreeNodeFlagsLeaf) {
						imgui.TreePop()
					}
				}
				imgui.TreePop()
			}
			imgui.TreePop()
		}
		// this call allows drag/drop from a collapsed tree node
		beginPrefabDragDrop(prefab.ID)
	}

	imgui.EndChild()
}

func beginPrefabDragDrop(id int) {
	if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
		str := fmt.Sprintf("%d", id)
		imgui.SetDragDropPayload("prefabid", []byte(str), imgui.ConditionNone)
		imgui.EndDragDropSource()
	}
}
