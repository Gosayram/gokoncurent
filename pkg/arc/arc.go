// Package arc provides atomic reference counting for shared immutable data.
// It is inspired by Rust's Arc<T> and provides thread-safe reference counting
// with automatic cleanup when the last reference is dropped.
package arc

import (
	"fmt"
	"sync/atomic"
)

// Arc represents an atomically reference-counted pointer to shared immutable data.
// It can be safely shared between multiple goroutines and automatically
// cleans up the underlying data when the last reference is dropped.
//
// Arc[T] is inspired by Rust's Arc<T> and provides similar safety guarantees
// for shared immutable data in Go.
type Arc[T any] struct {
	data     *T
	refCount *atomic.Int64
}

// NewArc creates a new Arc[T] with the given value.
// The returned Arc[T] has a reference count of 1.
//
// Example:
//
//	shared := NewArc("Hello, World!")
//	defer shared.Drop()
func NewArc[T any](value T) *Arc[T] {
	refCount := &atomic.Int64{}
	refCount.Store(1)

	return &Arc[T]{
		data:     &value,
		refCount: refCount,
	}
}

// NewFromPointer creates a new Arc[T] from an existing pointer.
// The caller is responsible for ensuring the pointer is valid and not shared.
// The returned Arc[T] has a reference count of 1.
//
// Example:
//
//	data := &MyStruct{Field: "value"}
//	arc := NewFromPointer(data)
//	defer arc.Drop()
func NewFromPointer[T any](ptr *T) *Arc[T] {
	if ptr == nil {
		return nil
	}

	refCount := &atomic.Int64{}
	refCount.Store(1)

	return &Arc[T]{
		data:     ptr,
		refCount: refCount,
	}
}

// Clone creates a new Arc[T] that shares the same underlying data.
// This increments the reference count and returns a new Arc[T] instance.
// The original Arc[T] remains valid and can still be used.
//
// Example:
//
//	original := NewArc(42)
//	clone := original.Clone()
//	// Both original and clone point to the same data
//	// Reference count is now 2
func (a *Arc[T]) Clone() *Arc[T] {
	if a == nil {
		return nil
	}

	// Increment reference count atomically
	newCount := a.refCount.Add(1)
	if newCount <= 1 {
		// This should never happen in normal usage
		// but we handle it gracefully
		a.refCount.Add(-1)
		return nil
	}

	return &Arc[T]{
		data:     a.data,
		refCount: a.refCount,
	}
}

// CloneMany creates multiple clones of the Arc[T] at once.
// This is more efficient than calling Clone() multiple times
// as it only increments the reference count once.
//
// Example:
//
//	original := NewArc("shared data")
//	clones := original.CloneMany(3)
//	// Reference count is now 4 (original + 3 clones)
func (a *Arc[T]) CloneMany(count int) []*Arc[T] {
	if a == nil || count <= 0 {
		return nil
	}

	// Increment reference count by count atomically
	newCount := a.refCount.Add(int64(count))
	if newCount <= int64(count) {
		// This should never happen in normal usage
		// but we handle it gracefully
		a.refCount.Add(-int64(count))
		return nil
	}

	clones := make([]*Arc[T], count)
	for i := 0; i < count; i++ {
		clones[i] = &Arc[T]{
			data:     a.data,
			refCount: a.refCount,
		}
	}

	return clones
}

// Get returns a pointer to the underlying data.
// The returned pointer is valid as long as the Arc[T] is valid.
// This operation is lock-free and safe for concurrent access.
//
// Example:
//
//	arc := NewArc("Hello")
//	data := arc.Get()
//	fmt.Println(*data) // "Hello"
func (a *Arc[T]) Get() *T {
	if a == nil || a.data == nil {
		return nil
	}
	return a.data
}

// RefCount returns the current reference count.
// This is mainly useful for debugging and should not be used
// for synchronization purposes.
func (a *Arc[T]) RefCount() int64 {
	if a == nil || a.refCount == nil {
		return 0
	}
	return a.refCount.Load()
}

// Drop decrements the reference count and potentially frees the underlying data.
// If this is the last reference, the data is freed.
// After calling Drop(), the Arc[T] should not be used.
//
// Returns true if this was the last reference and the data was freed.
//
// Example:
//
//	arc := NewArc("Hello")
//	clone := arc.Clone()
//	arc.Drop()  // Reference count is now 1
//	clone.Drop() // Reference count is now 0, data is freed
func (a *Arc[T]) Drop() bool {
	if a == nil || a.data == nil || a.refCount == nil {
		return false
	}

	newCount := a.refCount.Add(-1)
	if newCount == 0 {
		// This was the last reference, clean up
		a.data = nil
		a.refCount = nil
		return true
	}
	return false
}

// IsValid returns true if the Arc[T] is valid and can be used.
// An Arc[T] becomes invalid if it was nil or if Drop() was called
// and this was the last reference.
func (a *Arc[T]) IsValid() bool {
	return a != nil && a.data != nil && a.refCount != nil && a.refCount.Load() > 0
}

// Equal returns true if two Arc[T] instances point to the same underlying data.
// This is a pointer comparison, not a value comparison.
//
// Example:
//
//	arc1 := NewArc("hello")
//	arc2 := arc1.Clone()
//	arc3 := NewArc("hello")
//	arc1.Equal(arc2) // true (same data)
//	arc1.Equal(arc3) // false (different data)
func (a *Arc[T]) Equal(other *Arc[T]) bool {
	if a == nil || other == nil {
		return a == other
	}
	return a.data == other.data
}

// String implements fmt.Stringer interface.
// Returns a string representation of the Arc[T] including the reference count.
//
// Example:
//
//	arc := NewArc("hello")
//	fmt.Println(arc) // "Arc{refCount: 1}"
func (a *Arc[T]) String() string {
	if a == nil {
		return "Arc<nil>"
	}
	return fmt.Sprintf("Arc{refCount: %d}", a.RefCount())
}
