package drawer

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/windows"
)

var materialPopupMenu bool

func materialssUI(app renderiface.App) {
	for _, material := range app.AssetManager().GetMaterials() {
		var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone //| imgui.TreeNodeFlagsLeaf
		open := imgui.TreeNodeExStrV(material.Name, nodeFlags)

		id := material.Handle.String()
		if imgui.BeginPopupContextItemV(id, imgui.PopupFlagsMouseButtonRight) {
			if imgui.Button("Edit") {
				material := app.AssetManager().GetMaterial(material.Handle)
				windows.ShowEditMaterialWindow(app, material)
				imgui.CloseCurrentPopup()
			} else {
			}
			imgui.EndPopup()
		}

		if open {
			imgui.TreePop()
		}
	}
}
