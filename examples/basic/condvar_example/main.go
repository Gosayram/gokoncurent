package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Gosayram/gokoncurent/pkg/condvar"
)

func main() {
	fmt.Println("=== CondVar Examples ===\n")

	// Example 1: Basic signal/wait pattern
	fmt.Println("1. Basic Signal/Wait Pattern:")
	basicSignalWait()

	// Example 2: Broadcast pattern
	fmt.Println("\n2. Broadcast Pattern:")
	broadcastPattern()

	// Example 3: Context cancellation
	fmt.Println("\n3. Context Cancellation:")
	contextCancellation()

	// Example 4: Timeout pattern
	fmt.Println("\n4. Timeout Pattern:")
	timeoutPattern()

	// Example 5: Convenience functions
	fmt.Println("\n5. Convenience Functions:")
	convenienceFunctions()

	// Example 6: Producer-consumer pattern
	fmt.Println("\n6. Producer-Consumer Pattern:")
	producerConsumer()

	// Example 7: Reference counting
	fmt.Println("\n7. Reference Counting:")
	referenceCounting()
}

func basicSignalWait() {
	cv := condvar.NewCondVar()
	var result string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("  Goroutine waiting...")
		cv.Wait()
		result = "signaled"
		fmt.Println("  Goroutine received signal")
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("  Signaling goroutine...")
	cv.Signal()

	wg.Wait()
	fmt.Printf("  Result: %s\n", result)
}

func broadcastPattern() {
	cv := condvar.NewCondVar()
	var results []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start multiple goroutines waiting
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  Worker %d waiting...\n", id)
			cv.Wait()
			mu.Lock()
			results = append(results, fmt.Sprintf("worker-%d", id))
			mu.Unlock()
			fmt.Printf("  Worker %d woke up\n", id)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println("  Broadcasting to all workers...")
	cv.Broadcast()

	wg.Wait()
	fmt.Printf("  All workers completed: %v\n", results)
}

func contextCancellation() {
	cv := condvar.NewCondVar()
	ctx, cancel := context.WithCancel(context.Background())

	var result bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("  Goroutine waiting with context...")
		result = cv.WaitWithContext(ctx)
		if result {
			fmt.Println("  Goroutine received signal")
		} else {
			fmt.Println("  Goroutine context cancelled")
		}
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("  Cancelling context...")
	cancel()

	wg.Wait()
	fmt.Printf("  Result: %t\n", result)
}

func timeoutPattern() {
	cv := condvar.NewCondVar()

	var result bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("  Goroutine waiting with timeout...")
		result = cv.WaitWithTimeout(200 * time.Millisecond)
		if result {
			fmt.Println("  Goroutine received signal")
		} else {
			fmt.Println("  Goroutine timed out")
		}
	}()

	wg.Wait()
	fmt.Printf("  Result: %t\n", result)
}

func convenienceFunctions() {
	// Using Notify function
	cv, signal := condvar.Notify()
	var result string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("  Goroutine waiting for notification...")
		cv.Wait()
		result = "notified"
		fmt.Println("  Goroutine received notification")
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("  Sending notification...")
	signal()

	wg.Wait()
	fmt.Printf("  Result: %s\n", result)

	// Using NotifyBroadcast function
	cv2, broadcast := condvar.NotifyBroadcast()
	var results []string
	var mu sync.Mutex

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  Listener %d waiting...\n", id)
			cv2.Wait()
			mu.Lock()
			results = append(results, fmt.Sprintf("listener-%d", id))
			mu.Unlock()
			fmt.Printf("  Listener %d received broadcast\n", id)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println("  Broadcasting to all listeners...")
	broadcast()

	wg.Wait()
	fmt.Printf("  All listeners notified: %v\n", results)
}

func producerConsumer() {
	cv := condvar.NewCondVar()
	var data []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Consumer goroutines
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				cv.Wait()
				mu.Lock()
				if len(data) > 0 {
					item := data[0]
					data = data[1:]
					mu.Unlock()
					fmt.Printf("  Consumer %d processed: %d\n", id, item)
					if item == -1 { // Stop signal
						return
					}
				} else {
					mu.Unlock()
				}
			}
		}(i)
	}

	// Producer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 5; i++ {
			time.Sleep(50 * time.Millisecond)
			mu.Lock()
			data = append(data, i)
			mu.Unlock()
			fmt.Printf("  Producer added: %d\n", i)
			cv.Signal()
		}
		// Send stop signal
		mu.Lock()
		data = append(data, -1, -1)
		mu.Unlock()
		cv.Broadcast()
	}()

	wg.Wait()
	fmt.Println("  Producer-consumer completed")
}

func referenceCounting() {
	cv := condvar.NewCondVar()
	fmt.Printf("  Initial ref count: %d\n", cv.RefCount())

	// Clone the conditional variable
	clone1 := cv.Clone()
	fmt.Printf("  After clone 1: %d\n", cv.RefCount())

	clone2 := cv.Clone()
	fmt.Printf("  After clone 2: %d\n", cv.RefCount())

	// Drop clones
	clone1.Drop()
	fmt.Printf("  After dropping clone 1: %d\n", cv.RefCount())

	clone2.Drop()
	fmt.Printf("  After dropping clone 2: %d\n", cv.RefCount())

	// Drop original
	cv.Drop()
	fmt.Printf("  After dropping original: %d\n", cv.RefCount())

	fmt.Printf("  String representation: %s\n", cv.String())
}
