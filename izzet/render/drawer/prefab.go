package drawer

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/prefab"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

const deletePrefabConfirmationPopup = "Delete Prefab"

var pendingDeletePrefab *prefab.Prefab
var showDeletePrefabConfirmationPopup bool

func prefabsUI(app renderiface.App) {
	style := imgui.CurrentStyle()
	imgui.PushStyleVarVec2(
		imgui.StyleVarCellPadding,
		imgui.Vec2{X: style.CellPadding().X, Y: 5},
	)
	defer imgui.PopStyleVar()

	prefabs := prefab.Prefabs()
	if imgui.BeginTableV("PrefabsTable", itemsPerRow,
		imgui.TableFlagsSizingFixedSame, imgui.Vec2{X: 0, Y: 0}, 0) {

		for _, p := range prefabs {
			imgui.TableNextColumn()
			drawPrefabCell(app, p)
		}
		imgui.EndTable()
	}

	renderDeletePrefabConfirmationPopup()
}

func drawPrefabCell(app renderiface.App, p prefab.Prefab) {
	name := p.Name

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
			e := prefab.Instantiate(p.Handle, app.AssetManager())
			app.World().AddEntity(e)
			app.SelectEntity(e)
			imgui.CloseCurrentPopup()
		}
		if imgui.Button("Delete") {
			pendingDeletePrefab = &p
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
		fmt.Sprintf("Delete prefab [%s]?", pendingDeletePrefab.Handle),
		&showDeletePrefabConfirmationPopup,
		func() {
			prefab.Delete(pendingDeletePrefab.Handle)
			pendingDeletePrefab = nil
		},
		func() {
			pendingDeletePrefab = nil
		},
	)
}
