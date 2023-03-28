package panels

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

func sceneUI(world World) {
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 5, Y: 5})

	sceneHierarchy(world)

	if imgui.BeginDragDropTarget() {
		if payload := imgui.AcceptDragDropPayload("prefabid", imgui.DragDropFlagsNone); payload != nil {
			prefabID, err := strconv.Atoi(string(payload))

			if err != nil {
				panic(err)
			}

			prefab := world.GetPrefabByID(prefabID)
			parent := entities.CreateDummy(prefab.Name)
			world.AddEntity(parent)
			for _, entity := range entities.InstantiateFromPrefab(prefab) {
				world.AddEntity(entity)
				entities.BuildRelation(parent, entity)
			}
			SelectEntity(parent)
		}
		imgui.EndDragDropTarget()
	}
	imgui.PopStyleVar()
}

func sceneHierarchy(world World) {
	entityPopup := false
	imgui.BeginChildV("sceneHierarchy", imgui.Vec2{X: -1, Y: -1}, true, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)
	for _, entity := range world.Entities() {
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
			if imgui.Button("Add Triangle") {
				child := entities.CreateTriangle(mgl64.Vec3{-10, -10, 0}, mgl64.Vec3{10, -10, 0}, mgl64.Vec3{0, 10, 0})
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
	nodeFlags := imgui.TreeNodeFlagsNone
	if len(entity.Children) == 0 {
		nodeFlags |= imgui.TreeNodeFlagsLeaf
	}
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
				entities.BuildRelation(entity, child)
				SelectEntity(child)
				imgui.CloseCurrentPopup()
			}
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
				entities.BuildRelation(parent, child)
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
