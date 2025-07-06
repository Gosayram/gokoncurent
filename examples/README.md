# GoKoncurent Examples

This directory contains examples demonstrating how to use the GoKoncurent library's safe concurrency primitives.

## Available Examples

### advanced_usage.go

A comprehensive example showcasing all the core GoKoncurent primitives working together:

- **OnceCell[T]**: Lazy initialization of expensive resources
- **Arc[T]**: Safe sharing of immutable data across goroutines
- **ArcMutex[T]**: Thread-safe mutable shared state
- **Combined patterns**: Real-world usage scenarios

This example demonstrates:
1. Lazy configuration initialization with multiple workers
2. Atomic reference counting with automatic cleanup
3. Concurrent increments without race conditions
4. Lazy cache with thread-safe updates
5. Performance characteristics of lock-free operations

### map_slice_example

Demonstrates safe concurrent usage of `ArcMutex[map[string][]int]` where multiple
workers append random integers to slices stored in a shared map.  Highlights how
`WithLock` can guard complex data structures like maps and slices without data
races.

Run with:
```bash
cd examples/advanced/map_slice_example
go run .
```

### oncecell_error_example

Shows robust error-handling with `OnceCell[T]` using `GetOrInitWithRetry`.  The
example simulates an unreliable initialisation function that may fail a few
times before succeeding, illustrating retry and exponential back-off patterns.

Run with:
```bash
cd examples/advanced/oncecell_error_example
go run .
```

## Running Examples

### Option 1: Using Go Modules (Recommended)

1. Create a new directory for your project
2. Initialize a Go module:
   ```bash
   go mod init your-project-name
   ```
3. Add the GoKoncurent dependency:
   ```bash
   go get github.com/Gosayram/gokoncurent
   ```
4. Copy the example code and run it:
   ```bash
   go run main.go
   ```

### Option 2: Local Development

If you're working with the library locally:

1. Clone the repository
2. Navigate to the examples directory
3. Run the example directly:
   ```bash
   cd examples
   go run advanced_usage.go
   ```

## Key Concepts Demonstrated

### Arc[T] - Atomic Reference Counting

```go
// Create shared data
data := gokoncurent.NewArc("Hello, World!")

// Clone for sharing
clone := data.Clone()

// Safe concurrent access
value := data.Get()
fmt.Println(*value) // "Hello, World!"

// Automatic cleanup when all references are dropped
```

### ArcMutex[T] - Safe Shared Mutable State

```go
// Create shared mutable counter
counter := gokoncurent.NewArcMutex(0)

// Safe concurrent modification
counter.WithLock(func(value *int) {
    *value += 1
})

// Safe concurrent reading
result := counter.WithLockResult(func(value *int) interface{} {
    return *value
})
```

### OnceCell[T] - Lazy Initialization

```go
// Create once-cell
cell := gokoncurent.NewOnceCell[Config]()

// Initialize exactly once, even with concurrent access
config := cell.GetOrInit(func() Config {
    return loadExpensiveConfig()
})
```

## Performance Characteristics

- **OnceCell[T]**: Lock-free reads after initialization (billions of ops/sec)
- **Arc[T]**: Atomic reference counting with minimal overhead
- **ArcMutex[T]**: Controlled mutex access prevents deadlocks

## Real-World Use Cases

1. **Configuration Management**: Use OnceCell[T] for lazy-loaded config
2. **Shared Resources**: Use Arc[T] for immutable shared data
3. **Concurrent Counters**: Use ArcMutex[T] for thread-safe mutations
4. **Caching Systems**: Combine all three for robust caching solutions

## Best Practices

1. **Prefer immutability**: Use Arc[T] when data doesn't need mutation
2. **Minimize lock scope**: Keep ArcMutex[T] critical sections small
3. **Lazy initialization**: Use OnceCell[T] for expensive startup operations
4. **Resource cleanup**: Let reference counting handle memory management

## Learn More

- [Library Documentation](../README.md)
- [API Reference](https://godoc.org/github.com/Gosayram/gokoncurent)
- [Contributing Guidelines](../CONTRIBUTING.md)

---

🦀 **Rust-Inspired Safety**: All primitives follow Rust's ownership and borrowing principles
🔒 **Thread-Safe by Design**: No raw pointers or unsafe memory access
🚀 **Go 1.24 Optimized**: Leverages the latest Go features for maximum performance 