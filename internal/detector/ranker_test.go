package detector

import (
	"testing"
	"time"

	"github.com/Peter/trace-detector/internal/baseline"
	"github.com/Peter/trace-detector/internal/trace"
)

func ms(n int) time.Time {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return base.Add(time.Duration(n) * time.Millisecond)
}

// matureBaseline returns a Baseline seeded with `n` copies of latencyMs so that
// its mean (in ms) is latencyMs and it reports as mature.
func matureBaseline(latencyMs float64, n int) *baseline.Baseline {
	b := baseline.NewBaseline(n)
	for i := 0; i < n; i++ {
		b.Add(latencyMs)
	}
	return b
}

// buildABCDTrace constructs the canonical A/B/C/D trace:
//
//	A (root, 0→330)
//	├── B (0→320)  ← longest child, on critical path
//	│   └── D (0→300, leaf)
//	└── C (0→100)
//
// The critical path follows the longest-duration child at each level: A→B→D.
func buildABCDTrace() (a, b, c, d *trace.SpanNode) {
	d = &trace.SpanNode{
		Span: &trace.Span{
			SpanID:    "D",
			Service:   "svcD",
			Operation: "opD",
			StartTime: ms(0),
			EndTime:   ms(300),
		},
	}
	b = &trace.SpanNode{
		Span: &trace.Span{
			SpanID:    "B",
			Service:   "svcB",
			Operation: "opB",
			StartTime: ms(0),
			EndTime:   ms(320),
		},
		Children: []*trace.SpanNode{d},
	}
	d.Parent = b

	c = &trace.SpanNode{
		Span: &trace.Span{
			SpanID:    "C",
			Service:   "svcC",
			Operation: "opC",
			StartTime: ms(0),
			EndTime:   ms(100),
		},
	}
	a = &trace.SpanNode{
		Span: &trace.Span{
			SpanID:    "A",
			Service:   "svcA",
			Operation: "opA",
			StartTime: ms(0),
			EndTime:   ms(330),
		},
		Children: []*trace.SpanNode{b, c},
	}
	b.Parent = a
	c.Parent = a

	return a, b, c, d
}

func TestCriticalPath_ABCD(t *testing.T) {
	a, b, _, d := buildABCDTrace()

	path := CriticalPath(a)

	want := []*trace.SpanNode{a, b, d}
	if len(path) != len(want) {
		t.Fatalf("path length = %d, want %d", len(path), len(want))
	}
	for i := range want {
		if path[i] != want[i] {
			t.Errorf("path[%d] = %s, want %s", i, path[i].Span.SpanID, want[i].Span.SpanID)
		}
	}
}

func TestCriticalPath_SingleNode(t *testing.T) {
	node := &trace.SpanNode{
		Span: &trace.Span{
			SpanID:    "solo",
			StartTime: ms(0),
			EndTime:   ms(50),
		},
	}

	path := CriticalPath(node)

	if len(path) != 1 || path[0] != node {
		t.Fatalf("path = %v, want [solo]", path)
	}
}

func TestRankSpans_HighestDelta(t *testing.T) {
	_, b, _, d := buildABCDTrace()
	// D is a leaf → self-time 300ms; B → 320 - 300 = 20ms self-time.
	path := []*trace.SpanNode{b, d}

	baselines := map[string]*baseline.Baseline{
		"svcB.opB": matureBaseline(15, 50), // B drifts from ~15ms baseline → small delta
		"svcD.opD": matureBaseline(50, 50), // D drifts from ~50ms baseline → large delta
	}

	ranked := RankSpans(path, baselines)

	if len(ranked) != 2 {
		t.Fatalf("ranked length = %d, want 2", len(ranked))
	}
	if ranked[0].Node != d {
		t.Errorf("highest-delta node = %s, want D", ranked[0].Node.Span.SpanID)
	}
	if ranked[0].Delta <= ranked[1].Delta {
		t.Errorf("expected D delta (%v) > B delta (%v)", ranked[0].Delta, ranked[1].Delta)
	}

	// sanity-check the computed values
	if ranked[0].SelfTime != 300*time.Millisecond {
		t.Errorf("D self-time = %v, want 300ms", ranked[0].SelfTime)
	}
	if ranked[0].BaselineMean != 50*time.Millisecond {
		t.Errorf("D baseline mean = %v, want 50ms", ranked[0].BaselineMean)
	}
	if ranked[0].Delta != 250*time.Millisecond {
		t.Errorf("D delta = %v, want 250ms", ranked[0].Delta)
	}
}

func TestRankSpans_SkipsMissingBaseline(t *testing.T) {
	_, b, _, d := buildABCDTrace()
	path := []*trace.SpanNode{b, d}

	// only D has a baseline; B should be skipped
	baselines := map[string]*baseline.Baseline{
		"svcD.opD": matureBaseline(50, 50),
	}

	ranked := RankSpans(path, baselines)

	if len(ranked) != 1 {
		t.Fatalf("ranked length = %d, want 1 (B skipped)", len(ranked))
	}
	if ranked[0].Node != d {
		t.Errorf("ranked node = %s, want D", ranked[0].Node.Span.SpanID)
	}
}
