package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Gosayram/gokoncurent/pkg/rwarcmutex"
)

func main() {
	fmt.Println("=== RWArcMutex[T] Basic Example ===\n")

	// Create a shared counter with read-write mutex
	counter := rwarcmutex.NewRWArcMutex(0)
	fmt.Printf("Created counter with initial value: %d\n", 0)
	fmt.Printf("Initial reference count: %d\n", counter.RefCount())

	// Clone for multiple goroutines
	clone1 := counter.Clone()
	clone2 := counter.Clone()
	fmt.Printf("After cloning: reference count = %d\n", counter.RefCount())

	// Simulate multiple readers and one writer
	var wg sync.WaitGroup
	wg.Add(5) // 3 readers + 2 writers

	// Reader goroutines (can run concurrently)
	for i := 1; i <= 3; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				counter.WithRLock(func(v *int) {
					fmt.Printf("Reader %d: current value = %d\n", id, *v)
				})
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	// Writer goroutines (exclusive access)
	for i := 1; i <= 2; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				counter.WithLock(func(v *int) {
					*v += 10
					fmt.Printf("Writer %d: incremented to %d\n", id, *v)
				})
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Show final value
	counter.WithRLock(func(v *int) {
		fmt.Printf("Final counter value: %d\n", *v)
	})

	// Clean up
	counter.Drop()
	clone1.Drop()
	clone2.Drop()

	fmt.Printf("After cleanup: reference count = %d\n", counter.RefCount())
	fmt.Println("✓ RWArcMutex[T] example completed successfully!")
}
