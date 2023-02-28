package panels

import (
	"fmt"
	"sort"
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
		if entity.Parent == nil {
			drawEntity(entity, world)
		}
	}

	imgui.EndChild()
}

func drawEntity(entity *entities.Entity, world World) {
	nodeFlags := imgui.TreeNodeFlagsNone | imgui.TreeNodeFlagsLeaf
	if SelectedEntity() != nil && entity.ID == SelectedEntity().ID {
		nodeFlags |= imgui.TreeNodeFlagsSelected
	}

	// var childClicked bool
	if imgui.TreeNodeV(entity.NameID(), nodeFlags) {
		if imgui.IsItemClicked() && !imgui.IsItemToggledOpen() {
			SelectEntity(entity)
		}

		imgui.PushID(entity.NameID())
		imgui.PushStyleColor(imgui.StyleColorButton, imgui.Vec4{X: 66. / 255, Y: 17. / 255, Z: 212. / 255, W: 1})
		imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
		if imgui.BeginPopupContextItem() {
			if imgui.Button("Add Cube") {
				child := entities.CreateCube()
				world.AddEntity(child)
				world.BuildRelation(entity, child)
				SelectEntity(child)
				imgui.CloseCurrentPopup()
			}
			if entity.Parent != nil {
				if imgui.Button("Remove Parent") {
					world.RemoveParent(entity)
					imgui.CloseCurrentPopup()
				}
			}
			imgui.EndPopup()
		}
		imgui.PopStyleColor()
		imgui.PopStyleColor()
		imgui.PopID()

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

		childIDs := sortedIDs(entity.Children)
		for _, id := range childIDs {
			child := entity.Children[id]
			drawEntity(child, world)
		}

		imgui.TreePop()
	}
}

func sortedIDs(m map[int]*entities.Entity) []int {
	var ids []int
	for id, _ := range m {
		ids = append(ids, id)
	}

	sort.Ints(ids)
	return ids
}
