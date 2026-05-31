package drawer

import (
	"C"

	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)
import (
	"fmt"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/assets"
)

const (
	deleteDocumentConfirmationPopup = "Delete Document"
	deleteDocumentBlockedPopup      = "Cannot Delete Document"
)

var blockedDeleteDocumentName string
var blockedDeleteDocumentEntityIDs []int
var showDeleteDocumentBlockedPopup bool

var pendingDeleteDocument *assets.DocumentAsset
var showDeleteDocumentConfirmationPopup bool

func contentBrowser(app renderiface.App) {
	style := imgui.CurrentStyle()
	imgui.PushStyleVarVec2(
		imgui.StyleVarCellPadding,
		imgui.Vec2{X: style.CellPadding().X, Y: 5},
	)
	defer imgui.PopStyleVar()

	if imgui.BeginTableV("DocumentsTable", itemsPerRow,
		imgui.TableFlagsSizingFixedSame, imgui.Vec2{X: 0, Y: 0}, 0) {

		for _, document := range app.AssetManager().GetDocuments() {
			imgui.TableNextColumn()
			drawDocumentCell(app, document)
		}
		imgui.EndTable()
	}

	renderDeleteDocumentConfirmationPopup(app)
	renderDeleteDocumentBlockedPopup()
}

func drawDocumentCell(app renderiface.App, documentAsset assets.DocumentAsset) {
	documentName := documentAsset.Document.Name
	documentID := documentAsset.Config.Name

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

	if imgui.BeginPopupContextItemV(fmt.Sprintf("document-context-%s", documentID), imgui.PopupFlagsMouseButtonRight) {
		if imgui.Button("Instantiate Entities") {
			app.CreateEntitiesFromDocumentAsset(documentAsset, false)
			imgui.CloseCurrentPopup()
		}
		if imgui.Button("Instantiate As Merged Entity") {
			app.CreateEntitiesFromDocumentAsset(documentAsset, true)
			imgui.CloseCurrentPopup()
		}
		if imgui.Button("Delete") {
			pendingDeleteDocument = &documentAsset
			showDeleteDocumentConfirmationPopup = true
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}

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

func renderDeleteDocumentConfirmationPopup(app renderiface.App) {
	if pendingDeleteDocument == nil {
		return
	}

	center := imgui.MainViewport().Center()
	imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})

	if showDeleteDocumentConfirmationPopup {
		imgui.OpenPopupStr(deleteDocumentConfirmationPopup)
		showDeleteDocumentConfirmationPopup = false
	}

	if imgui.BeginPopupModalV(deleteDocumentConfirmationPopup, nil, imgui.WindowFlagsAlwaysAutoResize) {
		imgui.Text(fmt.Sprintf("Delete document [%s]?", pendingDeleteDocument.Config.Name))
		imgui.Separator()
		if imgui.Button("Delete") {
			referencingEntityIDs := app.DeleteDocument(*pendingDeleteDocument)
			if len(referencingEntityIDs) > 0 {
				blockedDeleteDocumentName = pendingDeleteDocument.Config.Name
				blockedDeleteDocumentEntityIDs = referencingEntityIDs
				showDeleteDocumentBlockedPopup = true
			}
			pendingDeleteDocument = nil
			imgui.CloseCurrentPopup()
		}
		imgui.SameLine()
		if imgui.Button("Cancel") {
			pendingDeleteDocument = nil
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}
}

func renderDeleteDocumentBlockedPopup() {
	if len(blockedDeleteDocumentEntityIDs) == 0 {
		return
	}

	center := imgui.MainViewport().Center()
	imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})

	if showDeleteDocumentBlockedPopup {
		imgui.OpenPopupStr(deleteDocumentBlockedPopup)
		showDeleteDocumentBlockedPopup = false
	}

	if imgui.BeginPopupModalV(deleteDocumentBlockedPopup, nil, imgui.WindowFlagsAlwaysAutoResize) {
		imgui.Text(fmt.Sprintf("Document [%s] is still referenced by entities.", blockedDeleteDocumentName))
		imgui.Separator()
		imgui.Text("Referencing entity IDs:")
		for _, entityID := range blockedDeleteDocumentEntityIDs {
			imgui.Text(fmt.Sprintf("%d", entityID))
		}
		if imgui.Button("OK") {
			blockedDeleteDocumentEntityIDs = nil
			blockedDeleteDocumentName = ""
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}
}
