package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Gosayram/gokoncurent/pkg/barrier"
)

func main() {
	fmt.Println("=== Barrier Examples ===\n")

	// Example 1: Basic barrier synchronization
	fmt.Println("1. Basic Barrier Synchronization:")
	basicBarrier()

	// Example 2: Barrier with different arrival times
	fmt.Println("\n2. Barrier with Different Arrival Times:")
	barrierWithDelays()

	// Example 3: Barrier reset
	fmt.Println("\n3. Barrier Reset:")
	barrierReset()

	// Example 4: Barrier broken by drop
	fmt.Println("\n4. Barrier Broken by Drop:")
	barrierBroken()

	// Example 5: Multiple cycles
	fmt.Println("\n5. Multiple Cycles:")
	multipleCycles()

	// Example 6: Reference counting
	fmt.Println("\n6. Reference Counting:")
	referenceCounting()

	// Example 7: Worker coordination
	fmt.Println("\n7. Worker Coordination:")
	workerCoordination()
}

func basicBarrier() {
	b := barrier.NewBarrier(3)
	var wg sync.WaitGroup
	var results []string
	var mu sync.Mutex

	// Start 3 workers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  Worker %d starting work...\n", id)

			// Simulate some work
			time.Sleep(time.Duration(id*50) * time.Millisecond)

			fmt.Printf("  Worker %d finished work, waiting at barrier\n", id)
			result := b.Wait()

			mu.Lock()
			results = append(results, fmt.Sprintf("worker-%d", id))
			mu.Unlock()

			if result {
				fmt.Printf("  Worker %d crossed barrier successfully\n", id)
			} else {
				fmt.Printf("  Worker %d barrier was broken\n", id)
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("  All workers completed: %v\n", results)
}

func barrierWithDelays() {
	b := barrier.NewBarrier(4)
	var wg sync.WaitGroup
	var results []string
	var mu sync.Mutex

	// Start workers with different delays
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  Worker %d starting (will take %dms)\n", id, id*100)

			// Simulate different work durations
			time.Sleep(time.Duration(id*100) * time.Millisecond)

			fmt.Printf("  Worker %d arrived at barrier\n", id)
			result := b.Wait()

			mu.Lock()
			results = append(results, fmt.Sprintf("worker-%d", id))
			mu.Unlock()

			if result {
				fmt.Printf("  Worker %d: all workers synchronized!\n", id)
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("  Synchronization order: %v\n", results)
}

func barrierReset() {
	b := barrier.NewBarrier(3)
	var wg sync.WaitGroup

	// First cycle with 3 workers
	fmt.Println("  First cycle (3 workers):")
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("    Worker %d in first cycle\n", id)
			b.Wait()
			fmt.Printf("    Worker %d completed first cycle\n", id)
		}(i)
	}
	wg.Wait()

	// Reset barrier for 2 workers
	fmt.Println("  Resetting barrier for 2 workers...")
	b.Reset(2)

	// Second cycle with 2 workers
	fmt.Println("  Second cycle (2 workers):")
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("    Worker %d in second cycle\n", id)
			b.Wait()
			fmt.Printf("    Worker %d completed second cycle\n", id)
		}(i)
	}
	wg.Wait()
}

func barrierBroken() {
	b := barrier.NewBarrier(3)
	var wg sync.WaitGroup
	var results []bool
	var mu sync.Mutex

	// Start 2 workers that will wait
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  Worker %d waiting at barrier...\n", id)
			result := b.Wait()

			mu.Lock()
			results = append(results, result)
			mu.Unlock()

			if result {
				fmt.Printf("  Worker %d: barrier crossed\n", id)
			} else {
				fmt.Printf("  Worker %d: barrier was broken\n", id)
			}
		}(i)
	}

	// Give workers time to start waiting
	time.Sleep(100 * time.Millisecond)

	// Break the barrier by dropping the last reference
	fmt.Println("  Breaking barrier...")
	b.Drop()

	wg.Wait()
	fmt.Printf("  Results: %v\n", results)
}

func multipleCycles() {
	b := barrier.NewBarrier(3)
	var wg sync.WaitGroup

	// Run multiple cycles
	for cycle := 1; cycle <= 3; cycle++ {
		fmt.Printf("  Cycle %d:\n", cycle)

		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id, cycleNum int) {
				defer wg.Done()
				fmt.Printf("    Worker %d starting cycle %d\n", id, cycleNum)

				// Simulate work
				time.Sleep(time.Duration(id*20) * time.Millisecond)

				result := b.Wait()
				if result {
					fmt.Printf("    Worker %d completed cycle %d\n", id, cycleNum)
				}
			}(i, cycle)
		}
		wg.Wait()
	}
}

func referenceCounting() {
	b := barrier.NewBarrier(2)
	fmt.Printf("  Initial ref count: %d\n", b.RefCount())

	// Clone the barrier
	clone1 := b.Clone()
	fmt.Printf("  After clone 1: %d\n", b.RefCount())

	clone2 := b.Clone()
	fmt.Printf("  After clone 2: %d\n", b.RefCount())

	// Drop clones
	clone1.Drop()
	fmt.Printf("  After dropping clone 1: %d\n", b.RefCount())

	clone2.Drop()
	fmt.Printf("  After dropping clone 2: %d\n", b.RefCount())

	// Drop original
	b.Drop()
	fmt.Printf("  After dropping original: %d\n", b.RefCount())

	fmt.Printf("  String representation: %s\n", b.String())
}

func workerCoordination() {
	b := barrier.NewBarrier(4)
	var wg sync.WaitGroup

	// Start 4 workers that coordinate in phases
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for phaseNum := 1; phaseNum <= 3; phaseNum++ {
				fmt.Printf("  Worker %d starting phase %d\n", id, phaseNum)

				// Simulate phase work
				time.Sleep(time.Duration(id*30) * time.Millisecond)

				// Wait for all workers to complete this phase
				result := b.Wait()
				if result {
					fmt.Printf("  Worker %d: all workers completed phase %d\n", id, phaseNum)
				} else {
					fmt.Printf("  Worker %d: barrier broken in phase %d\n", id, phaseNum)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("  All phases completed successfully\n")
}
