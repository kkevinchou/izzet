package drawer

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var prefabsSelectIndex = -1

const prefabsContextItemID = "prefabsContextItem"

func prefabsUI(app renderiface.App, world renderiface.GameWorld, ps []*prefabs.Prefab) bool {
	var menuOpen bool
	if imgui.BeginTabItem("Prefabs") {
		for i, prefab := range ps {
			nodeFlags := imgui.TreeNodeFlagsNone //| imgui.TreeNodeFlagsLeaf

			open := imgui.TreeNodeV(prefab.Name, nodeFlags)

			if prefabsSelectIndex == -1 || prefabsSelectIndex == i {
				menuOpen = prefabsBeginPopupContextItem(app, world, i, prefab)
			}

			if open {
				// this call allows drag/drop from an expanded tree node
				beginPrefabDragDrop(prefab.ID)
				// if imgui.TreeNodeV("meshes", imgui.TreeNodeFlagsNone) {
				// 	for _, mr := range prefab.ModelRefs() {
				// 		if imgui.TreeNodeV(mr.Name, imgui.TreeNodeFlagsLeaf) {
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
		if !imgui.IsPopupOpen(prefabsContextItemID) {
			prefabsSelectIndex = -1
		}
		imgui.EndTabItem()
	}
	return menuOpen
}

func beginPrefabDragDrop(id int) {
	if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
		str := fmt.Sprintf("%d", id)
		imgui.SetDragDropPayload("prefabid", []byte(str), imgui.ConditionNone)
		imgui.EndDragDropSource()
	}
}

func prefabsBeginPopupContextItem(app renderiface.App, world renderiface.GameWorld, index int, prefab *prefabs.Prefab) bool {
	var menuOpen bool
	if imgui.BeginPopupContextItemV(prefabsContextItemID, imgui.PopupFlagsMouseButtonRight) {
		menuOpen = true
		if imgui.Button("Instantiate") {
			entities := entities.InstantiateFromPrefab(prefab, app.ModelLibrary())
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