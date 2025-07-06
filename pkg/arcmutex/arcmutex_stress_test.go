package arcmutex

import (
	"sync"
	"testing"
	"time"
)

// TestArcMutexConcurrentAccess performs a stress test on ArcMutex[T] by spawning
// many goroutines that increment a shared counter using both blocking and
// non-blocking lock acquisition methods. The test ensures that the final value
// matches the expected number of successful increments and that no data races
// occur when executed with the `-race` flag.
func TestArcMutexConcurrentAccess(t *testing.T) {
	const (
		goroutines           = 100
		incrementsPerRoutine = 1000
	)

	counter := NewArcMutex(0)
	defer counter.Drop()

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < incrementsPerRoutine; j++ {
				// Alternate between blocking and timed try-lock.
				if j%2 == 0 {
					// Blocking variant ensures the update eventually happens.
					counter.WithLock(func(val *int) {
						*val += 1
					})
				} else {
					// Timed TryLock with a short timeout.
					ok := counter.TryLock(1*time.Millisecond, func(val *int) {
						*val += 1
					})
					if !ok {
						// If the lock could not be obtained, retry with blocking variant.
						counter.WithLock(func(val *int) {
							*val += 1
						})
					}
				}
			}
		}(i)
	}

	wg.Wait()

	expected := goroutines * incrementsPerRoutine
	got := counter.WithLockResult(func(val *int) interface{} { return *val }).(int)
	if got != expected {
		t.Fatalf("unexpected counter value: want %d, got %d", expected, got)
	}
}
