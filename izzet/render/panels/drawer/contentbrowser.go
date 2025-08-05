package drawer

import (
	"C"
	"fmt"

	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)
import (
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
)

func contentBrowser(app renderiface.App) bool {
	var menuOpen bool

	var width float32 = 100
	const maxPerRow = 5

	for i, document := range app.AssetManager().GetDocuments() {
		doc := document.Document
		size := imgui.Vec2{X: width, Y: 100}
		// invert the Y axis since opengl vs texture coordinate systems differ
		// https://learnopengl.com/Getting-started/Textures
		imgui.BeginGroup()

		imgui.Dummy(imgui.Vec2{X: 10, Y: 10})

		imgui.PushIDStr(fmt.Sprintf("image %d", i))

		if documentTexture == nil {
			t := app.AssetManager().GetTexture("document")
			texture := imgui.TextureID(t.ID)
			documentTexture = &texture
		}

		imgui.ImageV(*documentTexture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
		if imgui.BeginPopupContextItemV("NULL", imgui.PopupFlagsMouseButtonRight) {
			menuOpen = true
			if imgui.Button("Instantiate") {
				app.CreateEntitiesFromDocumentAsset(document)
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
		imgui.PopID()

		if imgui.IsItemHovered() {
			imgui.BeginTooltip()
			imgui.Text(doc.Name)
			imgui.EndTooltip()
		}

		if imgui.BeginDragDropSourceV(imgui.DragDropFlagsSourceAllowNullID) {
			s := doc.Name
			ptr := unsafe.Pointer(&s)
			size := uint64(unsafe.Sizeof(doc.Name))
			imgui.SetDragDropPayloadV("content_browser_item", uintptr(ptr), size, imgui.CondOnce)
			imgui.EndDragDropSource()
		}

		if imgui.BeginChildStrV(
			fmt.Sprintf("##%s", doc.Name),
			imgui.Vec2{X: width, Y: 0},
			imgui.ChildFlagsNone,
			imgui.WindowFlagsNoScrollbar|imgui.WindowFlagsNoScrollWithMouse,
		) {
			imgui.Text(doc.Name)
		}
		imgui.EndChild()

		imgui.EndGroup()
		if (i+1)%(maxPerRow) != 0 {
			imgui.SameLine()
		}
	}

	return menuOpen
}
