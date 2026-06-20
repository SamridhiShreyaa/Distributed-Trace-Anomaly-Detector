package trace

// TraceCollector accumulates spans from multiple services and builds complete traces
type TraceCollector struct {
	traces map[string][]*Span
}

// NewTraceCollector creates a new trace collector
func NewTraceCollector() *TraceCollector {
	return &TraceCollector{
		traces: make(map[string][]*Span),
	}
}

// AddSpan adds a span to the collector
func (tc *TraceCollector) AddSpan(span *Span) {
	if tc.traces[span.TraceID] == nil {
		tc.traces[span.TraceID] = make([]*Span, 0)
	}
	tc.traces[span.TraceID] = append(tc.traces[span.TraceID], span)
}

// GetTrace retrieves a complete trace by ID
func (tc *TraceCollector) GetTrace(traceID string) (*Trace, error) {
	spans, exists := tc.traces[traceID]
	if !exists {
		return nil, ErrInvalidSpanID
	}
	return BuildTrace(traceID, spans)
}

// Reset clears the collector
func (tc *TraceCollector) Reset() {
	tc.traces = make(map[string][]*Span)
}
