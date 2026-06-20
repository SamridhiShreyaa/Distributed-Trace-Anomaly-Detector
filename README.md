# Distributed Trace Anomaly Detector

A high-performance daemon that ingests OpenTelemetry traces from a microservice cluster, builds statistical baselines of normal latency per span, and automatically identifies the root-cause span when a request slows down — not just that something, somewhere, got slow.

## Key Features

- **Automatic Root-Cause Analysis**: Identifies the exact span causing latency regressions using CUSUM change-point detection and causal ranking
- **Self-Time Attribution**: Accurately computes per-span self-time handling overlapping parallel children and complex trace trees
- **Statistical Baselines**: Fits log-normal distributions to span latencies using Huber estimators for robust outlier handling
- **OpenTelemetry Native**: Ingests OTLP traces over gRPC from any instrumented service
- **Flexible Storage**: SQLite for development, ClickHouse for production deployments
- **Real-Time Alerting**: Publishes root-cause findings to Slack with extensible alert backends
- **Demo Ready**: Includes fully instrumented demo microservices to test and showcase the detector

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   Instrumented Services                         │
│          (emit OpenTelemetry traces via gRPC)                   │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Detector Daemon                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐  ┌────────────────┐  ┌──────────────────┐ │
│  │ OTLP Receiver   │  │ Trace Builder  │  │   Baseline       │ │
│  │ (gRPC)          │→ │ + Self-Time    │→ │   Model          │ │
│  └─────────────────┘  └────────────────┘  │   (Log-Normal)   │ │
│                                            └────────┬─────────┘ │
│                                                     │            │
│                                            ┌────────▼─────────┐ │
│                                            │  CUSUM Detector  │ │
│                                            │  + Ranker        │ │
│                                            └────────┬─────────┘ │
│                                                     │            │
│  ┌─────────────────┐  ┌──────────────────────────┐ │ │
│  │   Storage       │  │   Alert Publisher        │ │ │
│  │  (SQLite/CH)    │  │   (Slack, etc.)         │ │ │
│  └─────────────────┘  └──────────────────────────┘ │ │
│                                                      │ │
│  ┌──────────────────────────────────────────────────┘ │
│  │                                                     │
│  ▼                                                     │
│  gRPC + REST API                                      │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Make

### Local Development

1. **Build the detector:**
   ```bash
   make build
   ```

2. **Run tests:**
   ```bash
   make test
   ```

3. **Start the demo stack:**
   ```bash
   make demo-up
   ```

4. **Inject a regression:**
   ```bash
   make inject-regression
   ```

5. **Stop the stack:**
   ```bash
   make demo-down
   ```

## Project Structure

```
trace-detector/
├── cmd/detector/              # Main entrypoint binary
├── internal/
│   ├── trace/                 # Span tree assembly and collection
│   ├── selftime/              # Self-time and critical-path computation
│   ├── baseline/              # Statistical baseline model (log-normal)
│   ├── detector/              # CUSUM detector and causal ranker
│   ├── ingest/                # OTLP receiver
│   ├── storage/               # Storage interface (SQLite/ClickHouse)
│   ├── alert/                 # Alert publishing (Slack, etc.)
│   └── api/                   # gRPC + REST API
├── demo/                      # Demo microservices
│   ├── auth/
│   ├── cart/
│   ├── inventory/
│   ├── payment/
│   └── load/                  # Regression injection tool
├── deploy/                    # Deployment configs
│   ├── docker/                # Dockerfiles
│   ├── otel-collector-config.yaml
│   └── grafana/               # Dashboards
├── docs/                      # Architecture and implementation guides
├── scripts/                   # Helper scripts
└── Makefile
```

## Key Concepts

### Self-Time
The time a span actually spent executing, excluding time spent waiting for child spans. This is crucial for identifying the true root cause in complex call trees with overlapping parallel operations.

### Critical Path
The longest sequence of dependent operations through a trace. Root-cause anomalies almost always appear on the critical path.

### CUSUM Detection
Cumulative Sum (CUSUM) control chart monitors each span-pair's latency stream. When cumulative deviation crosses a threshold, an anomaly is detected.

### Causal Ranking
Once an anomaly is detected, the ranker walks the critical path and ranks spans by self-time delta to produce the root-cause hypothesis.

## Documentation

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) — System design and component responsibilities
- [SELF_TIME.md](docs/SELF_TIME.md) — Self-time and critical-path computation with worked examples
- [DEMO.md](docs/DEMO.md) — Step-by-step demo walkthrough

## CI/CD

- GitHub Actions runs `go test ./... -race` and `go build ./...` on every push
- All code must pass tests before merging

## Deployment

The project is designed for deployment to cloud environments (Oracle Cloud free tier, AWS, GCP, etc.) using Docker and Docker Compose.

### Local Deployment

```bash
make demo-up       # Start detector + demo services + OTel Collector + Grafana
```

### Production Deployment

Replace SQLite with ClickHouse and deploy via container orchestration (Kubernetes, Docker Swarm, etc.).

## License

MIT

## Author

Built as a portfolio project to demonstrate:
- Advanced Go application architecture
- Statistical anomaly detection
- Distributed systems instrumentation
- Production-ready code quality
