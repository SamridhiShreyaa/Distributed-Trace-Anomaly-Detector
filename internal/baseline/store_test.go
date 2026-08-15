package baseline

import (
	"sync"
	"testing"
)

func TestGetOrCreate_SameKeyReturnsSamePointer(t *testing.T) {
	bs := NewBaselineStore(10)

	a := bs.GetOrCreate("cart-service", "checkout")
	b := bs.GetOrCreate("cart-service", "checkout")

	if a != b {
		t.Errorf("GetOrCreate returned different pointers for the same key: %p != %p", a, b)
	}
}

func TestGetOrCreate_DifferentKeysReturnDifferentPointers(t *testing.T) {
	bs := NewBaselineStore(10)

	a := bs.GetOrCreate("cart-service", "checkout")
	b := bs.GetOrCreate("payment-service", "charge")

	if a == b {
		t.Errorf("GetOrCreate returned the same pointer for different keys: %p == %p", a, b)
	}
}

func TestGetOrCreate_NoStringConcatenationCollision(t *testing.T) {
	bs := NewBaselineStore(10)

	// "a.b" + "." + "c" == "a" + "." + "b.c" == "a.b.c" if keyed by
	// concatenated string. The struct key must keep these distinct.
	a := bs.GetOrCreate("a.b", "c")
	b := bs.GetOrCreate("a", "b.c")

	if a == b {
		t.Errorf("GetOrCreate collided on {%q,%q} vs {%q,%q}: got same pointer %p", "a.b", "c", "a", "b.c", a)
	}

	// each key must still be stable on repeat lookup
	if bs.GetOrCreate("a.b", "c") != a {
		t.Errorf("GetOrCreate(%q, %q) not stable across calls", "a.b", "c")
	}
	if bs.GetOrCreate("a", "b.c") != b {
		t.Errorf("GetOrCreate(%q, %q) not stable across calls", "a", "b.c")
	}
}

func TestGetOrCreate_ConcurrentAccessIsRaceFree(t *testing.T) {
	bs := NewBaselineStore(10)

	const goroutines = 50
	const keysPerGoroutine = 10

	var wg sync.WaitGroup
	results := make([][]*Baseline, goroutines)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			local := make([]*Baseline, keysPerGoroutine)
			for k := 0; k < keysPerGoroutine; k++ {
				local[k] = bs.GetOrCreate("service", "op")
			}
			results[g] = local
		}(g)
	}
	wg.Wait()

	want := results[0][0]
	for g := 0; g < goroutines; g++ {
		for k := 0; k < keysPerGoroutine; k++ {
			if results[g][k] != want {
				t.Fatalf("goroutine %d call %d returned a different pointer under concurrent access", g, k)
			}
		}
	}
}
