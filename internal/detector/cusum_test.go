package detector

import "testing"

// jitter holds ~100ms samples with slight variation so the seeded baseline has
// a nonzero standard deviation.
var jitter = []float64{95, 98, 100, 102, 105, 99, 101, 97, 103, 100}

// seed fills the baseline with n ~100ms samples to make it mature.
func seed(c *CUSUM, n int) {
	for i := 0; i < n; i++ {
		c.Add(jitter[i%len(jitter)])
	}
}

func TestAdd_NoAlertBeforeMature(t *testing.T) {
	c := NewCUSUM(0.5, 4.0, 50)
	// fewer samples than capacity → never mature, never alerts
	for i := 0; i < 49; i++ {
		if c.Add(jitter[i%len(jitter)]) {
			t.Fatalf("alert fired at sample %d before baseline matured", i)
		}
	}
	if c.Score() != 0 {
		t.Errorf("Score before maturity = %v, want 0", c.Score())
	}
}

func TestAdd_AccumulatesAndAlerts(t *testing.T) {
	c := NewCUSUM(0.5, 4.0, 50)
	seed(c, 50)

	alerted := false
	for i := 0; i < 5; i++ {
		if c.Add(200) {
			alerted = true
			break
		}
	}
	if !alerted {
		t.Errorf("expected alert after sustained 200ms observations, S = %v", c.Score())
	}
}

func TestReset(t *testing.T) {
	c := NewCUSUM(0.5, 4.0, 50)
	seed(c, 50)
	c.Add(200)
	if c.Score() == 0 {
		t.Fatalf("expected nonzero score after a 200ms spike")
	}

	c.Reset()
	if c.Score() != 0 {
		t.Errorf("Score after Reset = %v, want 0", c.Score())
	}
}

func TestAdd_NormalObservationsNoAlert(t *testing.T) {
	c := NewCUSUM(0.5, 4.0, 50)
	seed(c, 50)

	for i := 0; i < 20; i++ {
		if c.Add(jitter[i%len(jitter)]) {
			t.Fatalf("normal observation %d triggered an alert, S = %v", i, c.Score())
		}
	}
}
