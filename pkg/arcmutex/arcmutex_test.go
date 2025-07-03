package arcmutex

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewArcMutex(t *testing.T) {
	t.Run("int value", func(t *testing.T) {
		am := NewArcMutex(42)
		if am == nil {
			t.Fatal("NewArcMutex should not return nil")
		}
		if !am.IsValid() {
			t.Fatal("NewArcMutex should create a valid instance")
		}
		if am.RefCount() != 1 {
			t.Errorf("Expected reference count 1, got %d", am.RefCount())
		}
	})

	t.Run("string value", func(t *testing.T) {
		am := NewArcMutex("hello")
		var result string
		am.WithLock(func(s *string) {
			result = *s
		})
		if result != "hello" {
			t.Errorf("Expected 'hello', got '%s'", result)
		}
	})

	t.Run("struct value", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		original := TestStruct{Name: "Alice", Age: 30}
		am := NewArcMutex(original)

		var result TestStruct
		am.WithLock(func(s *TestStruct) {
			result = *s
		})

		if result.Name != "Alice" || result.Age != 30 {
			t.Errorf("Expected {Alice 30}, got %+v", result)
		}
	})
}

func TestArcMutexClone(t *testing.T) {
	original := NewArcMutex(100)

	t.Run("clone increments reference count", func(t *testing.T) {
		clone := original.Clone()
		if clone == nil {
			t.Fatal("Clone should not return nil")
		}
		if original.RefCount() != 2 {
			t.Errorf("Expected reference count 2, got %d", original.RefCount())
		}
		if clone.RefCount() != 2 {
			t.Errorf("Expected reference count 2, got %d", clone.RefCount())
		}
	})

	t.Run("clone shares same data", func(t *testing.T) {
		clone := original.Clone()

		// Modify through original
		original.WithLock(func(value *int) {
			*value = 200
		})

		// Check through clone
		var cloneValue int
		clone.WithLock(func(value *int) {
			cloneValue = *value
		})

		if cloneValue != 200 {
			t.Errorf("Expected clone to see modified value 200, got %d", cloneValue)
		}
	})

	t.Run("clone of nil returns nil", func(t *testing.T) {
		var nilArcMutex *ArcMutex[int]
		clone := nilArcMutex.Clone()
		if clone != nil {
			t.Error("Clone of nil should return nil")
		}
	})
}

func TestArcMutexWithLock(t *testing.T) {
	t.Run("basic modification", func(t *testing.T) {
		am := NewArcMutex(0)

		am.WithLock(func(value *int) {
			*value = 42
		})

		var result int
		am.WithLock(func(value *int) {
			result = *value
		})

		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("multiple modifications", func(t *testing.T) {
		am := NewArcMutex(0)

		// Increment 5 times
		for i := 0; i < 5; i++ {
			am.WithLock(func(value *int) {
				*value++
			})
		}

		var result int
		am.WithLock(func(value *int) {
			result = *value
		})

		if result != 5 {
			t.Errorf("Expected 5, got %d", result)
		}
	})

	t.Run("nil function", func(t *testing.T) {
		am := NewArcMutex(42)
		// Should not panic
		am.WithLock(nil)
	})

	t.Run("nil ArcMutex", func(t *testing.T) {
		var nilAM *ArcMutex[int]
		// Should not panic
		nilAM.WithLock(func(value *int) {
			*value = 100
		})
	})
}

func TestArcMutexTryWithLock(t *testing.T) {
	t.Run("successful lock", func(t *testing.T) {
		am := NewArcMutex(0)

		success := am.TryWithLock(func(value *int) {
			*value = 42
		})

		if !success {
			t.Error("TryWithLock should succeed when mutex is available")
		}

		var result int
		am.WithLock(func(value *int) {
			result = *value
		})

		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("contention", func(t *testing.T) {
		am := NewArcMutex(0)

		// Hold lock in goroutine
		lockHeld := make(chan struct{})
		canRelease := make(chan struct{})

		go func() {
			am.WithLock(func(value *int) {
				close(lockHeld)
				<-canRelease
			})
		}()

		// Wait for lock to be held
		<-lockHeld

		// Try to acquire lock (should fail)
		success := am.TryWithLock(func(value *int) {
			*value = 999
		})

		if success {
			t.Error("TryWithLock should fail when mutex is held")
		}

		// Release the lock
		close(canRelease)
	})

	t.Run("nil function", func(t *testing.T) {
		am := NewArcMutex(42)
		success := am.TryWithLock(nil)
		if success {
			t.Error("TryWithLock should return false for nil function")
		}
	})

	t.Run("nil ArcMutex", func(t *testing.T) {
		var nilAM *ArcMutex[int]
		success := nilAM.TryWithLock(func(value *int) {
			*value = 100
		})
		if success {
			t.Error("TryWithLock should return false for nil ArcMutex")
		}
	})
}

func TestArcMutexWithLockResult(t *testing.T) {
	t.Run("return value", func(t *testing.T) {
		am := NewArcMutex(42)

		result := am.WithLockResult(func(value *int) interface{} {
			return *value * 2
		})

		if result != 84 {
			t.Errorf("Expected 84, got %v", result)
		}
	})

	t.Run("return string", func(t *testing.T) {
		am := NewArcMutex("hello")

		result := am.WithLockResult(func(value *string) interface{} {
			return *value + " world"
		})

		if result != "hello world" {
			t.Errorf("Expected 'hello world', got %v", result)
		}
	})

	t.Run("nil function", func(t *testing.T) {
		am := NewArcMutex(42)
		result := am.WithLockResult(nil)
		if result != nil {
			t.Error("WithLockResult should return nil for nil function")
		}
	})

	t.Run("nil ArcMutex", func(t *testing.T) {
		var nilAM *ArcMutex[int]
		result := nilAM.WithLockResult(func(value *int) interface{} {
			return *value
		})
		if result != nil {
			t.Error("WithLockResult should return nil for nil ArcMutex")
		}
	})
}

func TestArcMutexConcurrency(t *testing.T) {
	t.Run("concurrent increments", func(t *testing.T) {
		am := NewArcMutex(0)
		const numGoroutines = 100
		const numIncrements = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < numIncrements; j++ {
					am.WithLock(func(value *int) {
						*value++
					})
				}
			}()
		}

		wg.Wait()

		var result int
		am.WithLock(func(value *int) {
			result = *value
		})

		expected := numGoroutines * numIncrements
		if result != expected {
			t.Errorf("Expected %d, got %d", expected, result)
		}
	})

	t.Run("concurrent cloning and modification", func(t *testing.T) {
		am := NewArcMutex(0)
		const numGoroutines = 50
		var wg sync.WaitGroup
		var counter int64

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				clone := am.Clone()
				if clone == nil {
					t.Error("Clone should not be nil")
					return
				}

				// Increment through clone
				clone.WithLock(func(value *int) {
					*value++
				})

				atomic.AddInt64(&counter, 1)
			}()
		}

		wg.Wait()

		var result int
		am.WithLock(func(value *int) {
			result = *value
		})

		if result != numGoroutines {
			t.Errorf("Expected %d, got %d", numGoroutines, result)
		}

		if counter != numGoroutines {
			t.Errorf("Expected counter %d, got %d", numGoroutines, counter)
		}
	})
}

func TestArcMutexThreadSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping thread safety test in short mode")
	}

	t.Run("stress test", func(t *testing.T) {
		am := NewArcMutex(0)
		const numGoroutines = 20
		const testDuration = 100 * time.Millisecond

		var wg sync.WaitGroup
		stop := make(chan struct{})

		// Start incrementing goroutines
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stop:
						return
					default:
						am.WithLock(func(value *int) {
							*value++
							// Small delay to increase contention
							runtime.Gosched()
						})
					}
				}
			}()
		}

		// Let it run for a while
		time.Sleep(testDuration)
		close(stop)
		wg.Wait()

		// Check that the value is consistent
		var result int
		am.WithLock(func(value *int) {
			result = *value
		})

		if result <= 0 {
			t.Error("Expected positive result after stress test")
		}

		// Reference count should be 1 (only original remains)
		if am.RefCount() != 1 {
			t.Errorf("Expected reference count 1, got %d", am.RefCount())
		}
	})
}

func TestArcMutexEdgeCases(t *testing.T) {
	t.Run("drop and access", func(t *testing.T) {
		am := NewArcMutex(42)
		clone := am.Clone()

		// Drop original
		am.Drop()

		// Clone should still work
		var result int
		clone.WithLock(func(value *int) {
			result = *value
		})

		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("zero value", func(t *testing.T) {
		am := NewArcMutex(0)
		var result int
		am.WithLock(func(value *int) {
			result = *value
		})
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})
}

// Benchmark tests
func BenchmarkArcMutexNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am := NewArcMutex(i)
		_ = am
	}
}

func BenchmarkArcMutexClone(b *testing.B) {
	am := NewArcMutex(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clone := am.Clone()
		clone.Drop()
	}
}

func BenchmarkArcMutexWithLock(b *testing.B) {
	am := NewArcMutex(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am.WithLock(func(value *int) {
			*value++
		})
	}
}

func BenchmarkArcMutexTryWithLock(b *testing.B) {
	am := NewArcMutex(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am.TryWithLock(func(value *int) {
			*value++
		})
	}
}

func BenchmarkArcMutexConcurrent(b *testing.B) {
	am := NewArcMutex(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			am.WithLock(func(value *int) {
				*value++
			})
		}
	})
}

// Example tests for documentation
func ExampleNewArcMutex() {
	counter := NewArcMutex(0)
	counter.WithLock(func(value *int) {
		*value = 42
	})

	result := counter.WithLockResult(func(value *int) interface{} {
		return *value
	})

	fmt.Println(result)
	// Output: 42
}

func ExampleArcMutex_Clone() {
	original := NewArcMutex(100)
	clone := original.Clone()

	// Modify through original
	original.WithLock(func(value *int) {
		*value += 50
	})

	// Read through clone
	result := clone.WithLockResult(func(value *int) interface{} {
		return *value
	})

	fmt.Println(result)
	// Output: 150
}

func ExampleArcMutex_WithLock() {
	am := NewArcMutex("hello")

	am.WithLock(func(value *string) {
		*value += " world"
	})

	result := am.WithLockResult(func(value *string) interface{} {
		return *value
	})

	fmt.Println(result)
	// Output: hello world
}
