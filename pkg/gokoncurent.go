// Package gokoncurent provides safe and structured concurrency primitives for Go,
// inspired by Rust's ownership and sync model.
//
// This library is designed to work with Go 1.24 and later versions,
// taking advantage of features like atomic.Pointer[T], maps.Clone, and enhanced
// compile-time error checking to provide memory-safe concurrency patterns.
//
// The library provides several core primitives:
//
//   - Arc[T]: Atomic reference counting for shared ownership
//   - ArcMutex[T]: Safe shared mutable state with controlled access
//   - RWArcMutex[T]: Thread-safe read-write mutex for shared mutable state
//   - CondVar: Conditional variables for goroutine coordination
//   - Barrier: Synchronization primitive for waiting for multiple goroutines
//   - OnceCell[T]: Thread-safe lazy initialization
//   - SafeMap[K,V]: Concurrent map operations without data races
//   - TaskPool & Future[T]: Structured async task management
//
// Example usage:
//
//	import "github.com/Gosayram/gokoncurent"
//
//	// Arc[T] - Atomic Reference Counting
//	data := gokoncurent.NewArc("Hello, World!")
//	clone := data.Clone()
//	fmt.Println(*data.Get()) // "Hello, World!"
//
//	// ArcMutex[T] - Safe shared mutable state
//	counter := gokoncurent.NewArcMutex(0)
//	counter.WithLock(func(value *int) {
//	    *value += 1
//	})
//
//	// RWArcMutex[T] - Read-write mutex for shared state
//	rwCounter := gokoncurent.NewRWArcMutex(0)
//	rwCounter.WithRLock(func(value *int) { fmt.Println(*value) })
//	rwCounter.WithLock(func(value *int) { *value += 1 })
//
//	// CondVar - Conditional variables for coordination
//	cv := gokoncurent.NewCondVar()
//	go func() { cv.Wait(); fmt.Println("Signaled!") }()
//	cv.Signal()
//
//	// Barrier - Synchronization for multiple goroutines
//	b := gokoncurent.NewBarrier(3)
//	go func() { b.Wait(); fmt.Println("All workers synchronized!") }()
//
//	// OnceCell[T] - Lazy initialization
//	cell := gokoncurent.NewOnceCell[string]()
//	cell.Set("initialized once")
//	value, ok := cell.Get()
//	fmt.Println(value, ok) // "initialized once", true
//
// For more examples, see the examples/ directory.
package gokoncurent

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/Gosayram/gokoncurent/pkg/arc"
	"github.com/Gosayram/gokoncurent/pkg/arcmutex"
	"github.com/Gosayram/gokoncurent/pkg/barrier"
	"github.com/Gosayram/gokoncurent/pkg/condvar"
	"github.com/Gosayram/gokoncurent/pkg/oncecell"
	"github.com/Gosayram/gokoncurent/pkg/rwarcmutex"
)

// Version information
const (
	Version   = "0.1.0"
	BuildTime = "2025-01-04T00:00:00Z"
)

// Arc re-exports the Arc[T] type from the arc package for convenience.
type Arc[T any] struct {
	*arc.Arc[T]
}

// ArcMutex re-exports the ArcMutex[T] type from the arcmutex package for convenience.
type ArcMutex[T any] struct {
	*arcmutex.ArcMutex[T]
}

// RWArcMutex re-exports the RWArcMutex[T] type from the rwarcmutex package for convenience.
type RWArcMutex[T any] struct {
	*rwarcmutex.RWArcMutex[T]
}

// CondVar re-exports the CondVar type from the condvar package for convenience.
type CondVar struct {
	*condvar.CondVar
}

// Barrier re-exports the Barrier type from the barrier package for convenience.
type Barrier struct {
	*barrier.Barrier
}

// OnceCell re-exports the OnceCell[T] type from the oncecell package for convenience.
type OnceCell[T any] struct {
	*oncecell.OnceCell[T]
}

// NewArc creates a new Arc[T] with the given value.
func NewArc[T any](value T) *Arc[T] {
	return &Arc[T]{Arc: arc.NewArc(value)}
}

// NewArcMutex creates a new ArcMutex[T] with the given value.
func NewArcMutex[T any](value T) *ArcMutex[T] {
	return &ArcMutex[T]{ArcMutex: arcmutex.NewArcMutex(value)}
}

// NewRWArcMutex creates a new RWArcMutex[T] with the given value.
func NewRWArcMutex[T any](value T) *RWArcMutex[T] {
	return &RWArcMutex[T]{RWArcMutex: rwarcmutex.NewRWArcMutex(value)}
}

// NewCondVar creates a new CondVar for goroutine coordination.
func NewCondVar() *CondVar {
	return &CondVar{CondVar: condvar.NewCondVar()}
}

// NewBarrier creates a new Barrier for synchronizing multiple goroutines.
func NewBarrier(n int) *Barrier {
	return &Barrier{Barrier: barrier.NewBarrier(n)}
}

// NewOnceCell creates a new OnceCell[T] for lazy initialization.
func NewOnceCell[T any]() *OnceCell[T] {
	return &OnceCell[T]{OnceCell: oncecell.NewOnceCell[T]()}
}

// Info contains information about the library
type Info struct {
	Version     string
	GoVersion   string
	BuildTime   string
	GitCommit   string
	GitBranch   string
	GitModified bool
}

// GetInfo returns information about the library
func GetInfo() *Info {
	info := &Info{
		Version:   Version,
		GoVersion: runtime.Version(),
		BuildTime: BuildTime,
	}

	// Try to get git information from build info
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				info.GitCommit = setting.Value
			case "vcs.modified":
				info.GitModified = setting.Value == "true"
			}
		}
	}

	return info
}

// String returns a string representation of the Info
func (i *Info) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Version: %s", i.Version), fmt.Sprintf("Go Version: %s", i.GoVersion))

	if i.BuildTime != "" {
		parts = append(parts, fmt.Sprintf("Build Time: %s", i.BuildTime))
	}

	if i.GitCommit != "" {
		parts = append(parts, fmt.Sprintf("Git Commit: %s", i.GitCommit))
	}

	if i.GitModified {
		parts = append(parts, "Git Modified: true")
	}

	return strings.Join(parts, "\n")
}
