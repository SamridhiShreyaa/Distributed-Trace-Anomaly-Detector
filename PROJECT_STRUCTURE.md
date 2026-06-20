# Distributed Trace Anomaly Detector — Project Structure

A daemon that ingests OpenTelemetry traces from a microservice cluster, builds
statistical baselines of normal latency per span, and automatically identifies
the root-cause span when a request slows down — not just that something,
somewhere, got slow.

This document describes the full repository layout, what each piece is
responsible for, and the order in which things get built.

---

## Repository layout

```
trace-detector/
├── README.md
├── go.mod
├── go.sum
├── Makefile
├── docker-compose.yml
├── .gitignore
├── .github/
│   └── workflows/
│       └── ci.yml
│
├── cmd/
│   └── detector/
│       └── main.go
│
├── internal/
│   ├── trace/
│   │   ├── span.go
│   │   ├── tree.go
│   │   ├── tree_test.go
│   │   ├── collector.go
│   │   └── collector_test.go
│   │
│   ├── selftime/
│   │   ├── selftime.go
│   │   ├── selftime_test.go
│   │   ├── criticalpath.go
│   │   └── criticalpath_test.go
│   │
│   ├── baseline/
│   │   ├── model.go
│   │   ├── lognormal.go
│   │   ├── huber.go
│   │   └── baseline_test.go
│   │
│   ├── detector/
│   │   ├── cusum.go
│   │   ├── cusum_test.go
│   │   ├── ranker.go
│   │   └── ranker_test.go
│   │
│   ├── ingest/
│   │   ├── otlp_receiver.go
│   │   └── otlp_receiver_test.go
│   │
│   ├── storage/
│   │   ├── clickhouse.go
│   │   ├── sqlite.go
│   │   └── store.go
│   │
│   ├── alert/
│   │   ├── publisher.go
│   │   └── slack.go
│   │
│   └── api/
│       ├── server.go
│       ├── handlers.go
│       └── proto/
│           ├── detector.proto
│           └── detector.pb.go
│
├── demo/
│   ├── auth/
│   │   └── main.go
│   ├── cart/
│   │   └── main.go
│   ├── inventory/
│   │   └── main.go
│   ├── payment/
│   │   └── main.go
│   └── load/
│       └── inject_regression.go
│
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile.detector
│   │   └── Dockerfile.demo-service
│   ├── otel-collector-config.yaml
│   └── grafana/
│       ├── dashboards/
│       │   └── service-latency.json
│       └── datasources.yaml
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── SELF_TIME.md
│   └── DEMO.md
│
└── scripts/
    ├── seed-demo-data.sh
    └── run-local.sh
```

---

## What goes where, and why

### `cmd/detector/`
The single entrypoint binary. `main.go` wires everything together — reads
config, starts the OTLP receiver, starts the API server, connects storage —
and nothing else. No business logic lives here. This is the file that
compiles into the actual deployable binary.

### `internal/trace/`
**Status: built.** This is everything we've written so far together —
`Span`, `SpanNode`, `Trace`, `BuildTrace`, and `TraceCollector`. It owns the
problem of turning a flat, out-of-order stream of spans into a correctly
assembled tree. Nothing above this layer should ever see a raw `Span` slice —
they only ever see a finished `*Trace`.

`internal/` is a Go convention: any package under `internal/` can only be
imported by code inside this repo. It's how Go enforces "this is not a public
library, don't depend on it from outside."

### `internal/selftime/`
**Status: next up.** This is where the self-time and critical-path logic we
discussed lives — walking the tree, handling overlapping parallel children,
computing `self_time = duration − merged child intervals`, and identifying
the critical path through the trace.

### `internal/baseline/`
The statistical model of "what does normal look like" for a given
`(service, operation)` pair. Fits a log-normal distribution per span type
over a sliding window, using a Huber estimator so a handful of outliers don't
wreck the baseline.

### `internal/detector/`
Two responsibilities:
- `cusum.go` — the CUSUM change-point detector that watches each span-pair's
  latency stream and fires when cumulative deviation crosses a threshold.
- `ranker.go` — the causal ranker. Given a regressed trace, walks the
  critical path and ranks spans by self-time delta to produce the root-cause
  hypothesis.

