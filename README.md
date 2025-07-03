# GoKoncurent

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/Gosayram/gokoncurent)](https://goreportcard.com/report/github.com/Gosayram/gokoncurent)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/Gosayram/gokoncurent?status.svg)](https://godoc.org/github.com/Gosayram/gokoncurent)

Safe and structured concurrency primitives for Go, inspired by Rust's ownership and sync model.

## Features

- 🦀 **Rust-Inspired Safety**: Memory-safe concurrency patterns inspired by Rust's Arc, Mutex, OnceCell
- 🔒 **Thread-Safe by Design**: All operations guarantee safe concurrent access
- 🚀 **Go 1.24 Optimized**: Leverages latest Go features like `atomic.Pointer[T]` and `maps.Clone`
- 🛡️ **No Raw Access**: Controlled API prevents data races and memory corruption
- 📦 **Zero Dependencies**: Pure Go implementation with no external dependencies
- 🧪 **Well Tested**: Comprehensive test suite with >90% coverage
- 🎯 **Production Ready**: Battle-tested concurrency primitives

## Installation

```bash
go get github.com/Gosayram/gokoncurent
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/Gosayram/gokoncurent"
)

func main() {
    // Arc[T] - Atomic Reference Counting
    data := gokoncurent.NewArc("Hello, World!")
    clone := data.Clone()
    fmt.Println(*data.Get()) // "Hello, World!"
    
    // ArcMutex[T] - Safe shared mutable state
    counter := gokoncurent.NewArcMutex(0)
    counter.WithLock(func(value *int) {
        *value += 1
    })
    
    // OnceCell[T] - Lazy initialization
    cell := gokoncurent.NewOnceCell[string]()
    cell.Set("initialized once")
    value, ok := cell.Get()
    fmt.Println(value, ok) // "initialized once", true
}
```

## Architecture

GoKoncurent provides safe concurrency primitives organized in phases:

### 📦 Phase 1: Arc[T] - Atomic Reference Counting
- Thread-safe reference counting using `atomic.Int64`
- Rust-like `Clone()` method for sharing between goroutines
- Automatic resource cleanup when reference count reaches zero
- Safe `Get()` access without raw pointer exposure

### 📦 Phase 2: ArcMutex[T] - Shared Mutable State
- Wrapper around `sync.Mutex` with safe API
- Access only through `WithLock(func(*T))` to prevent deadlocks
- Built on Arc[T] for safe sharing between goroutines

### 📦 Phase 3: OnceCell[T] - Lazy Initialization
- Rust-like OnceCell/Lazy equivalent
- Uses `sync.Once` and `atomic.Pointer[T]` from Go 1.24
- `Set(value T)` and `Get() (T, bool)` methods

### 📦 Phase 4: SafeMap[K, V] - Concurrent Map Operations
- Race-free map operations with `sync.RWMutex`
- Utilizes `maps.Clone` from Go 1.24
- Snapshot and iteration support without data races

### 📦 Phase 5: TaskPool & Future[T] - Async Task Management
- Simplified API for managing N goroutines
- `TaskPool.Run(ctx, func())` with context control
- `Future[T]` for async result handling

## Go 1.24 Features Used

| Feature | Usage |
|---------|-------|
| `atomic.Pointer[T]` | OnceCell and Arc implementations |
| `maps.Clone`, `maps.Equal` | SafeMap operations |
| `slices.Compact`, `slices.Delete` | Future SafeSlice support |
| Generic `sync.Pool` | Arc[T] and Future[T] allocation |
| Enhanced compile errors | Early detection of unsafe access patterns |

## Requirements

- Go 1.24 or later
- No external dependencies

## Documentation

### API Reference

Full API documentation is available at [GoDoc](https://godoc.org/github.com/Gosayram/gokoncurent).

### Examples

See the [examples](examples/) directory for comprehensive usage examples.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/Gosayram/gokoncurent.git
   cd gokoncurent
   ```

2. Install development tools:
   ```bash
   go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

4. Run linter:
   ```bash
   go tool golangci-lint run
   ```

### Code Quality

- All code must pass `go fmt`, `go vet`, and `golangci-lint`
- Test coverage must be maintained at >90%
- All public APIs must be documented
- Follow semantic versioning for releases

## Versioning

This project uses [Semantic Versioning](https://semver.org/). For the versions available, see the [tags on this repository](https://github.com/Gosayram/gokoncurent/tags).

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://godoc.org/github.com/Gosayram/gokoncurent)
- 🐛 [Issues](https://github.com/Gosayram/gokoncurent/issues)
- 💬 [Discussions](https://github.com/Gosayram/gokoncurent/discussions)

## Acknowledgments

- Built with [Go 1.24](https://golang.org/)
- Inspired by [Rust's concurrency model](https://doc.rust-lang.org/book/ch16-00-fearless-concurrency.html)
- Thanks to all contributors

---

Made with ❤️ for safe Go concurrency 