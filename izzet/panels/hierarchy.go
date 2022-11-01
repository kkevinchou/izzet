package panels

import (
	"sort"
	"strconv"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

type World interface {
	AddEntity(entity *entities.Entity)
	GetPrefabByID(id int) *prefabs.Prefab
}

var hierarchySelection int

func sceneHierarchy(es map[int]*entities.Entity, world World) *entities.Entity {
	regionSize := imgui.ContentRegionAvail()
	windowSize := imgui.Vec2{X: regionSize.X, Y: regionSize.Y * 0.5}
	imgui.BeginChildV("sceneHierarchy", windowSize, true, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: .95, Y: .91, Z: 0.81, W: 1})
	imgui.Text("Scene Hierarchy")
	imgui.PopStyleColor()

	keys := make([]int, 0, len(es))
	for k := range es {
		keys = append(keys, k)
	}
	sort.Ints(keys)

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
			if imgui.BeginDragDropTarget() {
				if payload := imgui.AcceptDragDropPayload("prefabid", imgui.DragDropFlagsNone); payload != nil {
					prefabID, err := strconv.Atoi(string(payload))
					if err != nil {
						panic(err)
					}

					prefab := world.GetPrefabByID(prefabID)
					entity := entities.InstantiateFromPrefab(prefab)
					world.AddEntity(entity)
				}
			}
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
