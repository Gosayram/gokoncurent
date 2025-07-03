// Package main demonstrates advanced usage of the GoKoncurent library,
// showcasing Arc[T], ArcMutex[T], and OnceCell[T] working together
// in a practical concurrent programming scenario.
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Gosayram/gokoncurent/pkg/arc"
	"github.com/Gosayram/gokoncurent/pkg/arcmutex"
	"github.com/Gosayram/gokoncurent/pkg/oncecell"
)

const (
	numGoroutines = 3
	numWorkers    = 5
	numKeys       = 4
	configDelay   = 100 * time.Millisecond
	dbQueryDelay  = 50 * time.Millisecond
)

// Config represents application configuration that should be initialized once
type Config struct {
	DatabaseURL string
	MaxWorkers  int
	Timeout     time.Duration
}

// Counter represents a thread-safe counter that can be shared
type Counter struct {
	value    *arcmutex.ArcMutex[int]
	metadata *arc.Arc[string]
}

// NewCounter creates a new Counter with initial value and metadata
func NewCounter(initialValue int, metadata string) *Counter {
	return &Counter{
		value:    arcmutex.NewArcMutex(initialValue),
		metadata: arc.NewArc(metadata),
	}
}

// Clone creates a new Counter instance sharing the same data
func (c *Counter) Clone() *Counter {
	return &Counter{
		value:    c.value.Clone(),
		metadata: c.metadata.Clone(),
	}
}

// Increment increments the counter value
func (c *Counter) Increment() {
	c.value.WithLock(func(v *int) {
		*v++
	})
}

// Get returns the current counter value
func (c *Counter) Get() int {
	return c.value.WithLockResult(func(v *int) interface{} {
		return *v
	}).(int)
}

// GetMetadata returns the counter metadata
func (c *Counter) GetMetadata() string {
	return *c.metadata.Get()
}

// Example demonstrating advanced concurrent usage of all primitives
func main() {
	fmt.Println("=== GoKoncurent Advanced Usage Example ===")

	// 1. Arc[T] - Shared immutable data
	fmt.Println("1. Arc[T] - Shared Immutable Data")
	demoArc()
	fmt.Println()

	// 2. ArcMutex[T] - Shared mutable data with controlled access
	fmt.Println("2. ArcMutex[T] - Shared Mutable Data")
	demoArcMutex()
	fmt.Println()

	// 3. OnceCell[T] - Lazy initialization
	fmt.Println("3. OnceCell[T] - Lazy Initialization")
	demoOnceCell()
	fmt.Println()

	// 4. Combined usage - Real-world scenario
	fmt.Println("4. Combined Usage - Real-world Scenario")
	demoCombinedUsage()
}

func demoArc() {
	// Create shared immutable data
	sharedData := arc.NewArc("Hello, World!")

	// Clone for multiple goroutines
	clone1 := sharedData.Clone()
	clone2 := sharedData.Clone()

	fmt.Printf("Initial reference count: %d\n", sharedData.RefCount())

	// Use in multiple goroutines
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	go func() {
		defer wg.Done()
		data := sharedData.Get()
		fmt.Printf("Goroutine 1: %s\n", *data)
	}()

	go func() {
		defer wg.Done()
		data := clone1.Get()
		fmt.Printf("Goroutine 2: %s\n", *data)
	}()

	go func() {
		defer wg.Done()
		data := clone2.Get()
		fmt.Printf("Goroutine 3: %s\n", *data)
	}()

	wg.Wait()

	// Clean up
	sharedData.Drop()
	clone1.Drop()
	clone2.Drop()
}

func demoArcMutex() {
	// Create shared mutable counter
	counter := arcmutex.NewArcMutex(0)

	// Clone for multiple goroutines
	clone1 := counter.Clone()
	clone2 := counter.Clone()

	fmt.Printf("Initial counter value: %d\n", counter.WithLockResult(func(v *int) interface{} { return *v }).(int))

	// Increment from multiple goroutines
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			counter.WithLock(func(v *int) {
				*v++
			})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			clone1.WithLock(func(v *int) {
				*v++
			})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			clone2.WithLock(func(v *int) {
				*v++
			})
		}
	}()

	wg.Wait()

	finalValue := counter.WithLockResult(func(v *int) interface{} { return *v }).(int)
	fmt.Printf("Final counter value: %d (expected: 3000)\n", finalValue)

	// Clean up
	counter.Drop()
	clone1.Drop()
	clone2.Drop()
}

func demoOnceCell() {
	// Create lazy-initialized configuration
	config := oncecell.NewOnceCell[map[string]string]()

	// Initialize with expensive operation
	initFunc := func() map[string]string {
		fmt.Println("Initializing configuration...")
		time.Sleep(configDelay) // Simulate expensive operation
		return map[string]string{
			"database_url": "postgres://localhost:5432/mydb",
			"api_key":      "secret-key-123",
			"debug":        "true",
		}
	}

	// Multiple goroutines try to get the same configuration
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer wg.Done()
			cfg := config.GetOrInit(initFunc)
			fmt.Printf("Goroutine %d got config with %d items\n", id, len(cfg))
		}(i)
	}

	wg.Wait()

	// Verify all got the same configuration
	cfg, _ := config.Get()
	fmt.Printf("Configuration initialized: %v\n", config.IsInitialized())
	if cfg != nil {
		fmt.Printf("Database URL: %s\n", cfg["database_url"])
	}
}

func demoCombinedUsage() {
	// Real-world scenario: Shared cache with lazy initialization

	// Shared cache using ArcMutex
	cache := arcmutex.NewArcMutex(map[string]string{})

	// Lazy loader for cache entries
	loader := oncecell.NewOnceCell[func(string) string]()
	loader.Set(func(key string) string {
		fmt.Printf("Loading data for key: %s\n", key)
		time.Sleep(dbQueryDelay) // Simulate database query
		return fmt.Sprintf("Data for %s", key)
	})

	// Function to get or create cache entry
	getOrCreate := func(key string) string {
		// Try to get existing entry
		var value string
		found := cache.WithLockResult(func(c *map[string]string) interface{} {
			if v, exists := (*c)[key]; exists {
				value = v
				return true
			}
			return false
		}).(bool)

		if found {
			return value
		}

		// Create new entry
		loadFunc, _ := loader.Get()
		newValue := loadFunc(key)

		// Store in cache
		cache.WithLock(func(c *map[string]string) {
			(*c)[key] = newValue
		})

		return newValue
	}

	// Multiple goroutines accessing the same cache
	var wg sync.WaitGroup
	wg.Add(numKeys)

	keys := []string{"user:1", "user:2", "user:1", "user:3"}

	for i, key := range keys {
		go func(id int, k string) {
			defer wg.Done()
			value := getOrCreate(k)
			fmt.Printf("Goroutine %d: %s -> %s\n", id, k, value)
		}(i, key)
	}

	wg.Wait()

	// Show final cache state
	cache.WithLock(func(c *map[string]string) {
		fmt.Printf("Cache contains %d entries\n", len(*c))
		for key, value := range *c {
			fmt.Printf("  %s: %s\n", key, value)
		}
	})

	// Clean up
	cache.Drop()
	loader.Reset()
}
