package drawer

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func materialssUI(app renderiface.App) {
	for _, material := range app.MaterialBrowser().Items {
		var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone //| imgui.TreeNodeFlagsLeaf
		open := imgui.TreeNodeExStrV(material.ID, nodeFlags)
		if open {
			imgui.TreePop()
		}
	}
}
