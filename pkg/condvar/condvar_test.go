package condvar

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCondVar(t *testing.T) {
	cv := NewCondVar()
	assert.NotNil(t, cv)
	assert.Equal(t, int64(1), cv.RefCount())
}

func TestCondVar_Clone(t *testing.T) {
	cv := NewCondVar()
	initialCount := cv.RefCount()

	clone := cv.Clone()
	assert.Equal(t, cv, clone)
	assert.Equal(t, initialCount+1, cv.RefCount())
	assert.Equal(t, initialCount+1, clone.RefCount())
}

func TestCondVar_Drop(t *testing.T) {
	cv := NewCondVar()
	clone := cv.Clone()

	// Drop clone first
	clone.Drop()
	assert.Equal(t, int64(1), cv.RefCount())

	// Drop original
	cv.Drop()
	assert.Equal(t, int64(0), cv.RefCount())
}

func TestCondVar_WaitAndSignal(t *testing.T) {
	cv := NewCondVar()
	var result string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		cv.Wait()
		result = "signaled"
	}()

	// Give goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Signal the condition
	cv.Signal()

	wg.Wait()
	assert.Equal(t, "signaled", result)
}

func TestCondVar_WaitAndBroadcast(t *testing.T) {
	cv := NewCondVar()
	var results []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start multiple goroutines waiting
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cv.Wait()
			mu.Lock()
			results = append(results, fmt.Sprintf("goroutine-%d", id))
			mu.Unlock()
		}(i)
	}

	// Give goroutines time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Broadcast to all waiting goroutines
	cv.Broadcast()

	wg.Wait()
	assert.Len(t, results, 3)
}

func TestCondVar_WaitWithContext(t *testing.T) {
	cv := NewCondVar()
	ctx, cancel := context.WithCancel(context.Background())

	var result bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		result = cv.WaitWithContext(ctx)
	}()

	// Give goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Cancel context
	cancel()

	wg.Wait()
	assert.False(t, result)
}

func TestCondVar_WaitWithContext_Signaled(t *testing.T) {
	cv := NewCondVar()
	ctx := context.Background()

	var result bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		result = cv.WaitWithContext(ctx)
	}()

	// Give goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Signal the condition
	cv.Signal()

	wg.Wait()
	assert.True(t, result)
}

func TestCondVar_WaitWithTimeout(t *testing.T) {
	cv := NewCondVar()

	var result bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		result = cv.WaitWithTimeout(50 * time.Millisecond)
	}()

	wg.Wait()
	assert.False(t, result) // Should timeout
}

func TestCondVar_WaitWithTimeout_Signaled(t *testing.T) {
	cv := NewCondVar()

	var result bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		result = cv.WaitWithTimeout(100 * time.Millisecond)
	}()

	// Give goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Signal the condition
	cv.Signal()

	wg.Wait()
	assert.True(t, result)
}

func TestCondVar_LockUnlock(t *testing.T) {
	cv := NewCondVar()

	// Should not block
	cv.Lock()
	// Critical section - verify we can access the mutex
	_ = cv.RefCount()
	cv.Unlock()

	// Test basic lock/unlock functionality
	cv.Lock()
	// Critical section - verify we can access the mutex
	_ = cv.RefCount()
	cv.Unlock()
}

func TestCondVar_String(t *testing.T) {
	cv := NewCondVar()
	str := cv.String()
	assert.Contains(t, str, "CondVar{refCount: 1}")

	cv.Clone()
	str = cv.String()
	assert.Contains(t, str, "CondVar{refCount: 2}")
}

func TestNotify(t *testing.T) {
	cv, signal := Notify()
	assert.NotNil(t, cv)
	assert.NotNil(t, signal)

	var result string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		cv.Wait()
		result = "notified"
	}()

	// Give goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Signal using the convenience function
	signal()

	wg.Wait()
	assert.Equal(t, "notified", result)
}

func TestNotifyBroadcast(t *testing.T) {
	cv, broadcast := NotifyBroadcast()
	assert.NotNil(t, cv)
	assert.NotNil(t, broadcast)

	var results []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start multiple goroutines waiting
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cv.Wait()
			mu.Lock()
			results = append(results, fmt.Sprintf("broadcast-%d", id))
			mu.Unlock()
		}(i)
	}

	// Give goroutines time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Broadcast using the convenience function
	broadcast()

	wg.Wait()
	assert.Len(t, results, 3)
}

func TestCondVar_RaceCondition(t *testing.T) {
	cv := NewCondVar()
	var wg sync.WaitGroup

	// Test concurrent Clone operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clone := cv.Clone()
			clone.Drop()
		}()
	}

	// Test concurrent Signal operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cv.Signal()
		}()
	}

	// Test concurrent Broadcast operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cv.Broadcast()
		}()
	}

	wg.Wait()
	// Should not panic or cause race conditions
}

func TestCondVar_DropWithWaitingGoroutines(t *testing.T) {
	cv := NewCondVar()
	var wg sync.WaitGroup

	// Start goroutines waiting
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cv.Wait()
		}()
	}

	// Give goroutines time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Drop the last reference, which should wake up all waiting goroutines
	cv.Drop()

	wg.Wait()
	// Should not hang
}

func TestCondVar_RefCountUnderflow(t *testing.T) {
	cv := NewCondVar()

	// Drop more times than we have references
	cv.Drop() // Should go to 0
	cv.Drop() // Should not go negative

	assert.Equal(t, int64(0), cv.RefCount())
}

func TestCondVar_ConcurrentWaitAndSignal(t *testing.T) {
	cv := NewCondVar()
	var wg sync.WaitGroup
	var results []string
	var mu sync.Mutex

	// Start multiple goroutines waiting
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cv.Wait()
			mu.Lock()
			results = append(results, fmt.Sprintf("waiter-%d", id))
			mu.Unlock()
		}(i)
	}

	// Give goroutines time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Signal multiple times without delays
	for i := 0; i < 5; i++ {
		cv.Signal()
	}

	wg.Wait()
	assert.Len(t, results, 5)
}

func TestCondVar_StressTest(t *testing.T) {
	cv := NewCondVar()
	var wg sync.WaitGroup
	var counter int64
	var mu sync.Mutex

	// Start many goroutines that wait and increment counter
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cv.Wait()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	// Give goroutines time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Broadcast to wake all goroutines
	cv.Broadcast()

	wg.Wait()
	assert.Equal(t, int64(50), counter)
}
