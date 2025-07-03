// Package rwarcmutex provides a thread-safe reference-counted read-write mutex for shared mutable state.
//
// RWArcMutex[T] allows multiple readers or one writer, with atomic reference counting for safe sharing between goroutines.
//
// Example usage:
//
//	m := rwarcmutex.NewRWArcMutex(42)
//	m.WithRLock(func(v *int) { fmt.Println(*v) })
//	m.WithLock(func(v *int) { *v = 100 })
//	clone := m.Clone()
//	m.Drop()
//	clone.Drop()
package rwarcmutex

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// RWArcMutex provides a reference-counted, thread-safe read-write mutex for shared mutable state of type T.
type RWArcMutex[T any] struct {
	mu     sync.RWMutex
	refcnt atomic.Int64
	value  *T
	closed atomic.Bool
}

// NewRWArcMutex creates a new RWArcMutex with the given initial value.
func NewRWArcMutex[T any](value T) *RWArcMutex[T] {
	m := &RWArcMutex[T]{
		value: &value,
	}
	m.refcnt.Store(1)
	return m
}

// Clone creates a new reference to the same underlying value.
func (m *RWArcMutex[T]) Clone() *RWArcMutex[T] {
	if m == nil || m.closed.Load() {
		return nil
	}
	m.refcnt.Add(1)
	return m
}

// Drop decrements the reference count and cleans up if it reaches zero.
func (m *RWArcMutex[T]) Drop() {
	if m == nil {
		return
	}
	if m.refcnt.Add(-1) == 0 {
		m.closed.Store(true)
		m.value = nil
	}
}

// RefCount returns the current reference count.
func (m *RWArcMutex[T]) RefCount() int64 {
	if m == nil {
		return 0
	}
	return m.refcnt.Load()
}

// WithRLock executes fn with a read lock on the value.
func (m *RWArcMutex[T]) WithRLock(fn func(*T)) {
	if m == nil || m.closed.Load() {
		return
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	fn(m.value)
}

// WithLock executes fn with a write lock on the value.
func (m *RWArcMutex[T]) WithLock(fn func(*T)) {
	if m == nil || m.closed.Load() {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(m.value)
}

// String returns a string representation of the RWArcMutex.
func (m *RWArcMutex[T]) String() string {
	if m == nil {
		return "<nil RWArcMutex>"
	}
	return fmt.Sprintf("RWArcMutex{refcnt=%d, closed=%v}", m.RefCount(), m.closed.Load())
}
