package menus

import (
	"sort"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

var hierarchySelection int

func sceneHierarchy(es map[string]*entities.Entity) *entities.Entity {
	regionSize := imgui.ContentRegionAvail()
	windowSize := imgui.Vec2{X: regionSize.X, Y: regionSize.Y * 0.5}
	imgui.BeginChildV("sceneHierarchy", windowSize, true, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: .95, Y: .91, Z: 0.81, W: 1})
	imgui.Text("Scene Hierarchy")
	imgui.PopStyleColor()

	keys := make([]string, 0, len(es))
	for k := range es {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var selectedEntity *entities.Entity
	selectedItem := -1
	for i, k := range keys {
		entity := es[k]

		nodeFlags := imgui.TreeNodeFlagsNone | imgui.TreeNodeFlagsLeaf
		if hierarchySelection&(1<<i) != 0 {
			selectedEntity = entity
			nodeFlags |= imgui.TreeNodeFlagsSelected
		}

		if imgui.TreeNodeV(entity.Name, nodeFlags) {
			imgui.TreePop()
		}

		if imgui.IsItemClicked() || imgui.IsItemToggledOpen() {
			selectedItem = i
		}

	}
	if selectedItem != -1 {
		hierarchySelection = (1 << selectedItem)
	}

	imgui.EndChild()
	return selectedEntity
}
