package panels

import (
	"sort"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func SceneGraph(app renderiface.App) {
	world := app.World()
	entityPopup := false
	imgui.BeginChildStrV("sceneGraphNodes", imgui.Vec2{X: -1, Y: -1}, imgui.ChildFlagsBorders, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)
	for _, entity := range world.Entities() {
		if entity.Parent == nil {
			popup := drawSceneGraphEntity(entity, app)
			entityPopup = entityPopup || popup
		}
	}
	imgui.EndChild()

	if !entityPopup {
		if imgui.BeginPopupContextItemV("NULL", imgui.PopupFlagsMouseButtonRight) {
			if imgui.Button("Add Cube") {
				e := entity.CreateCube(app.AssetManager(), 1)
				e.Material = &entity.MaterialComponent{
					MaterialHandle: assets.DefaultMaterialHandle,
				}
				e.Static = true

				world.AddEntity(e)
				app.SelectEntity(e)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Point Light") {
				light := entity.CreatePointLight()
				world.AddEntity(light)
				app.SelectEntity(light)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Directional Light") {
				light := entity.CreateDirectionalLight()
				world.AddEntity(light)
				app.SelectEntity(light)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Empty Entity") {
				e := entity.CreateEmptyEntity("empty-entity")
				world.AddEntity(e)
				app.SelectEntity(e)
				imgui.CloseCurrentPopup()
			}
			if imgui.Button("Add Camera") {
				e := entity.CreateEmptyEntity("camera")
				e.CameraComponent = &entity.CameraComponent{}
				e.ImageInfo = entity.NewImageInfo("camera.png", 15)
				e.Billboard = true
				world.AddEntity(e)
				app.SelectEntity(e)
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
	}
}

func drawSceneGraphEntity(e *entity.Entity, app renderiface.App) bool {
	popup := false
	var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone
	if len(e.Children) == 0 {
		nodeFlags |= imgui.TreeNodeFlagsLeaf
	}
	if app.SelectedEntity() != nil && e.ID == app.SelectedEntity().ID {
		nodeFlags |= imgui.TreeNodeFlagsSelected
	}

	if imgui.TreeNodeExStrV(e.NameID(), nodeFlags) {
		if imgui.IsItemClicked() || imgui.IsItemToggledOpen() {
			app.SelectEntity(e)
		}

		imgui.PushIDStr(e.NameID())
		if imgui.BeginPopupContextItemV("NULL", imgui.PopupFlagsMouseButtonRight) {
			popup = true
			if e.Parent != nil {
				if imgui.Button("Remove Parent") {
					entity.RemoveParent(e)
					imgui.CloseCurrentPopup()
				}
			}
			imgui.EndPopup()
		}
		imgui.PopID()

		// if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
		// 	fmt.Println("BEGIN DRAG DROP")
		// 	str := fmt.Sprintf("%d", e.ID)
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
		// 		parent := world.GetEntityByID(e.ID)
		// 		entity.BuildRelation(parent, child)
		// 	}
		// 	imgui.EndDragDropTarget()
		// }

		childIDs := sortedIDs(e.Children)
		for _, id := range childIDs {
			child := e.Children[id]
			childPopup := drawEntity(child, app)
			popup = popup || childPopup
		}

		imgui.TreePop()
	}
	return popup
}

func sortedIDs(m map[int]*entity.Entity) []int {
	var ids []int
	for id, _ := range m {
		ids = append(ids, id)
	}

	sort.Ints(ids)
	return ids
}

func drawEntity(e *entity.Entity, app renderiface.App) bool {
	popup := false
	var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone
	if len(e.Children) == 0 {
		nodeFlags |= imgui.TreeNodeFlagsLeaf
	}
	if app.SelectedEntity() != nil && e.ID == app.SelectedEntity().ID {
		nodeFlags |= imgui.TreeNodeFlagsSelected
	}

	if imgui.TreeNodeExStrV(e.NameID(), nodeFlags) {
		if imgui.IsItemClicked() || imgui.IsItemToggledOpen() {
			app.SelectEntity(e)
		}

		imgui.PushIDStr(e.NameID())
		if imgui.BeginPopupContextItemV("NULL", imgui.PopupFlagsMouseButtonRight) {
			popup = true
			if e.Parent != nil {
				if imgui.Button("Remove Parent") {
					entity.RemoveParent(e)
					imgui.CloseCurrentPopup()
				}
			}
			imgui.EndPopup()
		}
		imgui.PopID()

		// if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
		// 	str := fmt.Sprintf("%d", e.ID)
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
		// 		parent := world.GetEntityByID(e.ID)
		// 		entity.BuildRelation(parent, child)
		// 	}
		// 	imgui.EndDragDropTarget()
		// }

		childIDs := sortedIDs(e.Children)
		for _, id := range childIDs {
			child := e.Children[id]
			childPopup := drawEntity(child, app)
			popup = popup || childPopup
		}

		imgui.TreePop()
	}
	return popup
}
