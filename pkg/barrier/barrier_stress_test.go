package barrier

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestBarrierConcurrentWait stresses the Barrier by repeatedly crossing it with
// many goroutines to ensure no deadlocks or data races occur when executed with
// the `-race` flag.
func TestBarrierConcurrentWait(t *testing.T) {
	const (
		participants = 20
		iterations   = 50
	)

	barrier := NewBarrier(participants)
	defer barrier.Drop()

	var wg sync.WaitGroup
	wg.Add(participants)

	var broken atomic.Bool

	for i := 0; i < participants; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if !barrier.Wait() {
					broken.Store(true)
					return
				}
			}
		}()
	}

	wg.Wait()

	if broken.Load() {
		t.Fatalf("barrier broken during concurrent wait")
	}

	if barrier.RefCount() != 1 {
		t.Fatalf("unexpected ref count after test: want 1, got %d", barrier.RefCount())
	}
}
