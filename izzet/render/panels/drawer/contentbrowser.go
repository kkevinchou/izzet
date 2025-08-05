package drawer

import (
	"C"

	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)
import (
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/assets"
)

// const (
// 	width       float32 = 100
// 	maxPerRow   int     = 7
// 	cellWidth   float32 = 100
// 	cellHeight  float32 = 100
// 	itemsPerRow int32   = 7
// 	iconPadding float32 = 10
// )

func contentBrowser(app renderiface.App) {
	style := imgui.CurrentStyle()
	imgui.PushStyleVarVec2(
		imgui.StyleVarCellPadding,
		imgui.Vec2{X: style.CellPadding().X, Y: 5},
	)
	defer imgui.PopStyleVar()

	if imgui.BeginTableV("DocumentsTable", itemsPerRow,
		imgui.TableFlagsSizingFixedSame, imgui.Vec2{X: 0, Y: 0}, 0) {

		for i, document := range app.AssetManager().GetDocuments() {
			imgui.TableNextColumn()
			drawDocumentCell(app, document, i)
		}
		imgui.EndTable()
	}
}

func drawDocumentCell(app renderiface.App, documentAsset assets.DocumentAsset, idx int) {
	documentName := documentAsset.Document.Name
	imgui.PushIDInt(int32(idx))

	t := app.AssetManager().GetTexture("document")

	// invert the Y axis since opengl vs texture coordinate systems differ
	// https://learnopengl.com/Getting-started/Textures

	// draw the thumbnail
	imgui.ImageV(
		imgui.TextureID(t.ID),
		imgui.Vec2{X: cellWidth, Y: cellHeight},
		imgui.Vec2{X: 0, Y: 1},
		imgui.Vec2{X: 1, Y: 0},
		imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1},
		imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0},
	)

	if imgui.BeginPopupContextItemV("NULL", imgui.PopupFlagsMouseButtonRight) {
		if imgui.Button("Instantiate") {
			app.CreateEntitiesFromDocumentAsset(documentAsset)
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}
	imgui.PopID()

	if imgui.IsItemHovered() {
		imgui.BeginTooltip()
		imgui.Text(documentName)
		imgui.EndTooltip()
	}

	if imgui.BeginDragDropSourceV(imgui.DragDropFlagsSourceAllowNullID) {
		s := documentName
		ptr := unsafe.Pointer(&s)
		size := uint64(unsafe.Sizeof(documentName))
		imgui.SetDragDropPayloadV("content_browser_item", uintptr(ptr), size, imgui.CondOnce)
		imgui.EndDragDropSource()
	}

	label := ellipsize(documentName, cellWidth)
	// center text
	textSize := imgui.CalcTextSizeV(label, false, 0)
	cur := imgui.CursorPos()
	imgui.SetCursorPosX(cur.X + (cellWidth-textSize.X)*0.5)
	imgui.TextUnformatted(label)
}
