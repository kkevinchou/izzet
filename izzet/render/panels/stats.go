package panels

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/internal/metrics"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	metricRange          = 2
	renderingMetricRange = 5
)

type metricPair struct {
	name  string
	value float64
}

func Stats(app renderiface.App, renderContext RenderContext) {
	mr := globals.ClientRegistry()
	runtimeConfig := app.RuntimeConfig()
	caser := cases.Title(language.English)

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("Render Time", func() { imgui.LabelText("", fmt.Sprintf("%.2f", mr.AvgOver("renderer_cpu_time", metricRange))) }, true)
		panelutils.SetupRow("Command Frame Time", func() {
			imgui.LabelText("", fmt.Sprintf("%.2f", mr.AvgOver("command_frame_nanoseconds", metricRange)/1000000))
		}, true)
		panelutils.SetupRow("Client Sleep Time", func() {
			imgui.LabelText("", fmt.Sprintf("%.2f", mr.AvgOver("client_sleep_nanoseconds", metricRange)/1000000))
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

	// rendering metrics tracked from gpu
	pairs, total := metricPairsByPrefix(mr, "render_gpu_")

	if imgui.CollapsingHeaderTreeNodeFlagsV(fmt.Sprintf("Rendering - GPU (%.2f)###gpu_rendering_header", total), imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		for _, pair := range pairs {
			panelutils.SetupRow(caser.String(pair.name), func() { imgui.LabelText("", fmt.Sprintf("%.2f", pair.value)) }, true)
		}
		imgui.EndTable()
	}

	// rendering metrics tracked from cpu
	pairs, total = metricPairsByPrefix(mr, "render_cpu_")

	if imgui.CollapsingHeaderTreeNodeFlagsV(fmt.Sprintf("Rendering - CPU (%.2f)###cpu_rendering_header", total), imgui.TreeNodeFlagsNone) {
		imgui.BeginTableV("", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()
		for _, pair := range pairs {
			panelutils.SetupRow(caser.String(pair.name), func() { imgui.LabelText("", fmt.Sprintf("%.2f", pair.value)) }, true)
		}
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

func metricPairsByPrefix(mr *metrics.Registry, prefix string) ([]metricPair, float64) {
	// TODO - this should probably support metric tags rather than using prefixes
	metrics := mr.MetricsByPrefix(prefix)

	var pairs []metricPair
	for _, metric := range metrics {
		pairs = append(
			pairs,
			metricPair{
				name:  strings.ReplaceAll(strings.TrimPrefix(metric, prefix), "_", " "),
				value: mr.AvgOver(metric, renderingMetricRange),
			},
		)
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].value == pairs[j].value {
			return pairs[i].name < pairs[j].name
		}
		return pairs[i].value > pairs[j].value
	})

	var total float64
	for _, pair := range pairs {
		total += pair.value
	}

	return pairs, total
}
