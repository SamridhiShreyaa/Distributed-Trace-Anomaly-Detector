package baseline

import "math"

// Baseline is a circular-buffer model of recent latency samples. It stores
// log-latencies so that the typically right-skewed latency distribution is
// closer to normal, then exposes mean / standard-deviation / z-score helpers
// for anomaly detection.
type Baseline struct {
	samples  []float64
	head     int
	count    int
	capacity int
}

// NewBaseline creates a Baseline that retains the most recent `capacity`
// samples.
func NewBaseline(capacity int) *Baseline {
	return &Baseline{
		samples:  make([]float64, capacity),
		capacity: capacity,
	}
}

// Add records log(latencyMs) in the circular buffer, overwriting the oldest
// sample once the buffer is full.
func (b *Baseline) Add(latencyMs float64) {
	b.samples[b.head] = math.Log(latencyMs)
	b.head = (b.head + 1) % b.capacity
	if b.count < b.capacity {
		b.count++
	}
}

// Mean returns the mean of the current samples, or 0 if there are none.
func (b *Baseline) Mean() float64 {
	if b.count == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < b.count; i++ {
		sum += b.samples[i]
	}
	return sum / float64(b.count)
}

// StdDev returns the standard deviation of the current samples, or 0 if there
// are fewer than 2 samples.
func (b *Baseline) StdDev() float64 {
	if b.count < 2 {
		return 0
	}

	mean := b.Mean()
	sumSquares := 0.0
	for i := 0; i < b.count; i++ {
		diff := b.samples[i] - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(b.count))
}

// ZScore returns how many standard deviations log(latencyMs) sits from the
// mean. It returns 0 when the standard deviation is 0 (not enough spread to
// judge).
func (b *Baseline) ZScore(latencyMs float64) float64 {
	stdDev := b.StdDev()
	if stdDev == 0 {
		return 0
	}
	return (math.Log(latencyMs) - b.Mean()) / stdDev
}

// IsAnomaly reports whether latencyMs is more than `threshold` standard
// deviations above the mean.
func (b *Baseline) IsAnomaly(latencyMs float64, threshold float64) bool {
	return b.ZScore(latencyMs) > threshold
}

// IsMature reports whether the buffer has filled, at which point the baseline
// has enough history to be trustworthy.
func (b *Baseline) IsMature() bool {
	return b.count >= b.capacity
}
