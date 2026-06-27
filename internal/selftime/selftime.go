package selftime

import (
	"sort"
	"time"

	"github.com/Peter/trace-detector/internal/trace"
)

// Interval represents a time window with a start and end.
type Interval struct {
	Start time.Time
	End   time.Time
}

// mergeAndSumOverlaps takes a slice of time intervals, merges any
// overlapping ones, and returns the total covered duration.
//
// Example:
//   A(100,300), B(250,450), C(400,600) → 500ms  (one continuous block)
//   A(100,200), B(500,600)             → 200ms  (two separate blocks, gap ignored)

func mergeAndSumOverlaps(intervals []Interval) time.Duration {
	if len(intervals) == 0 {
		return 0
	}

	// Step 1: sort by start time
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].Start.Before(intervals[j].Start)
	})

	// Step 2: walk through, merging overlapping blocks
	currentStart := intervals[0].Start
	currentEnd := intervals[0].End
	total := time.Duration(0)

	for i := 1; i < len(intervals); i++ {
		if !intervals[i].Start.After(currentEnd) {
			// overlaps — extend the current block if needed
			if intervals[i].End.After(currentEnd) {
				currentEnd = intervals[i].End
			}
		} else {
			// gap — close out the current block, start a new one
			total += currentEnd.Sub(currentStart)
			currentStart = intervals[i].Start
			currentEnd = intervals[i].End
		}
	}

	// don't forget the last block after the loop ends
	total += currentEnd.Sub(currentStart)

	return total
}
// SelfTime computes how long a span spent doing its own work,
// excluding time spent waiting on children.
func SelfTime(node *trace.SpanNode) time.Duration {
	if len(node.Children) == 0 {
		// leaf span — all time is self-time
		return node.Span.Duration()
	}

	// build intervals from children
	intervals := make([]Interval, 0, len(node.Children))
	for _, child := range node.Children {
		intervals = append(intervals, Interval{
			Start: child.Span.StartTime,
			End:   child.Span.EndTime,
		})
	}

	childTime := mergeAndSumOverlaps(intervals)
	selfTime := node.Span.Duration() - childTime

	// clamp to 0 — clock skew can occasionally produce a tiny negative value
	if selfTime < 0 {
		return 0
	}
	return selfTime
}