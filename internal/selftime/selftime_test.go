package selftime

import (
	"testing"
	"time"

	"github.com/Peter/trace-detector/internal/trace"
)

func ms(n int) time.Time {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return base.Add(time.Duration(n) * time.Millisecond)
}

func TestMergeAndSumOverlaps_Overlapping(t *testing.T) {
	intervals := []Interval{
		{Start: ms(100), End: ms(300)},
		{Start: ms(250), End: ms(450)},
		{Start: ms(400), End: ms(600)},
	}
	got := mergeAndSumOverlaps(intervals)
	want := 500 * time.Millisecond
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMergeAndSumOverlaps_WithGap(t *testing.T) {
	intervals := []Interval{
		{Start: ms(100), End: ms(200)},
		{Start: ms(500), End: ms(600)},
	}
	got := mergeAndSumOverlaps(intervals)
	want := 200 * time.Millisecond
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSelfTime_LeafSpan(t *testing.T) {
	// D has no children — all its time is self-time
	node := &trace.SpanNode{
		Span: &trace.Span{
			SpanID:    "D",
			StartTime: ms(15),
			EndTime:   ms(290),
		},
		Children: []*trace.SpanNode{},
	}
	got := SelfTime(node)
	want := 275 * time.Millisecond
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSelfTime_HandWorkedExample(t *testing.T) {
	// This is the exact A/B/C/D trace we worked through by hand:
	// B: 10→300, waited for D(15→290) = 275ms
	// B self-time = 290 - 275 = 15ms
	nodeD := &trace.SpanNode{
		Span: &trace.Span{
			SpanID:    "D",
			StartTime: ms(15),
			EndTime:   ms(290),
		},
		Children: []*trace.SpanNode{},
	}
	nodeB := &trace.SpanNode{
		Span: &trace.Span{
			SpanID:    "B",
			StartTime: ms(10),
			EndTime:   ms(300),
		},
		Children: []*trace.SpanNode{nodeD},
	}
	got := SelfTime(nodeB)
	want := 15 * time.Millisecond
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}