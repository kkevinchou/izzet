package drawer

import (
	"fmt"
	"sort"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/prefab"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

const deletePrefabConfirmationPopup = "Delete Prefab"

var pendingDeletePrefab *prefab.PrefabHandle
var showDeletePrefabConfirmationPopup bool

func prefabsUI(app renderiface.App) {
	style := imgui.CurrentStyle()
	imgui.PushStyleVarVec2(
		imgui.StyleVarCellPadding,
		imgui.Vec2{X: style.CellPadding().X, Y: 5},
	)
	defer imgui.PopStyleVar()

	handles := sortedPrefabHandles()
	if imgui.BeginTableV("PrefabsTable", itemsPerRow,
		imgui.TableFlagsSizingFixedSame, imgui.Vec2{X: 0, Y: 0}, 0) {

		for _, handle := range handles {
			imgui.TableNextColumn()
			drawPrefabCell(app, handle)
		}
		imgui.EndTable()
	}

	renderDeletePrefabConfirmationPopup()
}

func sortedPrefabHandles() []prefab.PrefabHandle {
	handles := make([]prefab.PrefabHandle, 0, len(prefab.PrefabRegistry))
	for handle := range prefab.PrefabRegistry {
		handles = append(handles, handle)
	}

	sort.Slice(handles, func(i, j int) bool {
		return string(handles[i]) < string(handles[j])
	})
	return handles
}

func drawPrefabCell(app renderiface.App, handle prefab.PrefabHandle) {
	name := string(handle)

	imgui.PushIDStr(name)
	defer imgui.PopID()

	t := app.AssetManager().GetTexture("document")

	imgui.ImageV(
		imgui.TextureID(t.ID),
		imgui.Vec2{X: cellWidth, Y: cellHeight},
		imgui.Vec2{X: 0, Y: 1},
		imgui.Vec2{X: 1, Y: 0},
		imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1},
		imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0},
	)

	if imgui.BeginPopupContextItemV("prefab-context-"+name, imgui.PopupFlagsMouseButtonRight) {
		if imgui.Button("Instantiate") {
			e := prefab.Instantiate(handle, app.AssetManager())
			app.World().AddEntity(e)
			app.SelectEntity(e)
			imgui.CloseCurrentPopup()
		}
		if imgui.Button("Delete") {
			pendingDeletePrefab = &handle
			showDeletePrefabConfirmationPopup = true
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}

	if imgui.IsItemHovered() {
		imgui.BeginTooltip()
		imgui.Text(name)
		imgui.EndTooltip()
	}

	label := ellipsize(name, cellWidth)
	textSize := imgui.CalcTextSizeV(label, false, 0)
	cur := imgui.CursorPos()
	imgui.SetCursorPosX(cur.X + (cellWidth-textSize.X)*0.5)
	imgui.TextUnformatted(label)
}

func renderDeletePrefabConfirmationPopup() {
	if pendingDeletePrefab == nil {
		return
	}

	renderConfirmationModal(
		deletePrefabConfirmationPopup,
		fmt.Sprintf("Delete prefab [%s]?", string(*pendingDeletePrefab)),
		&showDeletePrefabConfirmationPopup,
		func() {
			prefab.Delete(*pendingDeletePrefab)
			pendingDeletePrefab = nil
		},
		func() {
			pendingDeletePrefab = nil
		},
	)
}
