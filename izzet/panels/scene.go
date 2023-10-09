package panels

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision/collider"
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
			entities := entities.InstantiateFromPrefab(prefab, world.ModelLibrary())
			for _, entity := range entities {
				world.AddEntity(entity)
			}

			if len(entities) > 0 {
				SelectEntity(entities[0])
			}
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
				entity := uiCreateCube(world, true)
				SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Triangle") {
				entity := entities.CreateTriangle(mgl64.Vec3{-10, -10, 0}, mgl64.Vec3{10, -10, 0}, mgl64.Vec3{0, 10, 0})
				world.AddEntity(entity)
				SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Point Light") {
				light := entities.CreatePointLight()
				world.AddEntity(light)
				SelectEntity(light)
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
		imgui.PopID()
	}
}

func uiCreateCube(world World, capsuleCollider bool) *entities.Entity {
	entity := entities.CreateCube(world.ModelLibrary(), 25)

	if capsuleCollider {
		entity.Collider = &entities.ColliderComponent{
			CapsuleCollider: &collider.Capsule{
				Radius: 5,
				Top:    mgl64.Vec3{0, 20, 0},
				Bottom: mgl64.Vec3{0, -20, 0},
			},
			CollisionMask: entities.ColliderGroupFlagTerrain,
		}
	}

	world.AddEntity(entity)
	return entity
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
				child := uiCreateCube(world, true)
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
