package panels

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

func prefabsUI(ps []*prefabs.Prefab) {
	for _, prefab := range ps {
		nodeFlags := imgui.TreeNodeFlagsNone //| imgui.TreeNodeFlagsLeaf

		if imgui.TreeNodeV(prefab.Name, nodeFlags) {
			// this call allows drag/drop from an expanded tree node
			beginPrefabDragDrop(prefab.ID)
			if imgui.TreeNodeV("meshes", imgui.TreeNodeFlagsNone) {
				for _, mr := range prefab.ModelRefs {
					if imgui.TreeNodeV(mr.Name, imgui.TreeNodeFlagsLeaf) {
						imgui.TreePop()
					}
				}
				imgui.TreePop()
			}
			imgui.TreePop()
		}
		// this call allows drag/drop from a collapsed tree node
		beginPrefabDragDrop(prefab.ID)
	}
}

func beginPrefabDragDrop(id int) {
	if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
		str := fmt.Sprintf("%d", id)
		imgui.SetDragDropPayload("prefabid", []byte(str), imgui.ConditionNone)
		imgui.EndDragDropSource()
	}
}
