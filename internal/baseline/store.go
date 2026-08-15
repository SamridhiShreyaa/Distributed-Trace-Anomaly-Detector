package baseline

import "sync"

// baselineKey identifies a baseline by the (service, operation) pair it
// tracks. A struct key is used instead of a "service"+"."+"operation" string
// because concatenation collides: "a.b"+"."+"c" and "a"+"."+"b.c" produce the
// same string but are distinct pairs.
type baselineKey struct {
	service   string
	operation string
}

// BaselineStore hands out one Baseline per (service, operation) pair,
// creating it on first use.
type BaselineStore struct {
	mu        sync.Mutex
	baselines map[baselineKey]*Baseline
	capacity  int
}

// NewBaselineStore creates a BaselineStore whose baselines are each
// constructed with the given capacity.
func NewBaselineStore(capacity int) *BaselineStore {
	return &BaselineStore{
		baselines: make(map[baselineKey]*Baseline),
		capacity:  capacity,
	}
}

// GetOrCreate returns the Baseline for (service, operation), creating it if
// this is the first time the pair has been seen. The same pair always
// returns the same *Baseline.
func (bs *BaselineStore) GetOrCreate(service, operation string) *Baseline {
	key := baselineKey{service: service, operation: operation}

	bs.mu.Lock()
	defer bs.mu.Unlock()

	b, exists := bs.baselines[key]
	if !exists {
		b = NewBaseline(bs.capacity)
		bs.baselines[key] = b
	}
	return b
}
