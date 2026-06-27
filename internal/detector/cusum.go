package detector

import (
	"math"

	"github.com/Peter/trace-detector/internal/baseline"
)

// CUSUM is a one-sided cumulative-sum detector layered on top of a latency
// Baseline. It accumulates the amount by which recent z-scores exceed a slack
// allowance K, and raises an alert once that running sum crosses the decision
// threshold H. This catches sustained drifts that any single observation would
// not flag as anomalous on its own.
type CUSUM struct {
	K        float64 // slack parameter — how much z-score is tolerated per sample
	H        float64 // decision threshold — alert fires when S exceeds this
	S        float64 // current cumulative score
	Baseline *baseline.Baseline
}

// NewCUSUM creates a CUSUM with the given slack (k) and threshold (h), backed
// by a fresh Baseline retaining `capacity` samples.
func NewCUSUM(k, h float64, capacity int) *CUSUM {
	return &CUSUM{
		K:        k,
		H:        h,
		Baseline: baseline.NewBaseline(capacity),
	}
}

// Add records latencyMs in the baseline and updates the cumulative score. It
// returns true when the score crosses the decision threshold H (alert).
//
// While the baseline is still filling, Add records the sample but reports no
// alert, since z-scores are not yet trustworthy.
func (c *CUSUM) Add(latencyMs float64) bool {
	c.Baseline.Add(latencyMs)
	if !c.Baseline.IsMature() {
		return false
	}

	z := c.Baseline.ZScore(latencyMs)
	c.S = math.Max(0, c.S+(z-c.K))
	return c.S > c.H
}

// Reset clears the cumulative score back to 0.
func (c *CUSUM) Reset() {
	c.S = 0
}

// Score returns the current cumulative score.
func (c *CUSUM) Score() float64 {
	return c.S
}