### `internal/ingest/`
Receives OTLP (OpenTelemetry Protocol) traces over gRPC from instrumented
services, converts them into our internal `Span` type, and feeds them into
the `TraceCollector`.

### `internal/storage/`
An interface (`store.go`) with two implementations:
- `sqlite.go` — for local development, zero setup required.
- `clickhouse.go` — for the real deployment, since it's built for exactly
  this kind of high-volume columnar time-series data.

Keeping both behind one interface means the rest of the codebase never knows
or cares which one is active.

### `internal/alert/`
Takes a finished root-cause hypothesis and publishes it somewhere a human
will see it — a Slack webhook to start, with room to add others later
(PagerDuty, email, generic webhook) without touching the detection logic.

### `internal/api/`
The gRPC + REST surface (`grpc-gateway` translates one into the other). This
is how a dashboard or a curious engineer queries the daemon — streaming
recent anomalies, inspecting a baseline, or asking for root-cause analysis on
a specific `trace_id` on demand.

### `demo/`
This is what makes the project deployable and demoable, not just a library.
Four intentionally tiny Go HTTP services (`auth → cart → inventory →
payment`) that call each other in a chain and emit real OpenTelemetry
traces. `demo/load/inject_regression.go` is the script that artificially
slows one of them down on command — the thing a recruiter actually clicks to
see the detector work.

### `deploy/`
Everything needed to actually run this somewhere. Dockerfiles, the OTel
Collector config (tells it where to forward traces), and a pre-built Grafana
dashboard so the moment the stack comes up, there's something to look at.

### `docs/`
Three docs, each with a clear job:
- `ARCHITECTURE.md` — the system diagram and component responsibilities (a
  longer version of this file).
- `SELF_TIME.md` — a written explanation of self-time and critical-path
  attribution, with worked examples like the A/B/C/D trace we did by hand.
  This is the document that proves you understand the hardest part of the
  project, even before anyone reads your code.
- `DEMO.md` — exact steps to bring the stack up locally and trigger the demo
  regression.

### `Makefile`
The single source of truth for common commands, so nobody — including future
you — has to remember exact flags:

```makefile
test:        go test ./... -v -race
build:       go build -o bin/detector ./cmd/detector
demo-up:     docker compose up -d
demo-down:   docker compose down
inject-regression: go run demo/load/inject_regression.go --service=cart --latency=300ms --duration=2m
```

### `.github/workflows/ci.yml`
Runs `go test ./... -race` and `go build ./...` on every push. This is a
small thing that signals a lot — a green CI badge on your README tells a
recruiter "this person ships tested code" before they've read a single line.

---

## Build order

This mirrors the order we've actually been working in, and the order that
makes sense dependency-wise — each layer only depends on the ones before it.

1. `internal/trace/` — span, tree, collector *(done)*
2. `internal/selftime/` — self-time + critical path *(next)*
3. `internal/baseline/` — statistical baseline model
4. `internal/detector/` — CUSUM detector + causal ranker
5. `internal/ingest/` — real OTLP receiver (swap in for hand-built test spans)
6. `internal/storage/` — SQLite first, ClickHouse later
7. `demo/` — the four fake services + regression injector
8. `internal/alert/` + `internal/api/` — make results visible and queryable
9. `deploy/` — Docker Compose, Grafana dashboard, OTel Collector config
10. Deploy to Oracle Cloud free tier, write `docs/DEMO.md`, record a short
    demo video for the README

---

## Why this shape

A few deliberate choices worth knowing the reasoning for, since they'll come
up if anyone reviews this with you:

- **`internal/` over a flat package** — signals this is an application, not
  a library others are meant to import. Idiomatic Go for exactly this kind
  of project.
- **One package per concern** (`trace`, `selftime`, `baseline`, `detector`)
  instead of one giant package — each has its own test file sitting right
  next to it, and each can be understood (and reviewed) independently.
- **`demo/` is a first-class citizen, not an afterthought** — most portfolio
  projects fail to be deployable because there's nothing to point a deployed
  instance *at*. Building the fake cluster alongside the detector from the
  start avoids scrambling to bolt one on in week 11.
- **Storage behind an interface** — lets you develop entirely with SQLite
  locally (free, zero setup) and only stand up ClickHouse once you're ready
  to deploy for real, without rewriting detection logic.
