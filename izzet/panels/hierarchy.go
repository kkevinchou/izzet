package panels

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
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

	entityPopup := false
	for _, entity := range es {
		if entity.Parent == nil {
			popup := drawEntity(entity, world)
			entityPopup = entityPopup || popup
		}
	}

	imgui.EndChild()

	if !entityPopup {
		imgui.PushID("sceneHierarchy")
		if imgui.BeginPopupContextItem() {
			if imgui.Button("Add Cube") {
				child := entities.CreateCube(25)
				world.AddEntity(child)
				SelectEntity(child)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Point Light") {
				lightInfo := &entities.LightInfo{
					Type:    1,
					Diffuse: mgl64.Vec4{1, 1, 1, 8000},
				}
				light := entities.CreateLight(lightInfo)
				world.AddEntity(light)
				SelectEntity(light)
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
		imgui.PopID()
	}
}

func drawEntity(entity *entities.Entity, world World) bool {
	popup := false
	nodeFlags := imgui.TreeNodeFlagsNone | imgui.TreeNodeFlagsLeaf
	if SelectedEntity() != nil && entity.ID == SelectedEntity().ID {
		nodeFlags |= imgui.TreeNodeFlagsSelected
	}

	if imgui.TreeNodeV(entity.NameID(), nodeFlags) {
		if imgui.IsItemClicked() && !imgui.IsItemToggledOpen() {
			SelectEntity(entity)
		}

		imgui.PushID(entity.NameID())
		if imgui.BeginPopupContextItem() {
			popup = true
			if imgui.Button("Add Cube") {
				child := entities.CreateCube(25)
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
			childPopup := drawEntity(child, world)
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
