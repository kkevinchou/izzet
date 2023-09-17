package panels

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
)

func stats(world World, renderContext RenderContext) {
	if imgui.CollapsingHeaderV("Rendering", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)
		setupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", DBG.RenderTime)) })
		setupRow("FPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", DBG.FPS)) })
		setupRow("Triangle Draw Count", func() { imgui.LabelText("", formatNumber(DBG.TriangleDrawCount)) })
		setupRow("Draw Count", func() { imgui.LabelText("", formatNumber(DBG.DrawCount)) })
		setupRow("Texture", func() {
			imgui.PushItemWidth(tableColumn1Width)
			if imgui.BeginCombo("", string(SelectedComboOption)) {
				for _, option := range comboOptions {
					if imgui.Selectable(string(option)) {
						SelectedComboOption = option
					}
				}
				imgui.EndCombo()
			}
			imgui.PopItemWidth()
		})
		setupRow("Texture Viewer", func() {
			if DBG.DebugTexture != 0 {
				var imageWidth float32 = 500
				texture := createUserSpaceTextureHandle(DBG.DebugTexture)
				size := imgui.Vec2{X: imageWidth, Y: imageWidth / float32(renderContext.AspectRatio())}
				// invert the Y axis since opengl vs texture coordinate systems differ
				// https://learnopengl.com/Getting-started/Textures
				imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
			}
		})
		imgui.EndTable()
	}
}
