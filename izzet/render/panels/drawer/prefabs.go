package drawer

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var prefabsSelectIndex = -1

const prefabsContextItemID = "prefabsContextItem"

func prefabsUI(app renderiface.App, ps []*prefabs.Prefab) bool {
	var menuOpen bool
	if imgui.BeginTabItem("Prefabs") {
		for i, prefab := range ps {
			var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone //| imgui.TreeNodeFlagsLeaf

			open := imgui.TreeNodeExStrV(prefab.Name, nodeFlags)
			if prefabsSelectIndex == -1 || prefabsSelectIndex == i {
				menuOpen = prefabsBeginPopupContextItem(app, i, prefab)
			}

			if open {
				// this call allows drag/drop from an expanded tree node
				beginPrefabDragDrop(prefab.ID)
				// if imgui.TreeNodeExStrV("meshes", imgui.TreeNodeFlagsNone) {
				// 	for _, mr := range prefab.ModelRefs() {
				// 		if imgui.TreeNodeExStrV(mr.Name, imgui.TreeNodeFlagsLeaf) {
				// 			imgui.TreePop()
				// 		}
				// 	}
				// 	imgui.TreePop()
				// }
				imgui.TreePop()
			}
			// this call allows drag/drop from a collapsed tree node
			beginPrefabDragDrop(prefab.ID)
		}

		// if the pop up isn't open it means it was either never opened or dismissed, reset the select index
		if !imgui.IsPopupOpenStr(prefabsContextItemID) {
			prefabsSelectIndex = -1
		}
		imgui.EndTabItem()
	}
	return menuOpen
}

func beginPrefabDragDrop(id int) {
	// if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
	// 	str := fmt.Sprintf("%d", id)
	// 	imgui.SetDragDropPayload("prefabid", []byte(str), imgui.ConditionNone)
	// 	imgui.EndDragDropSource()
	// }
}

func prefabsBeginPopupContextItem(app renderiface.App, index int, prefab *prefabs.Prefab) bool {
	world := app.World()
	var menuOpen bool
	if imgui.BeginPopupContextItemV(prefabsContextItemID, imgui.PopupFlagsMouseButtonRight) {
		menuOpen = true
		if imgui.Button("Instantiate") {
			entities := entities.InstantiateFromPrefab(prefab, app.AssetManager())
			for _, entity := range entities {
				world.AddEntity(entity)
			}

			if len(entities) > 0 {
				app.SelectEntity(entities[0])
			}

			imgui.CloseCurrentPopup()
			prefabsSelectIndex = -1
		} else {
			prefabsSelectIndex = index
		}
		imgui.EndPopup()
	}
	return menuOpen
}
