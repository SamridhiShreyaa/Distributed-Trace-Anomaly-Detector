package ingest

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Peter/trace-detector/internal/trace"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

func stringAttr(key, value string) *commonpb.KeyValue {
	return &commonpb.KeyValue{
		Key:   key,
		Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: value}},
	}
}

func TestExtractServiceName_FindsAmongMultipleAttributes(t *testing.T) {
	resource := &resourcepb.Resource{
		Attributes: []*commonpb.KeyValue{
			stringAttr("host.name", "ip-10-0-0-1"),
			stringAttr("service.name", "cart-service"),
			stringAttr("deployment.environment", "prod"),
		},
	}

	got := extractServiceName(resource)
	if got != "cart-service" {
		t.Errorf("extractServiceName() = %q, want %q", got, "cart-service")
	}
}

func TestExtractServiceName_MissingReturnsUnknown(t *testing.T) {
	resource := &resourcepb.Resource{
		Attributes: []*commonpb.KeyValue{
			stringAttr("host.name", "ip-10-0-0-1"),
		},
	}

	got := extractServiceName(resource)
	if got != "unknown" {
		t.Errorf("extractServiceName() = %q, want %q", got, "unknown")
	}
}

func TestConvertSpan_AllFieldsPopulated(t *testing.T) {
	traceID := bytes.Repeat([]byte{0xAB}, 16)
	spanID := bytes.Repeat([]byte{0xCD}, 8)
	parentID := bytes.Repeat([]byte{0xEF}, 8)

	rawSpan := &tracepb.Span{
		TraceId:           traceID,
		SpanId:            spanID,
		ParentSpanId:      parentID,
		Name:              "GET /checkout",
		StartTimeUnixNano: 1000,
		EndTimeUnixNano:   2000,
		Status:            &tracepb.Status{Code: tracepb.Status_STATUS_CODE_OK},
	}

	span := convertSpan(rawSpan, "auth-service")

	if len(span.TraceID) != 32 {
		t.Errorf("TraceID length = %d, want 32 (16 bytes hex-encoded)", len(span.TraceID))
	}
	if len(span.SpanID) != 16 {
		t.Errorf("SpanID length = %d, want 16 (8 bytes hex-encoded)", len(span.SpanID))
	}
	if span.ParentID != "efefefefefefefef" {
		t.Errorf("ParentID = %q, want %q", span.ParentID, "efefefefefefefef")
	}
	if span.Service != "auth-service" {
		t.Errorf("Service = %q, want %q", span.Service, "auth-service")
	}
	if span.Operation != "GET /checkout" {
		t.Errorf("Operation = %q, want %q", span.Operation, "GET /checkout")
	}
	if span.StartTime.UnixNano() != 1000 {
		t.Errorf("StartTime = %v, want UnixNano 1000", span.StartTime)
	}
	if span.EndTime.UnixNano() != 2000 {
		t.Errorf("EndTime = %v, want UnixNano 2000", span.EndTime)
	}
	if span.Status != "STATUS_CODE_OK" {
		t.Errorf("Status = %q, want %q", span.Status, "STATUS_CODE_OK")
	}
	if span.Attributes == nil || len(span.Attributes) != 0 {
		t.Errorf("Attributes = %v, want empty non-nil map", span.Attributes)
	}
}

func TestConvertSpan_EmptyParentSpanIdProducesEmptyParentID(t *testing.T) {
	rawSpan := &tracepb.Span{
		TraceId:      bytes.Repeat([]byte{0x01}, 16),
		SpanId:       bytes.Repeat([]byte{0x02}, 8),
		ParentSpanId: []byte{},
		Name:         "root",
	}

	span := convertSpan(rawSpan, "auth-service")

	if span.ParentID != "" {
		t.Errorf("ParentID = %q, want empty string", span.ParentID)
	}
}

func TestNewTraceHandler_ValidRequestIsRetrievableFromCollector(t *testing.T) {
	traceID := bytes.Repeat([]byte{0x11}, 16)
	rootSpanID := bytes.Repeat([]byte{0x22}, 8)
	childSpanID := bytes.Repeat([]byte{0x33}, 8)

	req := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{stringAttr("service.name", "cart-service")},
				},
				ScopeSpans: []*tracepb.ScopeSpans{
					{
						Spans: []*tracepb.Span{
							{
								TraceId:           traceID,
								SpanId:            rootSpanID,
								Name:              "root",
								StartTimeUnixNano: 1000,
								EndTimeUnixNano:   5000,
							},
							{
								TraceId:           traceID,
								SpanId:            childSpanID,
								ParentSpanId:      rootSpanID,
								Name:              "child",
								StartTimeUnixNano: 2000,
								EndTimeUnixNano:   3000,
							},
						},
					},
				},
			},
		},
	}

	body, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal test request: %v", err)
	}

	collector := trace.NewTraceCollector()
	handler := NewTraceHandler(collector)

	httpReq := httptest.NewRequest(http.MethodPost, "/v1/traces", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler(rec, httpReq)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	wantTraceID := hex.EncodeToString(traceID)
	got, err := collector.GetTrace(wantTraceID)
	if err != nil {
		t.Fatalf("collector.GetTrace(%q) returned error: %v", wantTraceID, err)
	}

	if got.TraceID != wantTraceID {
		t.Errorf("TraceID = %q, want %q", got.TraceID, wantTraceID)
	}
	if len(got.Spans) != 2 {
		t.Errorf("len(Spans) = %d, want 2", len(got.Spans))
	}
}
