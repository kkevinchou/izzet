package render

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/globals"
)

const gpuProfilerMetricPrefix = "render_gpu_"

type GPUTimerQuery struct {
	name  string
	query uint32
}

type GPUProfiler struct {
	pending []GPUTimerQuery
}

func NewGPUProfiler() *GPUProfiler {
	return &GPUProfiler{}
}

func (p *GPUProfiler) CollectAvailable() {
	var stillPending []GPUTimerQuery
	for _, timer := range p.pending {
		var available uint32
		gl.GetQueryObjectuiv(timer.query, gl.QUERY_RESULT_AVAILABLE, &available)
		if available != gl.TRUE {
			stillPending = append(stillPending, timer)
			continue
		}

		var elapsedNanoseconds uint64
		gl.GetQueryObjectui64v(timer.query, gl.QUERY_RESULT, &elapsedNanoseconds)
		gl.DeleteQueries(1, &timer.query)

		globals.ClientRegistry().Inc(gpuProfilerMetricPrefix+timer.name, float64(elapsedNanoseconds)/1000000.0)
	}
	p.pending = stillPending
}

func (p *GPUProfiler) Profile(name string, f func()) {
	var query uint32
	gl.GenQueries(1, &query)
	gl.BeginQuery(gl.TIME_ELAPSED, query)

	f()

	gl.EndQuery(gl.TIME_ELAPSED)
	p.pending = append(p.pending, GPUTimerQuery{name: name, query: query})
}
