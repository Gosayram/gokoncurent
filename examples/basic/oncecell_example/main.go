package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Gosayram/gokoncurent/pkg/oncecell"
)

func main() {
	fmt.Println("=== OnceCell[T] Basic Example ===\n")

	// Create lazy-initialized configuration
	config := oncecell.NewOnceCell[map[string]string]()

	// Initialize with expensive operation
	initFunc := func() map[string]string {
		fmt.Println("Initializing configuration...")
		time.Sleep(100 * time.Millisecond) // Simulate expensive operation
		return map[string]string{
			"database_url": "postgres://localhost:5432/mydb",
			"api_key":      "secret-key-123",
			"debug":        "true",
		}
	}

	// Multiple goroutines try to get the same configuration
	var wg sync.WaitGroup
	wg.Add(5)

	for i := 0; i < 5; i++ {
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
		fmt.Printf("API Key: %s\n", cfg["api_key"])
		fmt.Printf("Debug: %s\n", cfg["debug"])
	}

	// Demonstrate Set method
	newConfig := oncecell.NewOnceCell[string]()
	newConfig.Set("Hello, OnceCell!")
	value, _ := newConfig.Get()
	fmt.Printf("Set value: %s\n", value)

	fmt.Println("✓ OnceCell[T] example completed successfully!")
}
