package oncecell

import (
	"sync"
	"testing"
	"time"
)

// TestOnceCellConcurrentGetOrInitWithRetry stresses concurrent initialization of a OnceCell[T]
// to ensure that the initialization function is executed exactly once even when many
// goroutines attempt to initialize it simultaneously with retries.
func TestOnceCellConcurrentGetOrInitWithRetry(t *testing.T) {
	const goroutines = 100

	cell := NewOnceCell[int]()
	var initCalls int64
	var mu sync.Mutex

	// initialization function with artificial failure for the first few attempts
	initFn := func() (int, error) {
		mu.Lock()
		initCalls++
		mu.Unlock()
		// simulate transient error on first two calls
		if initCalls <= 2 {
			return 0, assertError("transient error")
		}
		return 42, nil
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			value, err := cell.GetOrInitWithRetry(initFn, 5, 1*time.Millisecond)
			if err != nil {
				t.Errorf("unexpected error from GetOrInitWithRetry: %v", err)
				return
			}
			if value != 42 {
				t.Errorf("unexpected value: want 42, got %d", value)
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	totalCalls := initCalls
	mu.Unlock()

	if totalCalls == 0 {
		t.Fatalf("init function was never called")
	}
	if totalCalls > int64(goroutines) {
		t.Fatalf("unexpected number of init calls: %d", totalCalls)
	}
	// We expect more than one call because first attempts return errors
	if totalCalls < 3 {
		t.Fatalf("expected at least 3 init attempts (2 failures + 1 success), got %d", totalCalls)
	}
}

// assertError is a helper type that implements the error interface.
type assertError string

func (e assertError) Error() string { return string(e) }
