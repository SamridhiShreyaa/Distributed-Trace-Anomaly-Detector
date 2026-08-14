package ingest

import (
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Peter/trace-detector/internal/trace"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

// NewTraceHandler returns an HTTP handler that accepts an OTLP
// ExportTraceServiceRequest, converts every span it contains into the
// internal trace.Span type, and feeds them into collector.
func NewTraceHandler(collector *trace.TraceCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		req := &coltracepb.ExportTraceServiceRequest{}
		if err := proto.Unmarshal(body, req); err != nil {
			log.Printf("ingest: failed to unmarshal OTLP trace request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		spanCount := 0
		for _, resourceSpans := range req.ResourceSpans {
			serviceName := extractServiceName(resourceSpans.Resource)

			for _, scopeSpans := range resourceSpans.ScopeSpans {
				for _, rawSpan := range scopeSpans.Spans {
					span := convertSpan(rawSpan, serviceName)
					collector.AddSpan(span)
					spanCount++
				}
			}
		}

		log.Printf("ingest: received %d span(s)", spanCount)
		w.WriteHeader(http.StatusOK)
	}
}

// extractServiceName reads the "service.name" resource attribute, returning
// "unknown" if it isn't present.
func extractServiceName(resource *resourcepb.Resource) string {
	if resource == nil {
		return "unknown"
	}

	for _, attr := range resource.Attributes {
		if attr.GetKey() == "service.name" {
			return attr.GetValue().GetStringValue()
		}
	}

	return "unknown"
}

// convertSpan translates a raw OTLP span into the internal trace.Span type.
func convertSpan(rawSpan *tracepb.Span, serviceName string) *trace.Span {
	return &trace.Span{
		TraceID:    hex.EncodeToString(rawSpan.TraceId),
		SpanID:     hex.EncodeToString(rawSpan.SpanId),
		ParentID:   hex.EncodeToString(rawSpan.ParentSpanId),
		Service:    serviceName,
		Operation:  rawSpan.Name,
		StartTime:  time.Unix(0, int64(rawSpan.StartTimeUnixNano)),
		EndTime:    time.Unix(0, int64(rawSpan.EndTimeUnixNano)),
		Status:     rawSpan.Status.GetCode().String(),
		Attributes: map[string]interface{}{},
	}
}
