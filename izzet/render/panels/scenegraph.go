package panels

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func sceneGraph(app renderiface.App, world GameWorld) {
	entityPopup := false
	for _, entity := range world.Entities() {
		if entity.Parent == nil {
			popup := drawSceneGraphEntity(entity, app, world)
			entityPopup = entityPopup || popup
		}
	}
}

func drawSceneGraphEntity(entity *entities.Entity, app renderiface.App, world GameWorld) bool {
	popup := false
	nodeFlags := imgui.TreeNodeFlagsNone
	if len(entity.Children) == 0 {
		nodeFlags |= imgui.TreeNodeFlagsLeaf
	}
	if SelectedEntity() != nil && entity.ID == SelectedEntity().ID {
		nodeFlags |= imgui.TreeNodeFlagsSelected
	}

	if imgui.TreeNodeV(entity.NameID(), nodeFlags) {
		if imgui.IsItemClicked() || imgui.IsItemToggledOpen() {
			SelectEntity(entity)
		}

		imgui.PushID(entity.NameID())
		if imgui.BeginPopupContextItem() {
			popup = true
			if entity.Parent != nil {
				if imgui.Button("Remove Parent") {
					entities.RemoveParent(entity)
					imgui.CloseCurrentPopup()
				}
			}
			imgui.EndPopup()
		}
		imgui.PopID()

		if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
			fmt.Println("BEGIN DRAG DROP")
			str := fmt.Sprintf("%d", entity.ID)
			imgui.SetDragDropPayload("childid", []byte(str), imgui.ConditionNone)
			imgui.EndDragDropSource()
		}
		if imgui.BeginDragDropTarget() {
			fmt.Println("END DRAG DROP")
			if payload := imgui.AcceptDragDropPayload("childid", imgui.DragDropFlagsNone); payload != nil {
				childID, err := strconv.Atoi(string(payload))
				if err != nil {
					panic(err)
				}
				child := world.GetEntityByID(childID)
				parent := world.GetEntityByID(entity.ID)
				entities.BuildRelation(parent, child)
			}
			imgui.EndDragDropTarget()
		}

		childIDs := sortedIDs(entity.Children)
		for _, id := range childIDs {
			child := entity.Children[id]
			childPopup := drawEntity(child, app, world)
			popup = popup || childPopup
		}

		imgui.TreePop()
	}
	return popup
}

func sortedIDs(m map[int]*entities.Entity) []int {
	var ids []int
	for id, _ := range m {
		ids = append(ids, id)
	}

	sort.Ints(ids)
	return ids
}
