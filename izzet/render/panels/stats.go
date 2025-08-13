package panels

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func stats(app renderiface.App, renderContext RenderContext) {
	runtimeConfig := app.RuntimeConfig()
	mr := globals.ClientRegistry()

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_time", 3))) }, true)
		panelutils.SetupRow("Command Frame Time", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.PerSecondRateOver("command_frame_nanoseconds", 3)/1000000))
		}, true)
		panelutils.SetupRow("FPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.PerSecondRateOver("fps", 3))) }, true)
		panelutils.SetupRow("CFPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.PerSecondRateOver("command_frames", 3))) }, true)
		panelutils.SetupRow("Command Frame", func() { imgui.LabelText("", fmt.Sprintf("%d", app.CommandFrame())) }, true)
		// panelutils.SetupRow("Ping", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetLatest("ping")))) }, true)
		panelutils.SetupRow("Prediction Hit", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.SumOver("prediction_hit", 3)))) }, true)
		panelutils.SetupRow("Prediction Miss", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.SumOver("prediction_miss", 3)))) }, true)

		panelutils.SetupRow("Triangle Draw Count", func() { imgui.LabelText("", formatNumber(runtimeConfig.TriangleDrawCount)) }, true)
		panelutils.SetupRow("Draw Count", func() { imgui.LabelText("", formatNumber(runtimeConfig.DrawCount)) }, true)
		// panelutils.SetupRow("Draw Entity Count", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetLatest("draw_entity_count")))) }, true)
		panelutils.SetupRow("gl.GenBuffers() count", func() { imgui.LabelText("", fmt.Sprintf("%0.f", mr.SumOver("gen_buffers", 3))) }, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Rendering", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		// Frame Profiling
		panelutils.SetupRow("Render Main Color Buffer", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_main", 3))) }, true)
		panelutils.SetupRow("Render Geometry Pass", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_gpass", 3))) }, true)
		panelutils.SetupRow("Render SSAO", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_ssao", 3))) }, true)
		panelutils.SetupRow("Render Depthmaps", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_depthmaps", 3))) }, true)
		panelutils.SetupRow("Render Query Shadowcasting", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_query_shadowcasting", 3)))
		}, true)
		panelutils.SetupRow("Render Query Renderable", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_query_renderable", 3))) }, true)
		panelutils.SetupRow("Render Swap", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_swap", 3))) }, true)
		panelutils.SetupRow("Render Annotations", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_annotations", 3))) }, true)
		panelutils.SetupRow("Render Context Setup", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_context_setup", 3))) }, true)
		panelutils.SetupRow("Render Skybox", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_skybox", 3))) }, true)
		panelutils.SetupRow("Render Volumetrics", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_volumetrics", 3))) }, true)
		panelutils.SetupRow("Render Gizmos", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_gizmos", 3))) }, true)
		panelutils.SetupRow("Render Colorpicking", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_colorpicking", 3))) }, true)
		panelutils.SetupRow("Render Bloom Pass", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_bloom", 3))) }, true)
		panelutils.SetupRow("Render Buffer Setup", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_buffer_setup", 3))) }, true)
		panelutils.SetupRow("Render Post Process", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_post_process", 3))) }, true)
		panelutils.SetupRow("Render Imgui", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_imgui", 3))) }, true)
		panelutils.SetupRow("Render Sleep", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AverageValueOver("render_sleep", 3))) }, true)

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
