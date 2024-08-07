package panels

import (
	"math"
	"sort"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/material"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/collision/collider"
)

func sceneGraph(app renderiface.App) {
	world := app.World()
	entityPopup := false
	imgui.BeginChildStrV("sceneGraphNodes", imgui.Vec2{X: -1, Y: -1}, imgui.ChildFlagsBorder, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)
	for _, entity := range world.Entities() {
		if entity.Parent == nil {
			popup := drawSceneGraphEntity(entity, app)
			entityPopup = entityPopup || popup
		}
	}
	imgui.EndChild()

	if !entityPopup {
		if imgui.BeginPopupContextItemV("NULL", imgui.PopupFlagsMouseButtonRight) {
			if imgui.Button("Add Player") {
				var radius float64 = 40
				var length float64 = 80
				entity := entities.InstantiateEntity("player")
				entity.Physics = &entities.PhysicsComponent{GravityEnabled: true}
				entity.Collider = &entities.ColliderComponent{
					CapsuleCollider: &collider.Capsule{
						Radius: radius,
						Top:    mgl64.Vec3{0, radius + length, 0},
						Bottom: mgl64.Vec3{0, radius, 0},
					},
					ColliderGroup: entities.ColliderGroupFlagPlayer,
					CollisionMask: entities.ColliderGroupFlagTerrain,
				}
				entity.CharacterControllerComponent = &entities.CharacterControllerComponent{Speed: settings.CharacterSpeed}

				capsule := entity.Collider.CapsuleCollider
				entity.InternalBoundingBox = collider.BoundingBox{MinVertex: capsule.Bottom.Sub(mgl64.Vec3{radius, radius, radius}), MaxVertex: capsule.Top.Add(mgl64.Vec3{radius, radius, radius})}

				handle := assets.NewGlobalHandle("alpha3")
				entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true}
				entity.Animation = entities.NewAnimationComponent("alpha3", app.AssetManager())
				entities.SetScale(entity, mgl64.Vec3{0.25, 0.25, 0.25})

				world.AddEntity(entity)
				app.SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Cube") {
				entity := entities.CreateCube(app.AssetManager(), 1)
				i := 0
				entity.Material = &entities.MaterialComponent{
					Material: material.Material{
						PBR: types.PBR{
							Roughness:        0.85,
							Metallic:         0,
							Diffuse:          [3]float32{1, 1, 1},
							DiffuseIntensity: 1,

							ColorTextureIndex: &i,
							TextureName:       settings.DefaultTexture,
						},
					},
				}

				meshHandle := entity.MeshComponent.MeshHandle
				primitives := app.AssetManager().GetPrimitives(meshHandle)
				entity.Collider = &entities.ColliderComponent{ColliderGroup: entities.ColliderGroupFlagTerrain, CollisionMask: entities.ColliderGroupFlagTerrain}
				entity.Collider.TriMeshCollider = collider.CreateTriMeshFromPrimitives(entities.MLPrimitivesTospecPrimitive(primitives))

				world.AddEntity(entity)
				app.SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Point Light") {
				light := entities.CreatePointLight()
				world.AddEntity(light)
				app.SelectEntity(light)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Directional Light") {
				light := entities.CreateDirectionalLight()
				world.AddEntity(light)
				app.SelectEntity(light)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Empty Entity") {
				entity := entities.InstantiateEntity("empty-entity")
				world.AddEntity(entity)
				app.SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Camera") {
				entity := entities.InstantiateEntity("camera")
				entity.CameraComponent = &entities.CameraComponent{}
				entity.ImageInfo = entities.NewImageInfo("camera.png", 15)
				entity.Billboard = true
				world.AddEntity(entity)
				app.SelectEntity(entity)
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
	}
}

func drawSceneGraphEntity(entity *entities.Entity, app renderiface.App) bool {
	popup := false
	var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone
	if len(entity.Children) == 0 {
		nodeFlags |= imgui.TreeNodeFlagsLeaf
	}
	if app.SelectedEntity() != nil && entity.ID == app.SelectedEntity().ID {
		nodeFlags |= imgui.TreeNodeFlagsSelected
	}

	if imgui.TreeNodeExStrV(entity.NameID(), nodeFlags) {
		if imgui.IsItemClicked() || imgui.IsItemToggledOpen() {
			app.SelectEntity(entity)
		}

		imgui.PushIDStr(entity.NameID())
		if imgui.BeginPopupContextItemV("NULL", imgui.PopupFlagsMouseButtonRight) {
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

		// if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
		// 	fmt.Println("BEGIN DRAG DROP")
		// 	str := fmt.Sprintf("%d", entity.ID)
		// 	imgui.SetDragDropPayload("childid", []byte(str), imgui.ConditionNone)
		// 	imgui.EndDragDropSource()
		// }
		// if imgui.BeginDragDropTarget() {
		// 	fmt.Println("END DRAG DROP")
		// 	if payload := imgui.AcceptDragDropPayload("childid", imgui.DragDropFlagsNone); payload != nil {
		// 		childID, err := strconv.Atoi(string(payload))
		// 		if err != nil {
		// 			panic(err)
		// 		}
		// 		child := world.GetEntityByID(childID)
		// 		parent := world.GetEntityByID(entity.ID)
		// 		entities.BuildRelation(parent, child)
		// 	}
		// 	imgui.EndDragDropTarget()
		// }

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

func drawEntity(entity *entities.Entity, app renderiface.App) bool {
	popup := false
	var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone
	if len(entity.Children) == 0 {
		nodeFlags |= imgui.TreeNodeFlagsLeaf
	}
	if app.SelectedEntity() != nil && entity.ID == app.SelectedEntity().ID {
		nodeFlags |= imgui.TreeNodeFlagsSelected
	}

	if imgui.TreeNodeExStrV(entity.NameID(), nodeFlags) {
		if imgui.IsItemClicked() || imgui.IsItemToggledOpen() {
			app.SelectEntity(entity)
		}

		imgui.PushIDStr(entity.NameID())
		if imgui.BeginPopupContextItemV("NULL", imgui.PopupFlagsMouseButtonRight) {
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

		// if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
		// 	str := fmt.Sprintf("%d", entity.ID)
		// 	imgui.SetDragDropPayload("childid", []byte(str), imgui.ConditionNone)
		// 	imgui.EndDragDropSource()
		// }
		// if imgui.BeginDragDropTarget() {
		// 	if payload := imgui.AcceptDragDropPayload("childid", imgui.DragDropFlagsNone); payload != nil {
		// 		childID, err := strconv.Atoi(string(payload))
		// 		if err != nil {
		// 			panic(err)
		// 		}
		// 		child := world.GetEntityByID(childID)
		// 		parent := world.GetEntityByID(entity.ID)
		// 		entities.BuildRelation(parent, child)
		// 	}
		// 	imgui.EndDragDropTarget()
		// }

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
