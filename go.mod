module github.com/Peter/trace-detector

go 1.21

require (
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.21.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/proto/otlp v1.0.0
	google.golang.org/grpc v1.60.0
	google.golang.org/protobuf v1.31.0
)
