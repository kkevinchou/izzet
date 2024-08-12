package panels

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

type DebugComboOption string

const (
	ComboOptionFinalRender    DebugComboOption = "FINALRENDER"
	ComboOptionColorPicking   DebugComboOption = "COLORPICKING"
	ComboOptionHDR            DebugComboOption = "HDR (bloom only)"
	ComboOptionBloom          DebugComboOption = "BLOOMTEXTURE (bloom only)"
	ComboOptionShadowDepthMap DebugComboOption = "SHADOW DEPTH MAP"
	ComboOptionCameraDepthMap DebugComboOption = "CAMERA DEPTH MAP"
	ComboOptionCubeDepthMap   DebugComboOption = "CUBE DEPTH MAP"
)

var SelectedDebugComboOption DebugComboOption = ComboOptionFinalRender

var (
	debugComboOptions []DebugComboOption = []DebugComboOption{
		ComboOptionFinalRender,
		ComboOptionColorPicking,
		ComboOptionHDR,
		ComboOptionBloom,
		ComboOptionShadowDepthMap,
		ComboOptionCameraDepthMap,
	}
)

func stats(app renderiface.App, renderContext RenderContext) {
	settings := app.RuntimeConfig()
	mr := app.MetricsRegistry()

	if imgui.CollapsingHeaderTreeNodeFlagsV("Rendering", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		// imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)

		// Frame Profiling
		panelutils.SetupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", settings.RenderTime)) }, true)
		panelutils.SetupRow("Command Frame Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", settings.CommandFrameTime)) }, true)
		panelutils.SetupRow("FPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", settings.FPS)) }, true)
		panelutils.SetupRow("Command Frame", func() { imgui.LabelText("", fmt.Sprintf("%d", app.CommandFrame())) }, true)
		panelutils.SetupRow("Prediction Hit", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetOneSecondSum("prediction_hit")))) }, true)
		panelutils.SetupRow("Prediction Miss", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetOneSecondSum("prediction_miss")))) }, true)
		panelutils.SetupRow("Ping", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetLatest("ping")))) }, true)

		// Rendering
		panelutils.SetupRow("Triangle Draw Count", func() { imgui.LabelText("", formatNumber(settings.TriangleDrawCount)) }, true)
		panelutils.SetupRow("Draw Count", func() { imgui.LabelText("", formatNumber(settings.DrawCount)) }, true)
		panelutils.SetupRow("gl.GenBuffers() count", func() { imgui.LabelText("", fmt.Sprintf("%0.f", mr.GetOneSecondSum("gen_buffers"))) }, true)

		panelutils.SetupRow("Texture Viewer", func() {
			if settings.DebugTexture != 0 {
				if imgui.Button("Toggle Texture Window") {
					settings.ShowDebugTexture = !settings.ShowDebugTexture
				}

				if settings.ShowDebugTexture {
					imgui.SetNextWindowSizeV(imgui.Vec2{X: 400}, imgui.CondFirstUseEver)
					if imgui.BeginV("Texture Viewer", &settings.ShowDebugTexture, imgui.WindowFlagsNone) {
						if imgui.BeginCombo("##", string(SelectedDebugComboOption)) {
							for _, option := range debugComboOptions {
								if imgui.SelectableBool(string(option)) {
									SelectedDebugComboOption = option
								}
							}
							imgui.EndCombo()
						}

						regionSize := imgui.ContentRegionAvail()
						imageWidth := regionSize.X

						texture := imgui.TextureID{Data: uintptr(settings.DebugTexture)}
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
	if imgui.CollapsingHeaderTreeNodeFlagsV("Server Stats", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Server Stats Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		stats := app.GetServerStats()
		for _, stat := range stats.Data {
			panelutils.SetupRow(stat.Name, func() { imgui.LabelText(stat.Name, stat.Value) }, true)
		}
		imgui.EndTable()
	}
}
