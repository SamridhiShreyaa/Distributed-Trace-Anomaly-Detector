package trace

import (
	"time"
)

// Span represents a single operation within a distributed trace
type Span struct {
	TraceID    string
	SpanID     string
	ParentID   string
	Service    string
	Operation  string
	StartTime  time.Time
	EndTime    time.Time
	Status     string
	Attributes map[string]interface{}
}

// Duration returns the duration of the span
func (s *Span) Duration() time.Duration {
	return s.EndTime.Sub(s.StartTime)
}

// Trace represents a complete distributed trace with all its spans
type Trace struct {
	TraceID string
	Root    *SpanNode
	Spans   map[string]*Span
}

// SpanNode represents a span in the trace tree
type SpanNode struct {
	Span     *Span
	Parent   *SpanNode
	Children []*SpanNode
}

// BuildTrace assembles a flat slice of spans into a tree structure
func BuildTrace(traceID string, spans []*Span) (*Trace, error) {
	if len(spans) == 0 {
		return nil, ErrEmptyTrace
	}

	spanMap := make(map[string]*Span)
	nodeMap := make(map[string]*SpanNode)

	for _, span := range spans {
		spanMap[span.SpanID] = span
		nodeMap[span.SpanID] = &SpanNode{Span: span}
	}

	var root *SpanNode
	for _, node := range nodeMap {
		if node.Span.ParentID == "" {
			if root != nil {
				return nil, ErrMultipleRoots
			}
			root = node
		} else {
			parent, exists := nodeMap[node.Span.ParentID]
			if !exists {
				return nil, ErrOrphanSpan
			}
			node.Parent = parent
			parent.Children = append(parent.Children, node)
		}
	}

	if root == nil {
		return nil, ErrNoRootSpan
	}

	return &Trace{
		TraceID: traceID,
		Root:    root,
		Spans:   spanMap,
	}, nil
}
