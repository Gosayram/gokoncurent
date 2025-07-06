// Package barrier provides a synchronization primitive for waiting for multiple goroutines.
// Supports atomic reference counting and state reset.
package barrier

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Barrier implements a synchronization primitive for waiting for N goroutines.
type Barrier struct {
	mu       sync.Mutex
	cond     *sync.Cond
	count    int
	waiting  int
	refCount atomic.Int64
	broken   bool
	gen      int // generation counter to distinguish cycles
}

// NewBarrier creates a new Barrier for n participants.
func NewBarrier(n int) *Barrier {
	if n <= 0 {
		panic("barrier: n must be > 0")
	}
	b := &Barrier{count: n}
	b.cond = sync.NewCond(&b.mu)
	b.refCount.Store(1)
	b.gen = 0
	return b
}

// Clone increments the reference count.
func (b *Barrier) Clone() *Barrier {
	b.refCount.Add(1)
	return b
}

// Drop decrements the reference count and wakes up all waiting goroutines when it reaches zero.
func (b *Barrier) Drop() {
	for {
		current := b.refCount.Load()
		if current <= 0 {
			return
		}
		if b.refCount.CompareAndSwap(current, current-1) {
			if current-1 == 0 {
				b.mu.Lock()
				b.broken = true
				b.cond.Broadcast()
				b.mu.Unlock()
			}
			return
		}
	}
}

// RefCount returns the current reference count.
func (b *Barrier) RefCount() int64 {
	return b.refCount.Load()
}

// Wait blocks the goroutine until all participants call Wait.
// Returns true if the barrier was successfully crossed, false if the barrier was broken.
func (b *Barrier) Wait() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.broken {
		return false
	}

	// Remember current generation.
	myGen := b.gen

	b.waiting++
	if b.waiting == b.count {
		// Last goroutine for this generation.
		b.gen++            // advance generation
		b.waiting = 0      // reset for next cycle
		b.cond.Broadcast() // wake up all waiters
		return true
	}

	for !b.broken && myGen == b.gen {
		b.cond.Wait()
	}
	return !b.broken
}

// Reset resets the barrier (can only be used when no goroutines are waiting).
func (b *Barrier) Reset(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.waiting != 0 {
		panic("barrier: cannot reset while goroutines are waiting")
	}
	b.count = n
	b.broken = false
}

// String returns a string representation of the barrier.
func (b *Barrier) String() string {
	return fmt.Sprintf("Barrier{count=%d, waiting=%d, refCount=%d, broken=%v}",
		b.count, b.waiting, b.refCount.Load(), b.broken)
}
