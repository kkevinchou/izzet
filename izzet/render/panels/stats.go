package panels

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func stats(app renderiface.App, renderContext RenderContext) {
	runtimeConfig := app.RuntimeConfig()
	mr := app.MetricsRegistry()

	if imgui.CollapsingHeaderTreeNodeFlagsV("Rendering", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("Bloom Table", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		// imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)

		// Frame Profiling
		panelutils.SetupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_time")))) }, true)
		panelutils.SetupRow("Render Swap", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_swap")))) }, true)
		panelutils.SetupRow("Render Context Setup", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_context_setup")))) }, true)
		panelutils.SetupRow("Render Query Renderable", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_query_renderable")))) }, true)
		panelutils.SetupRow("Render Query Shadowcasting", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_query_shadowcasting"))))
		}, true)
		panelutils.SetupRow("Render Skybox", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_skybox")))) }, true)
		panelutils.SetupRow("Render Depthmaps", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_depthmaps")))) }, true)
		panelutils.SetupRow("Render Main Color Buffer", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_main_color_buffer")))) }, true)
		panelutils.SetupRow("Render Gizmos", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_gizmos")))) }, true)
		panelutils.SetupRow("Render Colorpicking", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_colorpicking")))) }, true)
		panelutils.SetupRow("Render Bloom Pass", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_bloom")))) }, true)
		panelutils.SetupRow("Render Buffer Setup", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_buffer_setup")))) }, true)
		panelutils.SetupRow("Render Post Process", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_post_process")))) }, true)
		panelutils.SetupRow("Render Imgui", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_imgui")))) }, true)
		panelutils.SetupRow("Command Frame Time", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage("command_frame_nanoseconds")/1000000))
		}, true)
		panelutils.SetupRow("FPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondSum("fps"))) }, true)
		panelutils.SetupRow("Command Frame", func() { imgui.LabelText("", fmt.Sprintf("%d", app.CommandFrame())) }, true)
		panelutils.SetupRow("Prediction Hit", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetOneSecondSum("prediction_hit")))) }, true)
		panelutils.SetupRow("Prediction Miss", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetOneSecondSum("prediction_miss")))) }, true)
		panelutils.SetupRow("Ping", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetLatest("ping")))) }, true)

		// Rendering
		panelutils.SetupRow("Triangle Draw Count", func() { imgui.LabelText("", formatNumber(runtimeConfig.TriangleDrawCount)) }, true)
		panelutils.SetupRow("Draw Count", func() { imgui.LabelText("", formatNumber(runtimeConfig.DrawCount)) }, true)
		panelutils.SetupRow("gl.GenBuffers() count", func() { imgui.LabelText("", fmt.Sprintf("%0.f", mr.GetOneSecondSum("gen_buffers"))) }, true)

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
