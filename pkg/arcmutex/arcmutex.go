// Package arcmutex provides thread-safe mutable shared state with controlled access.
// It combines Arc[T] for reference counting with sync.Mutex for safe concurrent
// access to mutable data, inspired by Rust's Arc<Mutex<T>> pattern.
package arcmutex

import (
	"sync"
	"time"

	"github.com/Gosayram/gokoncurent/pkg/arc"
)

// ArcMutex represents a thread-safe mutable reference that can be shared
// between multiple goroutines. It combines Arc[T] for reference counting
// with sync.Mutex for safe concurrent access to mutable data.
//
// Unlike direct mutex usage, ArcMutex[T] provides a safe API that prevents
// deadlocks and ensures proper locking patterns through its WithLock method.
//
// This is inspired by Rust's Arc<Mutex<T>> pattern.
type ArcMutex[T any] struct {
	inner *arc.Arc[mutexData[T]]
}

// mutexData holds the actual data protected by a mutex.
type mutexData[T any] struct {
	mu   sync.Mutex
	data T
}

// NewArcMutex creates a new ArcMutex[T] with the given initial value.
// The value is protected by a mutex and can be safely shared between
// multiple goroutines.
//
// Example:
//
//	counter := NewArcMutex(0)
//	counter.WithLock(func(value *int) {
//	    *value += 1
//	})
func NewArcMutex[T any](value T) *ArcMutex[T] {
	inner := arc.NewArc(mutexData[T]{
		data: value,
	})

	return &ArcMutex[T]{
		inner: inner,
	}
}

// Clone creates a new ArcMutex[T] that shares the same underlying data.
// This is safe for concurrent use and allows multiple goroutines to
// access the same mutable data through their own ArcMutex[T] instances.
//
// Example:
//
//	original := NewArcMutex(42)
//	clone := original.Clone()
//	// Both original and clone protect the same data
func (am *ArcMutex[T]) Clone() *ArcMutex[T] {
	if am == nil || am.inner == nil {
		return nil
	}

	clonedInner := am.inner.Clone()
	if clonedInner == nil {
		return nil
	}

	return &ArcMutex[T]{
		inner: clonedInner,
	}
}

// WithLock provides safe access to the underlying data by acquiring the
// mutex and calling the provided function with a pointer to the data.
//
// This is the primary way to access and modify the data in an ArcMutex[T].
// The mutex is automatically released when the function returns, preventing
// deadlocks and ensuring thread safety.
//
// Example:
//
//	counter := NewArcMutex(0)
//	counter.WithLock(func(value *int) {
//	    *value += 1
//	    fmt.Println("Counter is now:", *value)
//	})
func (am *ArcMutex[T]) WithLock(fn func(*T)) {
	if am == nil || am.inner == nil || fn == nil {
		return
	}

	innerData := am.inner.Get()
	if innerData == nil {
		return
	}

	innerData.mu.Lock()
	defer innerData.mu.Unlock()

	fn(&innerData.data)
}

// TryWithLock attempts to acquire the mutex and execute the provided function.
// If the mutex is already locked, it returns false immediately without blocking.
// If the mutex is successfully acquired, it executes the function and returns true.
//
// This is useful for non-blocking operations where you want to skip the
// operation if the mutex is not immediately available.
//
// Example:
//
//	counter := NewArcMutex(0)
//	success := counter.TryWithLock(func(value *int) {
//	    *value += 1
//	})
//	if success {
//	    fmt.Println("Counter was incremented")
//	} else {
//	    fmt.Println("Counter was busy")
//	}
func (am *ArcMutex[T]) TryWithLock(fn func(*T)) bool {
	if am == nil || am.inner == nil || fn == nil {
		return false
	}

	innerData := am.inner.Get()
	if innerData == nil {
		return false
	}

	if !innerData.mu.TryLock() {
		return false
	}
	defer innerData.mu.Unlock()

	fn(&innerData.data)
	return true
}

// WithLockResult provides safe access to the underlying data and returns
// a result from the provided function. This is useful when you need to
// read data from the ArcMutex[T] and return it.
//
// Example:
//
//	counter := NewArcMutex(42)
//	value := counter.WithLockResult(func(data *int) int {
//	    return *data * 2
//	})
//	fmt.Println("Double the counter:", value) // 84
func (am *ArcMutex[T]) WithLockResult(fn func(*T) interface{}) interface{} {
	if am == nil || am.inner == nil || fn == nil {
		return nil
	}

	innerData := am.inner.Get()
	if innerData == nil {
		return nil
	}

	innerData.mu.Lock()
	defer innerData.mu.Unlock()

	return fn(&innerData.data)
}

// RefCount returns the current reference count for debugging purposes.
// This indicates how many ArcMutex[T] instances share the same underlying data.
func (am *ArcMutex[T]) RefCount() int64 {
	if am == nil || am.inner == nil {
		return 0
	}
	return am.inner.RefCount()
}

// IsValid returns true if the ArcMutex[T] is valid and can be used.
// An ArcMutex[T] becomes invalid if its underlying Arc[T] is invalid.
func (am *ArcMutex[T]) IsValid() bool {
	return am != nil && am.inner != nil && am.inner.IsValid()
}

// Drop decrements the reference count and potentially frees the underlying data.
// After calling Drop(), the ArcMutex[T] should not be used.
//
// Returns true if this was the last reference and the data was freed.
func (am *ArcMutex[T]) Drop() bool {
	if am == nil || am.inner == nil {
		return false
	}
	return am.inner.Drop()
}

// TryLock attempts to acquire the mutex and execute the provided function within the specified timeout.
// If timeout <= 0, behaves like TryWithLock (non-blocking).
// Returns true if lock was acquired and function executed, false otherwise.
//
// Example:
//
//	counter := NewArcMutex(0)
//	ok := counter.TryLock(10*time.Millisecond, func(val *int) { *val += 1 })
//	if ok { ... }
func (am *ArcMutex[T]) TryLock(timeout time.Duration, fn func(*T)) bool {
	if am == nil || am.inner == nil || fn == nil {
		return false
	}
	innerData := am.inner.Get()
	if innerData == nil {
		return false
	}
	if timeout <= 0 {
		if innerData.mu.TryLock() {
			defer innerData.mu.Unlock()
			fn(&innerData.data)
			return true
		}
		return false
	}
	deadline := time.Now().Add(timeout)
	for {
		if innerData.mu.TryLock() {
			defer innerData.mu.Unlock()
			fn(&innerData.data)
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(1 * time.Millisecond)
	}
}

// IsLocked returns true if the mutex is currently locked by any goroutine.
// This is a best-effort check and should only be used for debugging or metrics.
// It is not race-free and may be inaccurate in highly concurrent scenarios.
func (am *ArcMutex[T]) IsLocked() bool {
	if am == nil || am.inner == nil {
		return false
	}
	innerData := am.inner.Get()
	if innerData == nil {
		return false
	}
	if innerData.mu.TryLock() {
		innerData.mu.Unlock()
		return false
	}
	return true
}
