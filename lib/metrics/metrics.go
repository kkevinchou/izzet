package metrics

import (
	"time"
)

const (
	bucketSize = 1000
)

type DataPoint struct {
	recordedAt time.Time
	value      float64
}

type Metric struct {
	cursor          int
	oneSecondCursor int
	oneSecondSum    float64
	data            [bucketSize]DataPoint
}

type MetricsRegistry struct {
	metrics map[string]*Metric
}

func New() *MetricsRegistry {
	return &MetricsRegistry{
		metrics: map[string]*Metric{},
	}
}

func (m *MetricsRegistry) Inc(name string, value float64) {
	if _, ok := m.metrics[name]; !ok {
		m.metrics[name] = &Metric{}
	}

	metric := m.metrics[name]
	dataPoint := DataPoint{recordedAt: time.Now(), value: value}

	metric.data[metric.cursor] = dataPoint
	metric.oneSecondSum += dataPoint.value

	m.advanceMetricCursors(name, time.Second)
	metric.cursor = (metric.cursor + 1) % bucketSize
}

func (m *MetricsRegistry) GetLatest(name string) float64 {
	metric := m.metrics[name]
	if _, ok := m.metrics[name]; !ok {
		return 0
	}

	m.advanceMetricCursors(name, time.Second)
	cursor := (metric.cursor - 1) % bucketSize
	if cursor < 0 {
		cursor += bucketSize
	}
	return metric.data[cursor].value
}

func (m *MetricsRegistry) GetOneSecondSum(name string) float64 {
	metric := m.metrics[name]
	if _, ok := m.metrics[name]; !ok {
		return 0
	}

	m.advanceMetricCursors(name, time.Second)
	return metric.oneSecondSum
}

func (m *MetricsRegistry) GetOneSecondAverage(name string) float64 {
	if _, ok := m.metrics[name]; !ok {
		return 0
	}

	sum := m.GetOneSecondSum(name)
	metric := m.metrics[name]

	lastDataPointIndex := (metric.cursor - 1) % bucketSize
	numDataPoints := lastDataPointIndex - metric.oneSecondCursor
	if numDataPoints < 0 {
		numDataPoints += bucketSize
	}
	numDataPoints++

	if numDataPoints == 0 {
		return 0
	}

	return sum / float64(numDataPoints)
}

// TODO: handle overflowing - there's a chance we can overwrite into the bucket
// if the bucket size is too small which causes weird math calculations
func (m *MetricsRegistry) advanceMetricCursors(name string, duration time.Duration) {
	metric := m.metrics[name]
	if _, ok := m.metrics[name]; !ok {
		return
	}
	start := metric.data[metric.oneSecondCursor]
	for ((time.Since(start.recordedAt)) > duration) && metric.oneSecondCursor != metric.cursor {
		metric.oneSecondSum -= start.value
		metric.oneSecondCursor = (metric.oneSecondCursor + 1) % bucketSize
		start = metric.data[metric.oneSecondCursor]
	}
}
