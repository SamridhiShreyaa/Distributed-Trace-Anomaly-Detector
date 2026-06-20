package baseline

import (
	"time"
)

// Model represents a statistical baseline for span latencies
type Model interface {
	// Fit trains the baseline model on observed latencies
	Fit(latencies []time.Duration) error
	
	// IsAnomaly checks if a given latency is anomalous
	IsAnomaly(latency time.Duration) bool
	
	// Score returns an anomaly score for a latency
	Score(latency time.Duration) float64
}

// LogNormalModel fits a log-normal distribution to latencies
type LogNormalModel struct {
	mu    float64 // mean of log values
	sigma float64 // std dev of log values
}

// NewLogNormalModel creates a new log-normal baseline model
func NewLogNormalModel() *LogNormalModel {
	return &LogNormalModel{}
}

// Fit trains the log-normal model using Huber estimation
func (m *LogNormalModel) Fit(latencies []time.Duration) error {
	// TODO: Implement log-normal fitting with Huber estimation
	// This will compute robust mean and std dev of log-transformed latencies
	return nil
}

// IsAnomaly checks if a latency deviates from the baseline
func (m *LogNormalModel) IsAnomaly(latency time.Duration) bool {
	// TODO: Implement anomaly detection
	// This will check if the latency falls outside expected bounds
	return false
}

// Score returns an anomaly score
func (m *LogNormalModel) Score(latency time.Duration) float64 {
	// TODO: Implement scoring
	return 0.0
}
