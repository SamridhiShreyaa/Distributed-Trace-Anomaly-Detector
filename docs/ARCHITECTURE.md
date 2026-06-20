# Architecture

The Distributed Trace Anomaly Detector follows a layered, pipeline architecture designed for high performance and maintainability.

## Component Responsibilities

- **OTLP Receiver**: Ingests OpenTelemetry Protocol traces over gRPC
- **Trace Collector & Builder**: Assembles flat span streams into trace trees
- **Self-Time Calculator**: Computes per-span self-time and identifies critical paths
- **Baseline Model**: Maintains statistical distribution of normal latencies per span type
- **CUSUM Detector**: Monitors span-pair latency streams for change points
- **Causal Ranker**: Ranks spans by self-time delta on critical path to identify root cause
- **Storage**: Persists traces and baselines (SQLite for dev, ClickHouse for prod)
- **Alert Publisher**: Publishes findings to external systems (Slack, etc.)
- **API Server**: Provides gRPC + REST interface for queries
