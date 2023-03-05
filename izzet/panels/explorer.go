package panels

import (
	"strconv"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

var open bool = true

func BuildExplorer(es []*entities.Entity, world World, menuBarSize imgui.Vec2, renderContext RenderContext) {
	rect := imgui.Vec2{X: float32(renderContext.Width()), Y: float32(renderContext.Height()) - menuBarSize.Y}
	imgui.SetNextWindowBgAlpha(0.8)
	imgui.SetNextWindowPosV(imgui.Vec2{Y: menuBarSize.Y}, imgui.ConditionAlways, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: rect.X * 0.15, Y: rect.Y}, imgui.ConditionAlways)

	imgui.BeginV("explorer window", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize)
	// imgui.BeginChildV("explorer window", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize)
	imgui.BeginChildV("explorer", imgui.Vec2{}, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 5, Y: 5})
	sceneHierarchy(es, world)

	if imgui.BeginDragDropTarget() {
		if payload := imgui.AcceptDragDropPayload("prefabid", imgui.DragDropFlagsNone); payload != nil {
			prefabID, err := strconv.Atoi(string(payload))
			if err != nil {
				panic(err)
			}

			prefab := world.GetPrefabByID(prefabID)
			entity := entities.InstantiateFromPrefab(prefab)
			world.AddEntity(entity)
			SelectEntity(entity)
		}
		imgui.EndDragDropTarget()
	}

	entityProps(SelectedEntity())
	imgui.PopStyleVar()

	imgui.EndChild()
	imgui.End()
}
