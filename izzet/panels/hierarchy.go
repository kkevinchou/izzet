package panels

import (
	"fmt"
	"strconv"

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
			if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
				str := fmt.Sprintf("%d", entity.ID)
				imgui.SetDragDropPayload("childid", []byte(str), imgui.ConditionNone)
				imgui.EndDragDropSource()
			}
			if imgui.BeginDragDropTarget() {
				if payload := imgui.AcceptDragDropPayload("childid", imgui.DragDropFlagsNone); payload != nil {
					childID, err := strconv.Atoi(string(payload))
					if err != nil {
						panic(err)
					}
					child := world.GetEntityByID(childID)
					parent := world.GetEntityByID(entity.ID)
					world.BuildRelation(parent, child)
				}
				imgui.EndDragDropTarget()
			}
			if imgui.TreeNodeV("asdf", imgui.TreeNodeFlagsLeaf) {
				imgui.TreePop()
			}
			imgui.TreePop()
		}

		if imgui.IsItemClicked() || imgui.IsItemToggledOpen() {
			SelectEntity(entity)
		}
	}

	imgui.EndChild()
}
