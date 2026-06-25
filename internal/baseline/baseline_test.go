package baseline

import (
	"math"
	"testing"
)

func TestAdd_FillsBuffer(t *testing.T) {
	b := NewBaseline(3)
	b.Add(100)
	b.Add(100)
	if b.count != 2 {
		t.Errorf("count = %d, want 2", b.count)
	}

	b.Add(100)
	if b.count != 3 {
		t.Errorf("count = %d, want 3", b.count)
	}

	// adding beyond capacity must not grow count past capacity
	b.Add(100)
	if b.count != 3 {
		t.Errorf("count = %d, want 3 (capped at capacity)", b.count)
	}
}

func TestAdd_OverwritesOldest(t *testing.T) {
	b := NewBaseline(2)
	// log(e) == 1, log(e^2) == 2, log(e^3) == 3
	b.Add(math.E)        // slot 0 = 1
	b.Add(math.E * math.E) // slot 1 = 2
	// buffer full with {1, 2}, mean = 1.5
	if got := b.Mean(); math.Abs(got-1.5) > 1e-9 {
		t.Fatalf("Mean = %v, want 1.5", got)
	}

	// this overwrites the oldest (slot 0): buffer becomes {3, 2}, mean = 2.5
	b.Add(math.Exp(3))
	if got := b.Mean(); math.Abs(got-2.5) > 1e-9 {
		t.Errorf("Mean after overwrite = %v, want 2.5", got)
	}
}

func TestMean_KnownValues(t *testing.T) {
	b := NewBaseline(4)
	// logs: 1, 2, 3 → mean = 2
	b.Add(math.E)
	b.Add(math.Exp(2))
	b.Add(math.Exp(3))
	if got := b.Mean(); math.Abs(got-2.0) > 1e-9 {
		t.Errorf("Mean = %v, want 2.0", got)
	}
}

func TestMean_EmptyIsZero(t *testing.T) {
	b := NewBaseline(4)
	if got := b.Mean(); got != 0 {
		t.Errorf("Mean of empty = %v, want 0", got)
	}
}

func TestStdDev_FewerThanTwoIsZero(t *testing.T) {
	b := NewBaseline(4)
	if got := b.StdDev(); got != 0 {
		t.Errorf("StdDev of empty = %v, want 0", got)
	}
	b.Add(100)
	if got := b.StdDev(); got != 0 {
		t.Errorf("StdDev of single sample = %v, want 0", got)
	}
}

func TestZScore_IdentifiesOutlier(t *testing.T) {
	b := NewBaseline(64)
	// ~100ms samples with small jitter so the baseline has nonzero spread
	jitter := []float64{95, 98, 100, 102, 105, 99, 101, 97, 103, 100}
	for i := 0; i < 50; i++ {
		b.Add(jitter[i%len(jitter)])
	}
	z := b.ZScore(2000)
	if z <= 3.0 {
		t.Errorf("ZScore(2000) = %v, want > 3.0", z)
	}
}

func TestZScore_ZeroStdDevIsZero(t *testing.T) {
	b := NewBaseline(4)
	b.Add(100)
	b.Add(100)
	// all samples identical → StdDev 0 → ZScore 0
	if got := b.ZScore(2000); got != 0 {
		t.Errorf("ZScore with zero StdDev = %v, want 0", got)
	}
}

func TestIsAnomaly_SpikeAboveThreshold(t *testing.T) {
	b := NewBaseline(64)
	jitter := []float64{95, 98, 100, 102, 105, 99, 101, 97, 103, 100}
	for i := 0; i < 50; i++ {
		b.Add(jitter[i%len(jitter)])
	}
	if !b.IsAnomaly(2000, 3.0) {
		t.Errorf("IsAnomaly(2000, 3.0) = false, want true")
	}
}

func TestIsMature(t *testing.T) {
	b := NewBaseline(3)
	b.Add(100)
	b.Add(100)
	if b.IsMature() {
		t.Errorf("IsMature before full = true, want false")
	}

	b.Add(100)
	if !b.IsMature() {
		t.Errorf("IsMature after full = false, want true")
	}
}
