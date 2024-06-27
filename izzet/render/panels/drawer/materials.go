package drawer

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func materialssUI(app renderiface.App, ps []*prefabs.Prefab) bool {
	var menuOpen bool
	if imgui.BeginTabItem("Materials") {
		for i, _ := range app.MaterialBrowser().Items {
			var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone //| imgui.TreeNodeFlagsLeaf
			open := imgui.TreeNodeExStrV(fmt.Sprintf("material-%d", i), nodeFlags)
			if open {
				imgui.TreePop()
			}
		}
		imgui.EndTabItem()
	}
	return menuOpen
}
