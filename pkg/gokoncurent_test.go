package gokoncurent

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestGetInfo(t *testing.T) {
	info := GetInfo()
	if info == nil {
		t.Fatal("GetInfo() should not return nil")
	}

	if info.Version != Version {
		t.Errorf("Expected version %s, got %s", Version, info.Version)
	}

	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
}

func TestInfoString(t *testing.T) {
	info := GetInfo()
	str := info.String()

	if str == "" {
		t.Error("String() should not return empty string")
	}

	// Should contain version information
	if !contains(str, "Version:") {
		t.Error("String() should contain version information")
	}

	if !contains(str, "Go Version:") {
		t.Error("String() should contain Go version information")
	}
}

func TestInfoConcurrentAccess(t *testing.T) {
	// Test that GetInfo is safe for concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			info := GetInfo()
			if info == nil {
				t.Error("GetInfo() returned nil in goroutine")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestRWArcMutex tests the RWArcMutex functionality
func TestRWArcMutex(t *testing.T) {
	// Test basic creation and access
	rw := NewRWArcMutex(42)
	require.NotNil(t, rw)
	require.Equal(t, int64(1), rw.RefCount())

	// Test read access
	rw.WithRLock(func(v *int) {
		require.Equal(t, 42, *v)
	})

	// Test write access
	rw.WithLock(func(v *int) {
		*v = 100
	})

	// Verify write took effect
	rw.WithRLock(func(v *int) {
		require.Equal(t, 100, *v)
	})

	// Test cloning
	clone := rw.Clone()
	require.Equal(t, int64(2), rw.RefCount())

	// Test concurrent access
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			rw.WithRLock(func(v *int) {
				require.GreaterOrEqual(t, *v, 100)
			})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			rw.WithLock(func(v *int) {
				*v += 1
			})
		}
	}()

	wg.Wait()

	// Clean up
	rw.Drop()
	clone.Drop()
	require.Equal(t, int64(0), rw.RefCount())
}

// TestCondVar tests the CondVar functionality
func TestCondVar(t *testing.T) {
	// Test basic creation
	cv := NewCondVar()
	require.NotNil(t, cv)
	require.Equal(t, int64(1), cv.RefCount())

	// Test cloning
	clone := cv.Clone()
	require.Equal(t, int64(2), cv.RefCount())

	// Test basic signal/wait
	var result string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		cv.Wait()
		result = "signaled"
	}()

	time.Sleep(10 * time.Millisecond)
	cv.Signal()
	wg.Wait()
	require.Equal(t, "signaled", result)

	// Test broadcast
	var results []string
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cv.Wait()
			mu.Lock()
			results = append(results, fmt.Sprintf("worker-%d", id))
			mu.Unlock()
		}(i)
	}

	time.Sleep(10 * time.Millisecond)
	cv.Broadcast()
	wg.Wait()
	require.Len(t, results, 3)

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	var contextResult bool

	wg.Add(1)
	go func() {
		defer wg.Done()
		contextResult = cv.WaitWithContext(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()
	wg.Wait()
	require.False(t, contextResult)

	// Test timeout
	timeoutResult := cv.WaitWithTimeout(50 * time.Millisecond)
	require.False(t, timeoutResult)

	// Clean up
	cv.Drop()
	clone.Drop()
	require.Equal(t, int64(0), cv.RefCount())
}

// TestBarrier tests the Barrier functionality
func TestBarrier(t *testing.T) {
	// Test basic creation
	b := NewBarrier(3)
	require.NotNil(t, b)
	require.Equal(t, int64(1), b.RefCount())

	// Test cloning
	clone := b.Clone()
	require.Equal(t, int64(2), b.RefCount())

	// Test basic synchronization
	var results []bool
	var mu sync.Mutex
	var wg sync.WaitGroup

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
	require.Len(t, results, 3)
	for _, result := range results {
		require.True(t, result)
	}

	// Test reset functionality
	b.Reset(2)
	var results2 []bool

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := b.Wait()
			mu.Lock()
			results2 = append(results2, result)
			mu.Unlock()
		}()
	}

	wg.Wait()
	require.Len(t, results2, 2)
	for _, result := range results2 {
		require.True(t, result)
	}

	// Clean up
	b.Drop()
	clone.Drop()
	require.Equal(t, int64(0), b.RefCount())
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

// Benchmark tests
func BenchmarkGetInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetInfo()
	}
}

func BenchmarkInfoString(b *testing.B) {
	info := GetInfo()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = info.String()
	}
}

func BenchmarkGetInfoConcurrent(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GetInfo()
		}
	})
}
