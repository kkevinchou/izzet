package panels

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

func sceneUI(es []*entities.Entity, world World) {
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 5, Y: 5})

	sceneHierarchy(es, world)

	if imgui.BeginDragDropTarget() {
		fmt.Println("ACCEPT PAYLOAD2")
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
	} else {
		// fmt.Println("WAT")
	}
	imgui.PopStyleVar()
}

func sceneHierarchy(es []*entities.Entity, world World) {
	entityPopup := false
	for _, entity := range es {
		if entity.Parent == nil {
			popup := drawEntity(entity, world)
			entityPopup = entityPopup || popup
		}
	}

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
			fmt.Println("ACCEPT PAYLOAD1")
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
