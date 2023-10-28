package panels

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
)

func stats(app App, renderContext RenderContext) {
	settings := app.Settings()

	if imgui.CollapsingHeaderV("Rendering", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)

		// Frame Profiling
		setupRow("Command Frames Before Render", func() { imgui.LabelText("", fmt.Sprintf("%d", settings.CommandFramesPerRender)) }, true)
		setupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", settings.RenderTime)) }, true)
		setupRow("Command Frame Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", settings.CommandFrameTime)) }, true)
		setupRow("FPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", settings.FPS)) }, true)
		setupRow("Command Frame", func() { imgui.LabelText("", fmt.Sprintf("%d", app.CommandFrame())) }, true)

		// Rendering
		setupRow("Shadow Far Factor", func() { imgui.SliderFloat("", &settings.ShadowFarFactor, 0, 10) }, true)
		setupRow("Triangle Draw Count", func() { imgui.LabelText("", formatNumber(settings.TriangleDrawCount)) }, true)
		setupRow("Draw Count", func() { imgui.LabelText("", formatNumber(settings.DrawCount)) }, true)

		setupRow("Texture Viewer Table Row", func() {
			if settings.DebugTexture != 0 {
				if imgui.Button("Toggle Texture Window") {
					settings.ShowDebugTexture = !settings.ShowDebugTexture
				}

				if settings.ShowDebugTexture {
					imgui.SetNextWindowSizeV(imgui.Vec2{X: 400}, imgui.ConditionFirstUseEver)
					if imgui.BeginV("Texture Viewer", &settings.ShowDebugTexture, imgui.WindowFlagsNone) {
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

						texture := CreateUserSpaceTextureHandle(settings.DebugTexture)
						size := imgui.Vec2{X: imageWidth, Y: imageWidth / float32(renderContext.AspectRatio())}
						// invert the Y axis since opengl vs texture coordinate systems differ
						// https://learnopengl.com/Getting-started/Textures
						imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
					}
					imgui.End()
				}
			}
		}, true)
		imgui.EndTable()
	}
}
