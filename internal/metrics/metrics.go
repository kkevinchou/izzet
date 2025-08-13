package metrics

import (
	"sort"
	"strings"
	"time"
)

type Registry struct {
	clock      func() time.Time
	counters   map[string]*WindowCounter
	windowSize int
}

func NewRegistry(windowSize int, clock func() time.Time) *Registry {
	if clock == nil {
		clock = time.Now
	}
	return &Registry{
		clock:      time.Now,
		windowSize: windowSize,
		counters:   make(map[string]*WindowCounter),
	}
}

func (r *Registry) MetricsByPrefix(prefix string) []string {
	var result []string
	for m := range r.counters {
		if strings.HasPrefix(m, prefix) {
			result = append(result, m)
		}
	}
	sort.Strings(result)
	return result
}

func (r *Registry) Inc(name string, v float64) {
	if _, ok := r.counters[name]; !ok {
		r.counters[name] = NewWindowCounter(name, r.windowSize, r.clock)
	}

	r.counters[name].Inc(v)
}

// AvgOver the time window, useful for timings
func (r *Registry) AvgOver(name string, n int) float64 {
	if counter, ok := r.counters[name]; ok {
		return counter.AvgOver(n)
	}
	return 0
}

// SumOver - calculates the sum over the time window
func (r *Registry) SumOver(name string, n int) float64 {
	if counter, ok := r.counters[name]; ok {
		return counter.SumOver(n)
	}
	return 0
}

// RatePerSec - calculates the sum over the time window and divides by the window size
func (r *Registry) RatePerSec(name string, n int) float64 {
	if counter, ok := r.counters[name]; ok {
		return counter.RatePerSec(n)
	}
	return 0
}

type secondBucket struct {
	sec int64   // Unix seconds this bucket refers to
	sum float64 // sum of values observed in that second
	cnt int64   // count of observations (for averages)
}

// WindowCounter keeps a rolling window of S seconds.
type WindowCounter struct {
	name    string
	buckets []secondBucket
	window  int
	clock   func() time.Time
	cursor  int
}

// NewWindowCounter(windowSeconds int, clock func() time.Time) *WindowCounter
func NewWindowCounter(name string, windowSeconds int, clock func() time.Time) *WindowCounter {
	if clock == nil {
		clock = time.Now
	}
	return &WindowCounter{
		name:    name,
		buckets: make([]secondBucket, windowSeconds),
		window:  windowSeconds,
		clock:   clock,
	}
}

func (w *WindowCounter) Inc(v float64) {
	s := w.clock().Unix()
	n := len(w.buckets)

	b := &w.buckets[w.cursor]
	if b.sec != s {
		diff := s - b.sec

		switch {
		case b.sec == 0 || diff < 0:
			// First write or clock moved backwards: just retag current slot.
			w.buckets[w.cursor] = secondBucket{sec: s}

		case diff >= int64(n):
			// Big jump: O(1) fast-forward. We only need the *destination* slot.
			// Old slots are ignored by SumN because sec won't match now-k.
			w.cursor = int((int64(w.cursor) + diff) % int64(n))
			w.buckets[w.cursor] = secondBucket{sec: s}

		default:
			// Small jump (<= window): move cursor in O(1) and clear dest.
			steps := int(diff) // safe: diff < n
			w.cursor = (w.cursor + steps) % n
			w.buckets[w.cursor] = secondBucket{sec: s}
		}
		b = &w.buckets[w.cursor]
	}

	b.sum += v
	b.cnt++
}

// Sum over the last N seconds (<= window)
func (w *WindowCounter) SumOver(n int) float64 {
	if n <= 0 {
		return 0
	}
	if n > len(w.buckets) {
		n = len(w.buckets)
	}

	var total float64

	idx := (w.cursor + w.window - 1) % w.window
	base := w.buckets[idx].sec

	// handles cases where we may not record a metric over the last second
	// with some wiggle room. this allows us to return 0
	now := w.clock().Unix()
	if now-base > 2 {
		return 0
	}

	for k := 0; k < n; k++ {
		b := w.buckets[idx]
		if b.sec >= base-int64(k) {
			total += b.sum
		}
		idx = (idx + w.window - 1) % w.window
	}

	return total
}

func (w *WindowCounter) AvgOver(n int) float64 {
	if n <= 0 {
		return 0
	}
	if n > w.window {
		n = w.window
	}

	var sum float64
	var cnt int64

	idx := (w.cursor + w.window - 1) % w.window
	base := w.buckets[idx].sec

	// handles cases where we may not record a metric over the last second
	// with some wiggle room. this allows us to return 0
	now := w.clock().Unix()
	if now-base > 2 {
		return 0
	}

	for k := 0; k < n; k++ {
		b := w.buckets[idx]
		if b.sec >= base-int64(k) {
			sum += b.sum
			cnt += b.cnt
		}
		idx = (idx + w.window - 1) % w.window
	}

	if cnt == 0 {
		return 0
	}
	return sum / float64(cnt)
}

func (w *WindowCounter) RatePerSec(n int) float64 {
	if n <= 0 {
		return 0
	}
	sum := w.SumOver(n)
	result := sum / float64(n)
	return result
}
