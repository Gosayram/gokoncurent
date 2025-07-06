package rwarcmutex

import (
	"sync"
	"testing"
)

// TestRWArcMutexConcurrentAccess stresses RWArcMutex with many concurrent readers
// and writers to ensure no races happen when run with `-race`.
func TestRWArcMutexConcurrentAccess(t *testing.T) {
	const (
		readers         = 100
		writers         = 50
		writesPerWriter = 1000
	)

	m := NewRWArcMutex(0)
	defer m.Drop()

	var wg sync.WaitGroup

	// Writer goroutines
	wg.Add(writers)
	for i := 0; i < writers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < writesPerWriter; j++ {
				m.WithLock(func(v *int) {
					*v += 1
				})
			}
		}()
	}

	// Reader goroutines
	wg.Add(readers)
	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			localSum := 0
			for j := 0; j < writesPerWriter; j++ {
				m.WithRLock(func(v *int) {
					localSum += *v // read value to create contention
				})
			}
			_ = localSum // prevent compiler optimizations
		}()
	}

	wg.Wait()

	expected := writers * writesPerWriter
	var final int
	m.WithRLock(func(v *int) {
		final = *v
	})

	if final != expected {
		t.Fatalf("unexpected final value: want %d, got %d", expected, final)
	}
}
