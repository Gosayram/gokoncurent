package main

import (
	"fmt"
	"sync"

	"github.com/Gosayram/gokoncurent/pkg/arcmutex"
)

func main() {
	fmt.Println("=== ArcMutex[T] Basic Example ===\n")

	// Create shared mutable counter
	counter := arcmutex.NewArcMutex(0)
	fmt.Printf("Created counter with initial value: %d\n",
		counter.WithLockResult(func(v *int) interface{} { return *v }).(int))

	// Clone for multiple goroutines
	clone1 := counter.Clone()
	clone2 := counter.Clone()

	// Increment from multiple goroutines
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			counter.WithLock(func(v *int) {
				*v++
			})
		}
		fmt.Println("Goroutine 1: Completed 100 increments")
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			clone1.WithLock(func(v *int) {
				*v++
			})
		}
		fmt.Println("Goroutine 2: Completed 100 increments")
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			clone2.WithLock(func(v *int) {
				*v++
			})
		}
		fmt.Println("Goroutine 3: Completed 100 increments")
	}()

	wg.Wait()

	// Check final value
	finalValue := counter.WithLockResult(func(v *int) interface{} { return *v }).(int)
	fmt.Printf("Final counter value: %d (expected: 300)\n", finalValue)

	// Demonstrate TryWithLock
	success := counter.TryWithLock(func(v *int) {
		fmt.Printf("TryWithLock succeeded, current value: %d\n", *v)
	})
	if success {
		fmt.Println("✓ TryWithLock worked as expected")
	}

	// Clean up
	counter.Drop()
	clone1.Drop()
	clone2.Drop()

	fmt.Println("✓ ArcMutex[T] example completed successfully!")
}
