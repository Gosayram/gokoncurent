// Package condvar provides conditional variables for goroutine coordination
// with atomic reference counting, similar to sync.Cond but with Arc semantics.
package condvar

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// CondVar represents a conditional variable with atomic reference counting.
// It provides a way for goroutines to wait for a condition to become true
// while maintaining thread-safe reference counting.
type CondVar struct {
	mu       sync.Mutex
	cond     *sync.Cond
	refCount atomic.Int64
}

// NewCondVar creates a new conditional variable with initial reference count of 1.
func NewCondVar() *CondVar {
	cv := &CondVar{}
	cv.cond = sync.NewCond(&cv.mu)
	cv.refCount.Store(1)
	return cv
}

// Clone creates a new reference to the conditional variable, incrementing the reference count.
func (cv *CondVar) Clone() *CondVar {
	cv.refCount.Add(1)
	return cv
}

// Drop decrements the reference count. When the count reaches zero,
// the conditional variable is considered "dropped" and should not be used further.
func (cv *CondVar) Drop() {
	for {
		current := cv.refCount.Load()
		if current <= 0 {
			return // Already dropped or invalid
		}
		if cv.refCount.CompareAndSwap(current, current-1) {
			if current-1 == 0 {
				// Wake up all waiting goroutines when dropping the last reference
				cv.mu.Lock()
				cv.cond.Broadcast()
				cv.mu.Unlock()
			}
			return
		}
	}
}

// RefCount returns the current reference count.
func (cv *CondVar) RefCount() int64 {
	return cv.refCount.Load()
}

// Wait waits for the condition to be signaled. It atomically unlocks the mutex
// and suspends execution of the calling goroutine until the condition is signaled.
func (cv *CondVar) Wait() {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.cond.Wait()
}

// WaitWithContext waits for the condition to be signaled or context cancellation.
// Returns true if the condition was signaled, false if context was canceled.
func (cv *CondVar) WaitWithContext(ctx context.Context) bool {
	// Use a buffered channel to avoid goroutine leak
	done := make(chan bool, 1)

	go func() {
		cv.mu.Lock()
		defer cv.mu.Unlock()
		cv.cond.Wait()
		select {
		case done <- true:
		default:
		}
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		// Wake up the waiting goroutine
		cv.mu.Lock()
		cv.cond.Signal()
		cv.mu.Unlock()
		return false
	}
}

// WaitWithTimeout waits for the condition to be signaled with a timeout.
// Returns true if the condition was signaled, false if timeout occurred.
func (cv *CondVar) WaitWithTimeout(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cv.WaitWithContext(ctx)
}

// Signal wakes up one goroutine waiting on the condition.
func (cv *CondVar) Signal() {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.cond.Signal()
}

// Broadcast wakes up all goroutines waiting on the condition.
func (cv *CondVar) Broadcast() {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.cond.Broadcast()
}

// Lock locks the underlying mutex.
func (cv *CondVar) Lock() {
	cv.mu.Lock()
}

// Unlock unlocks the underlying mutex.
func (cv *CondVar) Unlock() {
	cv.mu.Unlock()
}

// String returns a string representation of the conditional variable.
func (cv *CondVar) String() string {
	return fmt.Sprintf("CondVar{refCount: %d}", cv.refCount.Load())
}

// Notify is a convenience function that creates a new CondVar and returns it
// along with a function to signal it. This is useful for simple notification patterns.
func Notify() (cv *CondVar, signal func()) {
	cv = NewCondVar()
	signal = func() {
		cv.Signal()
	}
	return cv, signal
}

// NotifyBroadcast is a convenience function that creates a new CondVar and returns it
// along with a function to broadcast to all waiting goroutines.
func NotifyBroadcast() (cv *CondVar, broadcast func()) {
	cv = NewCondVar()
	broadcast = func() {
		cv.Broadcast()
	}
	return cv, broadcast
}
