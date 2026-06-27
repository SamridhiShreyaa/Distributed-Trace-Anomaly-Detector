package detector

import (
	"math"
	"sort"
	"time"

	"github.com/Peter/trace-detector/internal/baseline"
	"github.com/Peter/trace-detector/internal/selftime"
	"github.com/Peter/trace-detector/internal/trace"
)

// RankedSpan pairs a span on the critical path with how far its current
// self-time has drifted from its historical baseline.
type RankedSpan struct {
	Node         *trace.SpanNode
	SelfTime     time.Duration
	BaselineMean time.Duration
	Delta        time.Duration
	ZScore       float64
}

// CriticalPath walks from root downward, at each node following the child with
// the longest Duration(), and stops at a leaf. It returns the chain of nodes
// from root to leaf in order.
func CriticalPath(root *trace.SpanNode) []*trace.SpanNode {
	path := []*trace.SpanNode{}

	node := root
	for node != nil {
		path = append(path, node)

		if len(node.Children) == 0 {
			break
		}

		// pick the child with the longest duration
		longest := node.Children[0]
		for _, child := range node.Children[1:] {
			if child.Span.Duration() > longest.Span.Duration() {
				longest = child
			}
		}
		node = longest
	}

	return path
}

// RankSpans scores each node on the path against its service+operation
// baseline, skipping nodes that have no baseline. The result is sorted by
// Delta (how much slower than baseline) in descending order.
func RankSpans(path []*trace.SpanNode, baselines map[string]*baseline.Baseline) []RankedSpan {
	ranked := []RankedSpan{}

	for _, node := range path {
		key := node.Span.Service + "." + node.Span.Operation
		b, exists := baselines[key]
		if !exists {
			continue
		}

		selfTime := selftime.SelfTime(node)

		// baseline stores log(ms), so convert the mean back to a duration
		baselineMean := time.Duration(math.Exp(b.Mean())) * time.Millisecond
		delta := selfTime - baselineMean
		// divide as floats so sub-millisecond self-times aren't truncated away
		selfTimeMs := float64(selfTime) / float64(time.Millisecond)
		zScore := b.ZScore(selfTimeMs)

		ranked = append(ranked, RankedSpan{
			Node:         node,
			SelfTime:     selfTime,
			BaselineMean: baselineMean,
			Delta:        delta,
			ZScore:       zScore,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Delta > ranked[j].Delta
	})

	return ranked
}
