package main

import (
	"fmt"
	"sync"

	"github.com/Gosayram/gokoncurent/pkg/arc"
)

func main() {
	fmt.Println("=== Arc[T] Basic Example ===\n")

	// Create shared immutable data
	sharedData := arc.NewArc("Hello, World!")
	fmt.Printf("Created Arc with value: %s\n", *sharedData.Get())
	fmt.Printf("Initial reference count: %d\n", sharedData.RefCount())

	// Clone the Arc
	clone1 := sharedData.Clone()
	clone2 := sharedData.Clone()
	fmt.Printf("After cloning: reference count = %d\n", sharedData.RefCount())

	// Use in multiple goroutines
	var wg sync.WaitGroup
	wg.Add(3)

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

	fmt.Printf("After cleanup: reference count = %d\n", sharedData.RefCount())
	fmt.Println("✓ Arc[T] example completed successfully!")
}
