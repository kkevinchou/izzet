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

func sceneUI(app App) {
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 5, Y: 5})

	sceneHierarchy(app)

	if imgui.BeginDragDropTarget() {
		if payload := imgui.AcceptDragDropPayload("prefabid", imgui.DragDropFlagsNone); payload != nil {
			prefabID, err := strconv.Atoi(string(payload))

			if err != nil {
				panic(err)
			}

			prefab := app.GetPrefabByID(prefabID)
			entities := entities.InstantiateFromPrefab(prefab, app.ModelLibrary())
			for _, entity := range entities {
				app.AddEntity(entity)
			}

			if len(entities) > 0 {
				SelectEntity(entities[0])
			}
		}
		imgui.EndDragDropTarget()
	}
	imgui.PopStyleVar()
}

func sceneHierarchy(app App) {
	entityPopup := false
	imgui.BeginChildV("sceneHierarchy", imgui.Vec2{X: -1, Y: -1}, true, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)
	for _, entity := range app.Entities() {
		if entity.Parent == nil {
			popup := drawEntity(entity, app)
			entityPopup = entityPopup || popup
		}
	}
	imgui.EndChild()

	if !entityPopup {
		imgui.PushID("sceneHierarchy")
		if imgui.BeginPopupContextItem() {
			if imgui.Button("Add Player") {
				entity := entities.CreateCapsule(app.ModelLibrary(), 20, 10)
				entity.CharacterControllerComponent = &entities.CharacterControllerComponent{Speed: 10}
				app.AddEntity(entity)
				SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Capsule") {
				entity := entities.CreateCapsule(app.ModelLibrary(), 20, 10)
				app.AddEntity(entity)
				SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Cube") {
				entity := uiCreateCube(app, 100)
				app.AddEntity(entity)
				SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Triangle") {
				entity := entities.CreateTriangle(mgl64.Vec3{-10, -10, 0}, mgl64.Vec3{10, -10, 0}, mgl64.Vec3{0, 10, 0})
				app.AddEntity(entity)
				SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Point Light") {
				light := entities.CreatePointLight()
				app.AddEntity(light)
				SelectEntity(light)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Directional Light") {
				light := entities.CreateDirectionalLight()
				app.AddEntity(light)
				SelectEntity(light)
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
		imgui.PopID()
	}
}

func uiCreateCube(app App, length int) *entities.Entity {
	entity := entities.CreateCube(app.ModelLibrary(), length)

	meshHandle := entity.MeshComponent.MeshHandle
	primitives := app.ModelLibrary().GetPrimitives(meshHandle)
	entity.Collider = &entities.ColliderComponent{ColliderGroup: entities.ColliderGroupFlagTerrain, CollisionMask: entities.ColliderGroupFlagTerrain}
	entity.Collider.TriMeshCollider = collider.CreateTriMeshFromPrimitives(entities.MLPrimitivesTospecPrimitive(primitives))

	app.AddEntity(entity)
	return entity
}

func uiCreateCapsule(app App, length int, capsuleCollider bool) *entities.Entity {
	entity := entities.CreateCube(app.ModelLibrary(), length)

	if capsuleCollider {
		entity.Collider = &entities.ColliderComponent{
			CapsuleCollider: &collider.Capsule{
				Radius: 10,
				Top:    mgl64.Vec3{0, 20, 0},
				Bottom: mgl64.Vec3{0, -20, 0},
			},
			ColliderGroup: entities.ColliderGroupFlagPlayer,
			CollisionMask: entities.ColliderGroupFlagTerrain,
		}
	} else {
		meshHandle := entity.MeshComponent.MeshHandle
		primitives := app.ModelLibrary().GetPrimitives(meshHandle)
		entity.Collider = &entities.ColliderComponent{ColliderGroup: entities.ColliderGroupFlagTerrain, CollisionMask: entities.ColliderGroupFlagTerrain}
		entity.Collider.TriMeshCollider = collider.CreateTriMeshFromPrimitives(entities.MLPrimitivesTospecPrimitive(primitives))
	}

	app.AddEntity(entity)
	return entity
}

func drawEntity(entity *entities.Entity, app App) bool {
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
				child := uiCreateCube(app, 25)
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
				child := app.GetEntityByID(childID)
				parent := app.GetEntityByID(entity.ID)
				entities.BuildRelation(parent, child)
			}
			imgui.EndDragDropTarget()
		}

		childIDs := sortedIDs(entity.Children)
		for _, id := range childIDs {
			child := entity.Children[id]
			childPopup := drawEntity(child, app)
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
