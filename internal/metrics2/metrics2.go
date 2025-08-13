package metrics2

import (
	"time"
)

type Registry struct {
	clock func() time.Time
	// maps: name+labels → instrument
	counters map[string]*WindowCounter // for windowed rates (engine overlay)
}

func NewRegistry(clock func() time.Time) *Registry {
	if clock == nil {
		clock = time.Now
	}
	return &Registry{
		clock:    clock,
		counters: make(map[string]*WindowCounter),
	}
}

func (r *Registry) Inc(name string, v float64) {
	if _, ok := r.counters[name]; !ok {
		r.counters[name] = NewWindowCounter(5, r.clock)
	}

	r.counters[name].Add(v)
}

// AverageValueOver the time window, useful for timings
func (r *Registry) AverageValueOver(name string, n int) float64 {
	if counter, ok := r.counters[name]; ok {
		return counter.AvgValue(n)
	}
	return 0
}

// SumOver - calculates the sum over the time window
func (r *Registry) SumOver(name string, n int) float64 {
	if counter, ok := r.counters[name]; ok {
		return counter.SumN(n)
	}
	return 0
}

// PerSecondRateOver - calculates the sum over the time window and divides by the window size
func (r *Registry) PerSecondRateOver(name string, n int) float64 {
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
	buckets []secondBucket
	window  int // number of seconds tracked
	i       int // cursor
	clock   func() time.Time
}

// NewWindowCounter(windowSeconds int, clock func() time.Time) *WindowCounter
func NewWindowCounter(windowSeconds int, clock func() time.Time) *WindowCounter {
	if clock == nil {
		clock = time.Now
	}
	return &WindowCounter{
		buckets: make([]secondBucket, windowSeconds),
		window:  windowSeconds,
		clock:   clock,
	}
}

func (w *WindowCounter) Add(v float64) {
	s := w.clock().Unix()
	n := len(w.buckets)

	b := &w.buckets[w.i]
	if b.sec != s {
		diff := s - b.sec

		switch {
		case b.sec == 0 || diff < 0:
			// First write or clock moved backwards: just retag current slot.
			w.buckets[w.i] = secondBucket{sec: s}

		case diff >= int64(n):
			// Big jump: O(1) fast-forward. We only need the *destination* slot.
			// Old slots are ignored by SumN because sec won't match now-k.
			w.i = int((int64(w.i) + diff) % int64(n))
			w.buckets[w.i] = secondBucket{sec: s}

		default:
			// Small jump (<= window): move cursor in O(1) and clear dest.
			steps := int(diff) // safe: diff < n
			w.i = (w.i + steps) % n
			w.buckets[w.i] = secondBucket{sec: s}
		}
		b = &w.buckets[w.i]
	}

	b.sum += v
	b.cnt++
}

// Sum over the last N seconds (<= window)
func (w *WindowCounter) SumN(n int) float64 {
	if n <= 0 {
		return 0
	}
	if n > len(w.buckets) {
		n = len(w.buckets)
	}
	total := 0.0
	base := w.clock().Unix() // start from the last *completed* second
	idx := w.i
	for k := 0; k < n; k++ {
		b := w.buckets[idx]
		if b.sec == base-int64(k) {
			total += b.sum
		}
		idx--
		if idx < 0 {
			idx += len(w.buckets)
		}
	}
	return total
}

func (w *WindowCounter) RatePerSec(n int) float64 {
	if n <= 0 {
		return 0
	}
	return w.SumN(n) / float64(n)
}

func (w *WindowCounter) AvgValue(n int) float64 {
	if n <= 0 {
		return 0
	}
	if n > w.window {
		n = w.window
	}

	var sum float64
	var cnt int64
	base := w.clock().Unix() // last fully completed second
	idx := w.i

	for k := 0; k < n; k++ {
		b := w.buckets[idx]
		if b.sec == base-int64(k) {
			sum += b.sum
			cnt += b.cnt
		}
		idx--
		if idx < 0 {
			idx += w.window
		}
	}
	if cnt == 0 {
		return 0
	}
	return sum / float64(cnt)
}
