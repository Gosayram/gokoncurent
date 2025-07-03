// Package oncecell provides lazy initialization with thread safety.
// It is inspired by Rust's OnceCell<T> and provides a way to initialize
// a value exactly once, even in the presence of concurrent access.
package oncecell

import (
	"sync"
	"sync/atomic"
)

// OnceCell represents a thread-safe cell that can be written to only once,
// but read from many times. This is inspired by Rust's OnceCell<T>.
//
// OnceCell[T] uses Go 1.24's atomic.Pointer[T] for efficient lock-free reads
// after the initial write, with sync.Once ensuring exactly one write operation.
//
// This is useful for lazy initialization patterns where you want to compute
// a value only once and share it across multiple goroutines.
type OnceCell[T any] struct {
	once  sync.Once
	value atomic.Pointer[T]
}

// NewOnceCell creates a new empty OnceCell[T].
// The cell starts uninitialized and can be set exactly once.
//
// Example:
//
//	cell := NewOnceCell[string]()
//	cell.Set("Hello, World!")
//	value, ok := cell.Get()
//	fmt.Println(value, ok) // "Hello, World!" true
func NewOnceCell[T any]() *OnceCell[T] {
	return &OnceCell[T]{}
}

// Set attempts to set the value in the cell.
// This operation can only succeed once - subsequent calls will be ignored.
//
// Returns true if the value was successfully set (i.e., this was the first call),
// false if the cell was already initialized.
//
// This method is safe for concurrent use.
//
// Example:
//
//	cell := NewOnceCell[int]()
//	success1 := cell.Set(42)  // true
//	success2 := cell.Set(100) // false (ignored)
func (oc *OnceCell[T]) Set(value T) bool {
	if oc == nil {
		return false
	}

	var wasSet bool
	oc.once.Do(func() {
		oc.value.Store(&value)
		wasSet = true
	})

	return wasSet
}

// Get retrieves the value from the cell.
// Returns the value and true if the cell has been initialized,
// or the zero value and false if the cell is empty.
//
// This method is safe for concurrent use and is lock-free after
// the initial write operation.
//
// Example:
//
//	cell := NewOnceCell[string]()
//	value, ok := cell.Get()
//	if !ok {
//	    fmt.Println("Cell is empty")
//	}
//
//	cell.Set("Hello")
//	value, ok = cell.Get()
//	fmt.Println(value, ok) // "Hello" true
func (oc *OnceCell[T]) Get() (T, bool) {
	if oc == nil {
		var zero T
		return zero, false
	}

	ptr := oc.value.Load()
	if ptr == nil {
		var zero T
		return zero, false
	}

	return *ptr, true
}

// GetOrInit returns the value from the cell if it's initialized,
// otherwise initializes it with the result of the provided function
// and returns that value.
//
// The initialization function is called at most once, even under
// concurrent access. If multiple goroutines call GetOrInit simultaneously,
// only one will execute the initialization function.
//
// Example:
//
//	cell := NewOnceCell[string]()
//	value := cell.GetOrInit(func() string {
//	    return "Lazy initialized value"
//	})
//	fmt.Println(value) // "Lazy initialized value"
func (oc *OnceCell[T]) GetOrInit(init func() T) T {
	if oc == nil {
		var zero T
		return zero
	}

	// Fast path: check if already initialized
	if ptr := oc.value.Load(); ptr != nil {
		return *ptr
	}

	// Slow path: initialize
	var result T
	oc.once.Do(func() {
		result = init()
		oc.value.Store(&result)
	})

	// If another goroutine initialized it first, return that value
	if ptr := oc.value.Load(); ptr != nil {
		return *ptr
	}

	// This should not happen under normal circumstances
	return result
}

// GetOrInitWith returns the value from the cell if it's initialized,
// otherwise initializes it with the provided value and returns it.
//
// This is a convenience method equivalent to GetOrInit(func() T { return value }).
//
// Example:
//
//	cell := NewOnceCell[int]()
//	value := cell.GetOrInitWith(42)
//	fmt.Println(value) // 42
func (oc *OnceCell[T]) GetOrInitWith(value T) T {
	return oc.GetOrInit(func() T { return value })
}

// IsInitialized returns true if the cell has been initialized.
// This method is safe for concurrent use and is lock-free.
//
// Example:
//
//	cell := NewOnceCell[string]()
//	fmt.Println(cell.IsInitialized()) // false
//	cell.Set("Hello")
//	fmt.Println(cell.IsInitialized()) // true
func (oc *OnceCell[T]) IsInitialized() bool {
	if oc == nil {
		return false
	}
	return oc.value.Load() != nil
}

// TryGet attempts to get the value from the cell without blocking.
// This is identical to Get() but provides a more explicit name
// for non-blocking access patterns.
//
// Returns the value and true if the cell is initialized,
// or the zero value and false if the cell is empty.
func (oc *OnceCell[T]) TryGet() (T, bool) {
	return oc.Get()
}

// Reset creates a new OnceCell[T] with the same type.
// This doesn't actually reset the current cell (which is impossible
// due to sync.Once semantics), but returns a new empty cell.
//
// This method is useful when you need to restart the lazy initialization
// process with a new cell instance.
//
// Example:
//
//	cell := NewOnceCell[string]()
//	cell.Set("old value")
//	newCell := cell.Reset()
//	// newCell is empty and can be initialized again
func (oc *OnceCell[T]) Reset() *OnceCell[T] {
	return NewOnceCell[T]()
}
