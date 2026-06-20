# Self-Time and Critical Path Computation

## Self-Time Definition

Self-time is the time a span actually spent executing, excluding time spent waiting for child operations to complete.

```
Span A: [====0---1---2---3====]  (duration: 4 units)
  Child B: [--0-1--]             (interval: 1-2)
  Child C: [-----1-2--]          (interval: 2-3)

Self-time of A = 4 - merge([1-2, 2-3]) = 4 - [1-3] = 1 unit
```

## Critical Path

The critical path is the longest sequence of dependent operations through a trace. Root-cause anomalies almost always manifest on the critical path.

## Implementation Notes

See code comments in `internal/selftime/` for detailed implementation.
