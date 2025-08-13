package panels

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

const (
	metricRange = 2
)

func stats(app renderiface.App, renderContext RenderContext) {
	runtimeConfig := app.RuntimeConfig()
	mr := globals.ClientRegistry()

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_time", metricRange))) }, true)
		panelutils.SetupRow("Command Frame Time", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.RatePerSec("command_frame_nanoseconds", metricRange)/1000000))
		}, true)
		panelutils.SetupRow("FPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.RatePerSec("fps", metricRange))) }, true)
		panelutils.SetupRow("CFPS", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.RatePerSec("command_frames", metricRange))) }, true)
		panelutils.SetupRow("Command Frame", func() { imgui.LabelText("", fmt.Sprintf("%d", app.CommandFrame())) }, true)
		// panelutils.SetupRow("Ping", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetLatest("ping")))) }, true)
		panelutils.SetupRow("Prediction Hit", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.RatePerSec("prediction_hit", metricRange)))) }, true)
		panelutils.SetupRow("Prediction Miss", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.RatePerSec("prediction_miss", metricRange)))) }, true)

		panelutils.SetupRow("Triangle Draw Count", func() { imgui.LabelText("", formatNumber(runtimeConfig.TriangleDrawCount)) }, true)
		panelutils.SetupRow("Draw Count", func() { imgui.LabelText("", formatNumber(runtimeConfig.DrawCount)) }, true)
		// panelutils.SetupRow("Draw Entity Count", func() { imgui.LabelText("", fmt.Sprintf("%d", int(mr.GetLatest("draw_entity_count")))) }, true)
		panelutils.SetupRow("gl.GenBuffers() count", func() { imgui.LabelText("", fmt.Sprintf("%0.f", mr.RatePerSec("gen_buffers", metricRange))) }, true)

		imgui.EndTable()
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Rendering", imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		// Frame Profiling
		panelutils.SetupRow("Render Main Color Buffer", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_main", metricRange))) }, true)
		panelutils.SetupRow("Render Geometry Pass", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_gpass", metricRange))) }, true)
		panelutils.SetupRow("Render SSAO", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_ssao", metricRange))) }, true)
		panelutils.SetupRow("Render Depthmaps", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_depthmaps", metricRange))) }, true)
		panelutils.SetupRow("Render Query Shadowcasting", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_query_shadowcasting", metricRange)))
		}, true)
		panelutils.SetupRow("Render Query Renderable", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_query_renderable", metricRange)))
		}, true)
		panelutils.SetupRow("Render Swap", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_swap", metricRange))) }, true)
		panelutils.SetupRow("Render Annotations", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_annotations", metricRange)))
		}, true)
		panelutils.SetupRow("Render Context Setup", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_context_setup", metricRange)))
		}, true)
		panelutils.SetupRow("Render Skybox", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_skybox", metricRange))) }, true)
		panelutils.SetupRow("Render Volumetrics", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_volumetrics", metricRange)))
		}, true)
		panelutils.SetupRow("Render Gizmos", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_gizmos", metricRange))) }, true)
		panelutils.SetupRow("Render Colorpicking", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_colorpicking", metricRange)))
		}, true)
		panelutils.SetupRow("Render Bloom Pass", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_bloom", metricRange))) }, true)
		panelutils.SetupRow("Render Buffer Setup", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_buffer_setup", metricRange)))
		}, true)
		panelutils.SetupRow("Render Post Process", func() {
			imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_post_process", metricRange)))
		}, true)
		panelutils.SetupRow("Render Imgui", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_imgui", metricRange))) }, true)
		panelutils.SetupRow("Render Sleep", func() { imgui.LabelText("", fmt.Sprintf("%.1f", mr.AvgOver("render_sleep", metricRange))) }, true)

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
