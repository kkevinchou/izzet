package panels

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/internal/metrics"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/ui"
	"github.com/kkevinchou/izzet/izzet/telemetry"
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
	mr := telemetry.ClientRegistry()
	runtimeConfig := app.RuntimeConfig()
	caser := cases.Title(language.English)

	if imgui.CollapsingHeaderTreeNodeFlagsV("General", imgui.TreeNodeFlagsDefaultOpen) {
		ui.Table("", func() {
			ui.LabelRow("Render Time", fmt.Sprintf("%.2f", mr.AvgOver("renderer_cpu_time", metricRange)))
			ui.LabelRow("Command Frame Time", fmt.Sprintf("%.2f", mr.AvgOver("command_frame_nanoseconds", metricRange)/1000000))
			ui.LabelRow("Client Sleep Time", fmt.Sprintf("%.2f", mr.AvgOver("client_sleep_nanoseconds", metricRange)/1000000))
			ui.LabelRow("FPS", fmt.Sprintf("%.1f", mr.RatePerSec("fps", metricRange)))
			ui.LabelRow("CFPS", fmt.Sprintf("%.1f", mr.RatePerSec("command_frames", metricRange)))
			ui.LabelRow("Command Frame", fmt.Sprintf("%d", app.CommandFrame()))
			ui.LabelRow("Ping", fmt.Sprintf("%d", int(mr.AvgOver("ping", 1))))
			ui.LabelRow("Prediction Hit", fmt.Sprintf("%d", int(mr.RatePerSec("prediction_hit", metricRange))))
			ui.LabelRow("Prediction Miss", fmt.Sprintf("%d", int(mr.RatePerSec("prediction_miss", metricRange))))

			ui.LabelRow("Triangle Draw Count", formatNumber(runtimeConfig.TriangleDrawCount))
			ui.LabelRow("Draw Count", formatNumber(runtimeConfig.DrawCount))
			ui.LabelRow("gl.GenBuffers() count", fmt.Sprintf("%0.f", mr.RatePerSec("gen_buffers", metricRange)))
		})
	}

	// rendering metrics tracked from gpu
	pairs, total := metricPairsByPrefix(mr, "render_gpu_")

	if imgui.CollapsingHeaderTreeNodeFlagsV(fmt.Sprintf("Rendering - GPU (%.2f)###gpu_rendering_header", total), imgui.TreeNodeFlagsNone) {
		ui.Table("", func() {
			for _, pair := range pairs {
				ui.LabelRow(caser.String(pair.name), fmt.Sprintf("%.2f", pair.value))
			}
		})
	}

	// rendering metrics tracked from cpu
	pairs, total = metricPairsByPrefix(mr, "render_cpu_")

	if imgui.CollapsingHeaderTreeNodeFlagsV(fmt.Sprintf("Rendering - CPU (%.2f)###cpu_rendering_header", total), imgui.TreeNodeFlagsNone) {
		ui.Table("", func() {
			for _, pair := range pairs {
				ui.LabelRow(caser.String(pair.name), fmt.Sprintf("%.2f", pair.value))
			}
		})
	}

	if imgui.CollapsingHeaderTreeNodeFlagsV("Server Stats", imgui.TreeNodeFlagsNone) {
		ui.Table("", func() {
			stats := app.GetServerStats()
			for _, stat := range stats.Data {
				ui.LabelRow(stat.Name, stat.Value)
			}
		})
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
