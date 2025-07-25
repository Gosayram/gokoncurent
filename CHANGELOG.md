# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- Fixed atomic.Int64 copying issue in Arc[T] by using pointer to atomic.Int64
- Fixed generic type aliases compatibility by using struct wrappers instead of type aliases
- Fixed append combine issue in Info.String() method
- Removed unused constant `numOperations` in arcmutex tests
- Added proper package comments for all packages
- Fixed all linting issues (gochecknoinits, gocritic, mnd, revive, unused)
- Fixed possible deadlock in CondVar.WaitWithContext and LockUnlock tests (no recursive locking, correct context handling)
- Fixed possible negative refCount in CondVar.Drop (now atomic and never goes below zero)
- Optimized CondVar stress and concurrent tests for speed and reliability
- Enhanced errcheck configuration with comprehensive exclusions for sync operations, time functions, and atomic operations
- ArcMutex[T]:
  - Fixed race in TryLock tests by synchronizing goroutine exit and main thread
  - All tests are now robust and race-free, comments and docs in English
  - Removed defer inside loop in TryLock (gocritic deferInLoop)
  - Fixed race in TryLock tests (synchronized goroutine exit)
- OnceCell[T]:
  - Correct error handling in GetOrInitWithRetry (returns lastErr on failure)
  - Long signature wrapped to satisfy lll lint rule
  - Benchmark updated to handle error (errcheck)

### Changed
- Cleaned up project structure by removing unused empty directories (docs/, internal/, testdata/)
- Reorganized project structure with proper package separation:
  - Moved Arc[T] to `pkg/arc/` package
  - Moved ArcMutex[T] to `pkg/arcmutex/` package  
  - Moved OnceCell[T] to `pkg/oncecell/` package
  - Updated main library file to re-export types from sub-packages
  - Created separate executable examples in `examples/basic/`
  - Added proper go.mod files for examples
- Simplified project structure by removing empty directories and unused folders
- Moved advanced usage example from `examples/advanced/` to `examples/` root

- Added comprehensive Makefile with development tools and CI/CD support

### Added
- RWArcMutex[T]: Thread-safe read-write mutex for shared mutable state with atomic reference counting
- CondVar: Conditional variables for goroutine coordination with atomic reference counting
  - Support for context cancellation and timeouts
  - Convenience functions `Notify()` and `NotifyBroadcast()`
  - Similar to `sync.Cond` but with Arc semantics
- Barrier: Synchronization primitive for waiting for multiple goroutines
  - Atomic reference counting with safe cleanup
  - Support for barrier reset and multiple cycles
  - Thread-safe coordination of N goroutines
- Arc[T]:
  - NewFromPointer: create Arc from existing pointer
  - CloneMany: efficient cloning of multiple references
  - Equal: pointer equality check for Arc
  - String: implements fmt.Stringer for debug output
- ArcMutex[T]:
  - TryLock: attempt to acquire mutex with timeout (race-free, polling, no goroutine)
  - IsLocked: best-effort check if mutex is currently locked (for debugging/metrics)
- OnceCell[T]:
  - ResetWithCallback: reset cell and invoke callback with old value (cleanup/logging)
  - GetOrInitWithRetry: lazy initialization with retry and exponential backoff
- Comprehensive Makefile with targets for:
  - Building and testing
  - Code quality checks (lint, staticcheck, security scan)
  - Benchmarking and performance analysis
  - SBOM generation
  - Version management
  - CI/CD support
- Development tools integration (golangci-lint, staticcheck, gosec, govulncheck, syft)
- Security scanning and vulnerability checking
- Software Bill of Materials (SBOM) generation
- Matrix testing support for multiple Go versions
- **Stress Test Suite**: High-concurrency race-condition tests for all primitives (`Arc`, `ArcMutex`, `RWArcMutex`, `OnceCell`, `Barrier`, `CondVar`) ensuring race-free operation under `go test -race`.
- **Advanced Examples**:
  - `advanced/map_slice_example`: Safe concurrent manipulation of `map[string][]int` using `ArcMutex`.
  - `advanced/oncecell_error_example`: Robust error-handling pattern with `OnceCell.GetOrInitWithRetry` and exponential back-off.

## [0.1.0] - 2025-01-04

### Added
- Initial library structure and project setup for safe concurrency primitives
- Go 1.24 support and modern tooling configuration
- Version management with `.go-version` and `.release-version` files
- Complete project documentation (README.md, CONTRIBUTING.md)
- Arc[T] - Atomic reference counting for shared immutable data
- ArcMutex[T] - Thread-safe shared mutable data with controlled access
- OnceCell[T] - Lazy initialization with thread safety
- Comprehensive test suite with 90%+ code coverage
- Benchmarks demonstrating performance characteristics
- Examples demonstrating real-world usage patterns

### Performance
- OnceCell[T] achieves 2.8+ billion reads/second after initialization
- Zero-allocation access patterns for Arc[T] and OnceCell[T]
- Lock-free operations where possible

### Testing
- Comprehensive unit tests for all components
- Concurrency tests with race detection
- Stress tests with multiple goroutines
- Performance benchmarks
- Documentation examples

### Changed
- Repository URL updated to github.com/Gosayram/gokoncurent
- Library focus shifted to safe concurrency primitives inspired by Rust
- Architecture designed around Go 1.24 features (atomic.Pointer[T], maps.Clone)

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- N/A (initial release)

[Unreleased]: https://github.com/Gosayram/gokoncurent/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Gosayram/gokoncurent/releases/tag/v0.1.0 