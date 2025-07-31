package panels

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func stats(app renderiface.App, renderContext RenderContext) {
	runtimeConfig := app.RuntimeConfig()
	mr := app.MetricsRegistry()

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_time")))) }, true)
		panelutils.SetupRow("Command Frame Time", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage("command_frame_nanoseconds")/1000000))
		}, true)
		panelutils.SetupRow("FPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondSum("fps"))) }, true)
		panelutils.SetupRow("CFPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondSum("command_frames"))) }, true)
		panelutils.SetupRow("Command Frame", func() { imgui.LabelText("", fmt.Sprintf("%d", app.CommandFrame())) }, true)
		panelutils.SetupRow("Ping", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetLatest("ping")))) }, true)
		panelutils.SetupRow("Prediction Hit", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetOneSecondSum("prediction_hit")))) }, true)
		panelutils.SetupRow("Prediction Miss", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetOneSecondSum("prediction_miss")))) }, true)

		panelutils.SetupRow("Triangle Draw Count", func() { imgui.LabelText("", formatNumber(runtimeConfig.TriangleDrawCount)) }, true)
		panelutils.SetupRow("Draw Count", func() { imgui.LabelText("", formatNumber(runtimeConfig.DrawCount)) }, true)
		panelutils.SetupRow("Draw Entity Count", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetLatest("draw_entity_count")))) }, true)
		panelutils.SetupRow("gl.GenBuffers() count", func() { imgui.LabelText("", fmt.Sprintf("%0.f", mr.GetOneSecondSum("gen_buffers"))) }, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Rendering", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		// Frame Profiling
		panelutils.SetupRow("Render Main Color Buffer", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_main_color_buffer")))) }, true)
		panelutils.SetupRow("Render Geometry Pass", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_gpass")))) }, true)
		panelutils.SetupRow("Render SSAO", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_ssao")))) }, true)
		panelutils.SetupRow("Render Depthmaps", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_depthmaps")))) }, true)
		panelutils.SetupRow("Render Query Shadowcasting", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_query_shadowcasting"))))
		}, true)
		panelutils.SetupRow("Render Query Renderable", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_query_renderable")))) }, true)
		panelutils.SetupRow("Render Swap", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_swap")))) }, true)
		panelutils.SetupRow("Render Annotations", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_annotations")))) }, true)
		panelutils.SetupRow("Render Context Setup", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_context_setup")))) }, true)
		panelutils.SetupRow("Render Skybox", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_skybox")))) }, true)
		panelutils.SetupRow("Render Volumetrics", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_volumetrics")))) }, true)
		panelutils.SetupRow("Render Gizmos", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_gizmos")))) }, true)
		panelutils.SetupRow("Render Colorpicking", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_colorpicking")))) }, true)
		panelutils.SetupRow("Render Bloom Pass", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_bloom")))) }, true)
		panelutils.SetupRow("Render Buffer Setup", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_buffer_setup")))) }, true)
		panelutils.SetupRow("Render Post Process", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_post_process")))) }, true)
		panelutils.SetupRow("Render Imgui", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_imgui")))) }, true)
		panelutils.SetupRow("Render Sleep", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.GetOneSecondAverage(("render_sleep")))) }, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Server Stats", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		stats := app.GetServerStats()
		for _, stat := range stats.Data {
			panelutils.SetupRow(stat.Name, func() { imgui.LabelText(stat.Name, stat.Value) }, true)
		}
		imgui.EndTable()
	}
}
