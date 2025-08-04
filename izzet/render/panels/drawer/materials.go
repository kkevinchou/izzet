package drawer

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/windows"
)

const (
	cellWidth   float32 = 100
	cellHeight  float32 = 100
	itemsPerRow int32   = 7
	iconPadding float32 = 10
)

func materialssUI(app renderiface.App, materialTextureMap map[string]uint32) {
	mats := app.AssetManager().GetMaterials()

	style := imgui.CurrentStyle()
	imgui.PushStyleVarVec2(
		imgui.StyleVarCellPadding,
		imgui.Vec2{X: style.CellPadding().X, Y: 5},
	)
	defer imgui.PopStyleVar()

	// 1) Begin a fixed-column table to handle layout for us
	if imgui.BeginTableV("MaterialsTable", itemsPerRow,
		imgui.TableFlagsSizingFixedSame, imgui.Vec2{X: 0, Y: 0}, 0) {

		for i, mat := range mats {
			imgui.TableNextColumn()
			drawMaterialCell(app, mat, materialTextureMap[mat.Name], i)
		}

		imgui.EndTable()
	}
}

func drawMaterialCell(app renderiface.App, material assets.MaterialAsset, textureID uint32, idx int) {
	// push a stable ID so nothing collides
	imgui.PushIDInt(int32(idx))
	defer imgui.PopID()

	if textureID == 0 {
		textureID = app.AssetManager().GetTexture("document").ID
	}

	// draw the thumbnail
	imgui.ImageV(
		imgui.TextureID(textureID),
		imgui.Vec2{X: cellWidth, Y: cellHeight},
		imgui.Vec2{X: 0, Y: 1},
		imgui.Vec2{X: 1, Y: 0},
		imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1},
		imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0},
	)

	// right-click menu
	if imgui.BeginPopupContextItemV(material.Name, imgui.PopupFlagsMouseButtonRight) {
		if imgui.Button("Edit") {
			material := app.AssetManager().GetMaterial(material.Handle)
			windows.ShowEditMaterialWindow(app, material)
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}

	// tooltip on hover
	if imgui.IsItemHovered() {
		imgui.BeginTooltip()
		imgui.Text(material.Name)
		imgui.EndTooltip()
	}

	label := ellipsize(material.Name, cellWidth)
	imgui.TextUnformatted(label)
}

// ellipsize returns a version of s that fits within maxWidth, adding "…" if it had to cut.
func ellipsize(s string, maxWidth float32) string {
	// measure full string
	fullSize := imgui.CalcTextSizeV(s, false, 0)
	if fullSize.X <= maxWidth {
		return s
	}

	ell := "..."
	ellSize := imgui.CalcTextSizeV(ell, false, 0)

	// accumulate runes until we hit (maxWidth - ellipsis width)
	var out []rune
	width := float32(0)
	for _, r := range s {
		w := imgui.CalcTextSizeV(string(r), false, 0).X
		if width+w+ellSize.X > maxWidth {
			break
		}
		out = append(out, r)
		width += w
	}
	return string(out) + ell
}
