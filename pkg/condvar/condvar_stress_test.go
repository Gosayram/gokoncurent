package condvar

import (
	"sync"
	"testing"
	"time"
)

// TestCondVarConcurrentSignalBroadcast stresses CondVar by having many goroutines wait
// concurrently while others signal/broadcast. The test ensures there are no deadlocks
// and all waiters are eventually released.
func TestCondVarConcurrentSignalBroadcast(t *testing.T) {
	const (
		waiters    = 30
		iterations = 50
	)

	cv := NewCondVar()
	defer cv.Drop()

	var wg sync.WaitGroup
	wg.Add(waiters)

	// Waiter goroutines
	for i := 0; i < waiters; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cv.Wait()
			}
		}()
	}

	// Signaler loop
	for j := 0; j < iterations; j++ {
		cv.Broadcast()
		time.Sleep(500 * time.Microsecond)
	}

	// Ensure any remaining waiters are released
	cv.Broadcast()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for goroutines to finish; possible deadlock")
	}

	if cv.RefCount() != 1 {
		t.Fatalf("unexpected refcount: want 1, got %d", cv.RefCount())
	}
}
