package panels

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
)

func stats(world World, renderContext RenderContext) {
	if imgui.CollapsingHeaderV("Rendering", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)
		setupRow("Command Frames Before Render", func() { imgui.LabelText("", fmt.Sprintf("%d", DBG.CommandFramesPerRender)) })
		setupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", DBG.RenderTime)) })
		setupRow("Command Frame Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", DBG.CommandFrameTime)) })
		setupRow("FPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", DBG.FPS)) })
		setupRow("Triangle Draw Count", func() { imgui.LabelText("", formatNumber(DBG.TriangleDrawCount)) })
		setupRow("Draw Count", func() { imgui.LabelText("", formatNumber(DBG.DrawCount)) })
		setupRow("Texture Viewer Table Row", func() {
			if DBG.DebugTexture != 0 {
				if imgui.Button("Toggle Texture Window") {
					DBG.ShowDebugTexture = !DBG.ShowDebugTexture
				}

				if DBG.ShowDebugTexture {
					imgui.SetNextWindowSizeV(imgui.Vec2{X: 400}, imgui.ConditionFirstUseEver)
					if imgui.BeginV("Texture Viewer", &DBG.ShowDebugTexture, imgui.WindowFlagsNone) {
						if imgui.BeginCombo("", string(SelectedComboOption)) {
							for _, option := range comboOptions {
								if imgui.Selectable(string(option)) {
									SelectedComboOption = option
								}
							}
							imgui.EndCombo()
						}

						regionSize := imgui.ContentRegionAvail()
						imageWidth := regionSize.X

						texture := createUserSpaceTextureHandle(DBG.DebugTexture)
						size := imgui.Vec2{X: imageWidth, Y: imageWidth / float32(renderContext.AspectRatio())}
						// invert the Y axis since opengl vs texture coordinate systems differ
						// https://learnopengl.com/Getting-started/Textures
						imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
					}
					imgui.End()
				}
			}
		})
		imgui.EndTable()
	}
}
