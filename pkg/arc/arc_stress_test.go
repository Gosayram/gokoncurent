package arc

import (
	"runtime"
	"sync"
	"testing"
)

// TestArcConcurrentCloneDrop performs a high-concurrency stress test on the Arc[T]
// implementation. It launches many goroutines that repeatedly clone and drop
// references to the same underlying data to ensure there are no data races or
// panics when run with the `-race` flag.
func TestArcConcurrentCloneDrop(t *testing.T) {
	const (
		goroutines         = 100  // number of concurrent workers
		clonesPerGoroutine = 1000 // clones (and subsequent drops) per worker
	)

	base := NewArc(42)
	defer base.Drop()

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			// Each goroutine performs a tight loop of Clone / Drop operations.
			for j := 0; j < clonesPerGoroutine; j++ {
				clone := base.Clone()
				if clone == nil {
					t.Errorf("unexpected nil clone at iteration %d", j)
					return
				}
				// Ensure the cloned value is as expected.
				if v := *clone.Get(); v != 42 {
					t.Errorf("unexpected value: want 42, got %d", v)
					return
				}
				clone.Drop()
			}
		}()
	}

	// Periodically yield the processor to increase scheduling variability.
	runtime.Gosched()

	wg.Wait()

	// After all clones have been dropped, the base reference count should be 1.
	if got := base.RefCount(); got != 1 {
		t.Fatalf("unexpected refcount after stress test: want 1, got %d", got)
	}
}
