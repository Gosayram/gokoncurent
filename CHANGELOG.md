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