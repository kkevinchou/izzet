package panels

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

func sceneHierarchy(es []*entities.Entity, world World) {
	regionSize := imgui.ContentRegionAvail()
	windowSize := imgui.Vec2{X: regionSize.X, Y: regionSize.Y * 0.5}
	imgui.BeginChildV("sceneHierarchy", windowSize, true, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: .95, Y: .91, Z: 0.81, W: 1})
	imgui.Text("Scene Hierarchy")
	imgui.PopStyleColor()

	for _, entity := range es {
		nodeFlags := imgui.TreeNodeFlagsNone | imgui.TreeNodeFlagsLeaf

		if SelectedEntity() != nil && entity.ID == SelectedEntity().ID {
			nodeFlags |= imgui.TreeNodeFlagsSelected
		}

		if imgui.TreeNodeV(entity.Name, nodeFlags) {
			// if imgui.BeginDragDropTarget() {
			// 	if payload := imgui.AcceptDragDropPayload("prefabid", imgui.DragDropFlagsNone); payload != nil {
			// 		prefabID, err := strconv.Atoi(string(payload))
			// 		if err != nil {
			// 			panic(err)
			// 		}

			// 		prefab := world.GetPrefabByID(prefabID)
			// 		entity := entities.InstantiateFromPrefab(prefab)
			// 		world.AddEntity(entity)
			// 	}
			// }
			imgui.TreePop()
		}

		if imgui.IsItemClicked() || imgui.IsItemToggledOpen() {
			SelectEntity(entity)
		}
	}

	imgui.EndChild()
}
