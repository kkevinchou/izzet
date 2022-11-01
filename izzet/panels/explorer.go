package panels

import (
	"strconv"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
)

var open bool = true

func BuildExplorer(es map[int]*entities.Entity, world World) {
	imgui.SetNextWindowBgAlpha(0.8)
	imgui.SetNextWindowPosV(imgui.Vec2{}, imgui.ConditionAlways, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: float32(settings.Width) * 0.15, Y: float32(settings.Height)}, imgui.ConditionAlways)

	imgui.BeginV("explorer window", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize)
	imgui.BeginChildV("explorer", imgui.Vec2{}, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 5, Y: 5})
	selectedEntity := sceneHierarchy(es, world)
	if imgui.BeginDragDropTarget() {
		if payload := imgui.AcceptDragDropPayload("prefabid", imgui.DragDropFlagsNone); payload != nil {
			prefabID, err := strconv.Atoi(string(payload))
			if err != nil {
				panic(err)
			}

			prefab := world.GetPrefabByID(prefabID)
			entity := entities.InstantiateFromPrefab(prefab)
			world.AddEntity(entity)
		}
	}
	entityProps(selectedEntity)
	imgui.PopStyleVar()

	imgui.EndChild()
	imgui.End()
}
