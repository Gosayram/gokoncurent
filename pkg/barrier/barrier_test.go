package barrier

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBarrier(t *testing.T) {
	// Test valid creation
	b := NewBarrier(3)
	assert.NotNil(t, b)
	assert.Equal(t, int64(1), b.RefCount())
	assert.Equal(t, "Barrier{count=3, waiting=0, refCount=1, broken=false}", b.String())

	// Test invalid creation
	assert.Panics(t, func() {
		NewBarrier(0)
	})

	assert.Panics(t, func() {
		NewBarrier(-1)
	})
}

func TestBarrier_Clone(t *testing.T) {
	b := NewBarrier(2)
	initialCount := b.RefCount()

	clone := b.Clone()
	assert.Equal(t, b, clone)
	assert.Equal(t, initialCount+1, b.RefCount())
	assert.Equal(t, initialCount+1, clone.RefCount())
}

func TestBarrier_Drop(t *testing.T) {
	b := NewBarrier(2)
	clone := b.Clone()

	// Drop clone first
	clone.Drop()
	assert.Equal(t, int64(1), b.RefCount())

	// Drop original
	b.Drop()
	assert.Equal(t, int64(0), b.RefCount())
}

func TestBarrier_Wait(t *testing.T) {
	b := NewBarrier(3)
	var results []bool
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start 3 goroutines that will wait
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := b.Wait()
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()
	}

	wg.Wait()
	assert.Len(t, results, 3)
	// All should return true since barrier was successfully crossed
	for _, result := range results {
		assert.True(t, result)
	}
}

func TestBarrier_Wait_Uneven(t *testing.T) {
	b := NewBarrier(3)
	var results []bool
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start only 2 goroutines (less than required 3)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := b.Wait()
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()
	}

	// Wait a bit to see if they're blocked
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, results, 0) // Should still be waiting

	// Start the third goroutine to complete the barrier
	wg.Add(1)
	go func() {
		defer wg.Done()
		result := b.Wait()
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	}()

	wg.Wait()
	assert.Len(t, results, 3)
	for _, result := range results {
		assert.True(t, result)
	}
}

func TestBarrier_Wait_Broken(t *testing.T) {
	b := NewBarrier(3)
	var results []bool
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start 2 goroutines that will wait
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := b.Wait()
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()
	}

	// Give them time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Break the barrier by dropping the last reference
	b.Drop()

	wg.Wait()
	assert.Len(t, results, 2)
	// All should return false since barrier was broken
	for _, result := range results {
		assert.False(t, result)
	}
}

func TestBarrier_Reset(t *testing.T) {
	b := NewBarrier(3)
	var wg sync.WaitGroup

	// Complete the barrier first
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Wait()
		}()
	}
	wg.Wait()

	// Now reset to 2 participants
	b.Reset(2)

	// Test with new count
	var results []bool
	var mu sync.Mutex

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := b.Wait()
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()
	}

	wg.Wait()
	assert.Len(t, results, 2)
	for _, result := range results {
		assert.True(t, result)
	}
}

func TestBarrier_Reset_WhileWaiting(t *testing.T) {
	b := NewBarrier(3)

	// Start a goroutine that will wait
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.Wait()
	}()

	// Give it time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Try to reset while someone is waiting - should panic
	assert.Panics(t, func() {
		b.Reset(2)
	})

	// Clean up
	b.Drop()
	wg.Wait()
}

func TestBarrier_String(t *testing.T) {
	b := NewBarrier(5)
	str := b.String()
	assert.Contains(t, str, "Barrier{count=5, waiting=0, refCount=1, broken=false}")

	b.Clone()
	str = b.String()
	assert.Contains(t, str, "refCount=2")
}

func TestBarrier_RaceCondition(t *testing.T) {
	b := NewBarrier(10)
	var wg sync.WaitGroup

	// Test concurrent Clone operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clone := b.Clone()
			clone.Drop()
		}()
	}

	// Test concurrent Wait operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Wait()
		}()
	}

	wg.Wait()
	// Should not panic or cause race conditions
}

func TestBarrier_RefCountUnderflow(t *testing.T) {
	b := NewBarrier(1)

	// Drop more times than we have references
	b.Drop() // Should go to 0
	b.Drop() // Should not go negative

	assert.Equal(t, int64(0), b.RefCount())
}

func TestBarrier_StressTest(t *testing.T) {
	b := NewBarrier(20)
	var wg sync.WaitGroup
	var counter int64
	var mu sync.Mutex

	// Start many goroutines that wait and increment counter
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := b.Wait()
			if result {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(20), counter)
}

func TestBarrier_MultipleCycles(t *testing.T) {
	b := NewBarrier(3)
	var wg sync.WaitGroup

	// Run multiple cycles
	for cycle := 0; cycle < 3; cycle++ {
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				b.Wait()
			}()
		}
		wg.Wait()
	}
}

func TestBarrier_ConcurrentReset(t *testing.T) {
	b := NewBarrier(3)
	var wg sync.WaitGroup

	// Complete the barrier first
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Wait()
		}()
	}
	wg.Wait()

	// Test concurrent reset operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Reset(3)
		}()
	}

	wg.Wait()
	// Should not panic
}
