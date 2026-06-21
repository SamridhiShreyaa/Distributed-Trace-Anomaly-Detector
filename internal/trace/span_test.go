package trace

import (
	"errors"
	"testing"
	"time"
)

func TestSpanDuration(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	span := &Span{
		StartTime: start,
		EndTime:   start.Add(150 * time.Millisecond),
	}

	if got := span.Duration(); got != 150*time.Millisecond {
		t.Errorf("Duration() = %v, want %v", got, 150*time.Millisecond)
	}
}

func TestBuildTraceErrors(t *testing.T) {
	tests := []struct {
		name    string
		spans   []*Span
		wantErr error
	}{
		{
			name:    "empty trace",
			spans:   nil,
			wantErr: ErrEmptyTrace,
		},
		{
			name: "multiple roots",
			spans: []*Span{
				{SpanID: "a", ParentID: ""},
				{SpanID: "b", ParentID: ""},
			},
			wantErr: ErrMultipleRoots,
		},
		{
			name: "orphan span",
			spans: []*Span{
				{SpanID: "root", ParentID: ""},
				{SpanID: "child", ParentID: "missing"},
			},
			wantErr: ErrOrphanSpan,
		},
		{
			name: "no root span (cycle)",
			spans: []*Span{
				{SpanID: "a", ParentID: "b"},
				{SpanID: "b", ParentID: "a"},
			},
			wantErr: ErrNoRootSpan,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := BuildTrace("trace-1", tt.spans)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("BuildTrace() error = %v, want %v", err, tt.wantErr)
			}
			if tr != nil {
				t.Errorf("BuildTrace() returned non-nil trace %v on error", tr)
			}
		})
	}
}

func TestBuildTraceSuccess(t *testing.T) {
	spans := []*Span{
		{SpanID: "root", ParentID: ""},
		{SpanID: "child-1", ParentID: "root"},
		{SpanID: "child-2", ParentID: "root"},
		{SpanID: "grandchild", ParentID: "child-1"},
	}

	tr, err := BuildTrace("trace-1", spans)
	if err != nil {
		t.Fatalf("BuildTrace() unexpected error: %v", err)
	}

	if tr.TraceID != "trace-1" {
		t.Errorf("TraceID = %q, want %q", tr.TraceID, "trace-1")
	}

	if len(tr.Spans) != len(spans) {
		t.Errorf("Spans length = %d, want %d", len(tr.Spans), len(spans))
	}

	if tr.Root == nil {
		t.Fatal("Root is nil, want the root span node")
	}
	if tr.Root.Span.SpanID != "root" {
		t.Errorf("Root span ID = %q, want %q", tr.Root.Span.SpanID, "root")
	}
	if tr.Root.Parent != nil {
		t.Errorf("Root.Parent = %v, want nil", tr.Root.Parent)
	}
	if len(tr.Root.Children) != 2 {
		t.Errorf("Root has %d children, want 2", len(tr.Root.Children))
	}

	// Every child node must point back at its parent.
	for _, child := range tr.Root.Children {
		if child.Parent != tr.Root {
			t.Errorf("child %q parent = %v, want root", child.Span.SpanID, child.Parent)
		}
	}
}
