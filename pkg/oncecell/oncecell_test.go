package oncecell

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewOnceCell(t *testing.T) {
	t.Run("string cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		if cell == nil {
			t.Fatal("NewOnceCell should not return nil")
		}
		if cell.IsInitialized() {
			t.Error("New OnceCell should not be initialized")
		}
	})

	t.Run("int cell", func(t *testing.T) {
		cell := NewOnceCell[int]()
		if cell == nil {
			t.Fatal("NewOnceCell should not return nil")
		}
		if cell.IsInitialized() {
			t.Error("New OnceCell should not be initialized")
		}
	})

	t.Run("struct cell", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		cell := NewOnceCell[TestStruct]()
		if cell == nil {
			t.Fatal("NewOnceCell should not return nil")
		}
		if cell.IsInitialized() {
			t.Error("New OnceCell should not be initialized")
		}
	})
}

func TestOnceCellSet(t *testing.T) {
	t.Run("first set succeeds", func(t *testing.T) {
		cell := NewOnceCell[string]()
		success := cell.Set("hello")
		if !success {
			t.Error("First Set should return true")
		}
		if !cell.IsInitialized() {
			t.Error("Cell should be initialized after Set")
		}
	})

	t.Run("second set fails", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("first")
		success := cell.Set("second")
		if success {
			t.Error("Second Set should return false")
		}

		value, ok := cell.Get()
		if !ok || value != "first" {
			t.Error("Cell should contain first value")
		}
	})

	t.Run("concurrent sets", func(t *testing.T) {
		cell := NewOnceCell[int]()
		const numGoroutines = 100
		var wg sync.WaitGroup
		var successCount int64

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(value int) {
				defer wg.Done()
				if cell.Set(value) {
					atomic.AddInt64(&successCount, 1)
				}
			}(i)
		}

		wg.Wait()

		if successCount != 1 {
			t.Errorf("Expected exactly 1 successful set, got %d", successCount)
		}

		if !cell.IsInitialized() {
			t.Error("Cell should be initialized after concurrent sets")
		}
	})

	t.Run("nil cell", func(t *testing.T) {
		var nilCell *OnceCell[string]
		success := nilCell.Set("test")
		if success {
			t.Error("Set on nil cell should return false")
		}
	})
}

func TestOnceCellGet(t *testing.T) {
	t.Run("empty cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		value, ok := cell.Get()
		if ok {
			t.Error("Get on empty cell should return false")
		}
		if value != "" {
			t.Error("Get on empty cell should return zero value")
		}
	})

	t.Run("initialized cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("hello")
		value, ok := cell.Get()
		if !ok {
			t.Error("Get on initialized cell should return true")
		}
		if value != "hello" {
			t.Errorf("Expected 'hello', got '%s'", value)
		}
	})

	t.Run("int cell", func(t *testing.T) {
		cell := NewOnceCell[int]()
		cell.Set(42)
		value, ok := cell.Get()
		if !ok {
			t.Error("Get on initialized cell should return true")
		}
		if value != 42 {
			t.Errorf("Expected 42, got %d", value)
		}
	})

	t.Run("struct cell", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		cell := NewOnceCell[TestStruct]()
		original := TestStruct{Name: "Alice", Age: 30}
		cell.Set(original)

		value, ok := cell.Get()
		if !ok {
			t.Error("Get on initialized cell should return true")
		}
		if value.Name != "Alice" || value.Age != 30 {
			t.Errorf("Expected {Alice 30}, got %+v", value)
		}
	})

	t.Run("nil cell", func(t *testing.T) {
		var nilCell *OnceCell[string]
		value, ok := nilCell.Get()
		if ok {
			t.Error("Get on nil cell should return false")
		}
		if value != "" {
			t.Error("Get on nil cell should return zero value")
		}
	})
}

func TestOnceCellGetOrInit(t *testing.T) {
	t.Run("initialize with function", func(t *testing.T) {
		cell := NewOnceCell[string]()
		called := false

		value := cell.GetOrInit(func() string {
			called = true
			return "initialized"
		})

		if !called {
			t.Error("Initialization function should be called")
		}
		if value != "initialized" {
			t.Errorf("Expected 'initialized', got '%s'", value)
		}
		if !cell.IsInitialized() {
			t.Error("Cell should be initialized after GetOrInit")
		}
	})

	t.Run("already initialized", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("existing")

		called := false
		value := cell.GetOrInit(func() string {
			called = true
			return "new"
		})

		if called {
			t.Error("Initialization function should not be called for initialized cell")
		}
		if value != "existing" {
			t.Errorf("Expected 'existing', got '%s'", value)
		}
	})

	t.Run("concurrent initialization", func(t *testing.T) {
		cell := NewOnceCell[int]()
		const numGoroutines = 100
		var wg sync.WaitGroup
		var callCount int64

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				value := cell.GetOrInit(func() int {
					atomic.AddInt64(&callCount, 1)
					return id
				})
				_ = value
			}(i)
		}

		wg.Wait()

		if callCount != 1 {
			t.Errorf("Expected exactly 1 initialization call, got %d", callCount)
		}

		if !cell.IsInitialized() {
			t.Error("Cell should be initialized after concurrent GetOrInit")
		}
	})

	t.Run("nil cell", func(t *testing.T) {
		var nilCell *OnceCell[string]
		value := nilCell.GetOrInit(func() string {
			return "test"
		})
		if value != "" {
			t.Error("GetOrInit on nil cell should return zero value")
		}
	})
}

func TestOnceCellGetOrInitWith(t *testing.T) {
	t.Run("initialize with value", func(t *testing.T) {
		cell := NewOnceCell[string]()
		value := cell.GetOrInitWith("hello")

		if value != "hello" {
			t.Errorf("Expected 'hello', got '%s'", value)
		}
		if !cell.IsInitialized() {
			t.Error("Cell should be initialized after GetOrInitWith")
		}
	})

	t.Run("already initialized", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("existing")

		value := cell.GetOrInitWith("new")
		if value != "existing" {
			t.Errorf("Expected 'existing', got '%s'", value)
		}
	})

	t.Run("int value", func(t *testing.T) {
		cell := NewOnceCell[int]()
		value := cell.GetOrInitWith(42)

		if value != 42 {
			t.Errorf("Expected 42, got %d", value)
		}
	})
}

func TestOnceCellIsInitialized(t *testing.T) {
	t.Run("empty cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		if cell.IsInitialized() {
			t.Error("Empty cell should not be initialized")
		}
	})

	t.Run("set cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("hello")
		if !cell.IsInitialized() {
			t.Error("Set cell should be initialized")
		}
	})

	t.Run("GetOrInit cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.GetOrInit(func() string { return "hello" })
		if !cell.IsInitialized() {
			t.Error("GetOrInit cell should be initialized")
		}
	})

	t.Run("nil cell", func(t *testing.T) {
		var nilCell *OnceCell[string]
		if nilCell.IsInitialized() {
			t.Error("Nil cell should not be initialized")
		}
	})
}

func TestOnceCellTryGet(t *testing.T) {
	t.Run("empty cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		value, ok := cell.TryGet()
		if ok {
			t.Error("TryGet on empty cell should return false")
		}
		if value != "" {
			t.Error("TryGet on empty cell should return zero value")
		}
	})

	t.Run("initialized cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("hello")
		value, ok := cell.TryGet()
		if !ok {
			t.Error("TryGet on initialized cell should return true")
		}
		if value != "hello" {
			t.Errorf("Expected 'hello', got '%s'", value)
		}
	})
}

func TestOnceCellReset(t *testing.T) {
	t.Run("reset creates new cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("hello")

		newCell := cell.Reset()
		if newCell == nil {
			t.Fatal("Reset should not return nil")
		}
		if newCell == cell {
			t.Error("Reset should return a different cell")
		}
		if newCell.IsInitialized() {
			t.Error("Reset cell should not be initialized")
		}

		// Original cell should still be initialized
		if !cell.IsInitialized() {
			t.Error("Original cell should still be initialized")
		}
	})

	t.Run("reset empty cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		newCell := cell.Reset()

		if newCell == nil {
			t.Fatal("Reset should not return nil")
		}
		if newCell.IsInitialized() {
			t.Error("Reset cell should not be initialized")
		}
	})
}

func TestOnceCellConcurrency(t *testing.T) {
	t.Run("concurrent set and get", func(t *testing.T) {
		cell := NewOnceCell[int]()
		const numGoroutines = 100
		var wg sync.WaitGroup
		var successfulSets int64
		var successfulGets int64

		wg.Add(numGoroutines * 2)

		// Setters
		for i := 0; i < numGoroutines; i++ {
			go func(value int) {
				defer wg.Done()
				if cell.Set(value) {
					atomic.AddInt64(&successfulSets, 1)
				}
			}(i)
		}

		// Getters
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				// Small delay to let some sets happen
				time.Sleep(time.Microsecond)
				if _, ok := cell.Get(); ok {
					atomic.AddInt64(&successfulGets, 1)
				}
			}()
		}

		wg.Wait()

		if successfulSets != 1 {
			t.Errorf("Expected exactly 1 successful set, got %d", successfulSets)
		}

		// All gets after the set should succeed
		if successfulGets == 0 {
			t.Error("Expected at least some successful gets")
		}
	})

	t.Run("concurrent GetOrInit", func(t *testing.T) {
		cell := NewOnceCell[string]()
		const numGoroutines = 100
		var wg sync.WaitGroup
		var initCount int64
		results := make([]string, numGoroutines)

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				defer wg.Done()
				value := cell.GetOrInit(func() string {
					atomic.AddInt64(&initCount, 1)
					return fmt.Sprintf("value_%d", index)
				})
				results[index] = value
			}(i)
		}

		wg.Wait()

		if initCount != 1 {
			t.Errorf("Expected exactly 1 initialization, got %d", initCount)
		}

		// All results should be the same
		firstResult := results[0]
		for i, result := range results {
			if result != firstResult {
				t.Errorf("Result %d (%s) differs from first result (%s)", i, result, firstResult)
			}
		}
	})
}

func TestOnceCellPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("read performance after initialization", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("performance test")

		const numReads = 1000000
		start := time.Now()

		for i := 0; i < numReads; i++ {
			_, _ = cell.Get()
		}

		duration := time.Since(start)
		readsPerSecond := float64(numReads) / duration.Seconds()

		// Should be very fast (millions of reads per second)
		if readsPerSecond < 1000000 {
			t.Logf("Read performance: %.0f reads/second", readsPerSecond)
		}
	})
}

// Benchmark tests
func BenchmarkOnceCellNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cell := NewOnceCell[int]()
		_ = cell
	}
}

func BenchmarkOnceCellSet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cell := NewOnceCell[int]()
		cell.Set(i)
	}
}

func BenchmarkOnceCellGet(b *testing.B) {
	cell := NewOnceCell[int]()
	cell.Set(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cell.Get()
	}
}

func BenchmarkOnceCellGetOrInit(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cell := NewOnceCell[int]()
		cell.GetOrInit(func() int { return i })
	}
}

func BenchmarkOnceCellConcurrentGet(b *testing.B) {
	cell := NewOnceCell[int]()
	cell.Set(42)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cell.Get()
		}
	})
}

func BenchmarkOnceCellConcurrentGetOrInit(b *testing.B) {
	cell := NewOnceCell[string]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cell.GetOrInit(func() string {
				return "concurrent value"
			})
		}
	})
}

func TestOnceCellResetWithCallback(t *testing.T) {
	t.Run("callback with initialized cell", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("old value")

		var callbackCalled bool
		var callbackValue string

		newCell := cell.ResetWithCallback(func(value string) {
			callbackCalled = true
			callbackValue = value
		})

		if !callbackCalled {
			t.Error("Callback should be called for initialized cell")
		}
		if callbackValue != "old value" {
			t.Errorf("Expected callback value 'old value', got '%s'", callbackValue)
		}
		if newCell == nil {
			t.Error("ResetWithCallback should return new cell")
		}
		if newCell.IsInitialized() {
			t.Error("New cell should not be initialized")
		}
		// Original cell should remain initialized (ResetWithCallback creates a new cell)
		if !cell.IsInitialized() {
			t.Error("Original cell should remain initialized")
		}
	})

	t.Run("callback with uninitialized cell", func(t *testing.T) {
		cell := NewOnceCell[string]()

		var callbackCalled bool

		newCell := cell.ResetWithCallback(func(value string) {
			callbackCalled = true
		})

		if callbackCalled {
			t.Error("Callback should not be called for uninitialized cell")
		}
		if newCell == nil {
			t.Error("ResetWithCallback should return new cell")
		}
		if newCell.IsInitialized() {
			t.Error("New cell should not be initialized")
		}
	})

	t.Run("nil cell", func(t *testing.T) {
		var nilCell *OnceCell[string]
		var callbackCalled bool

		newCell := nilCell.ResetWithCallback(func(value string) {
			callbackCalled = true
		})

		if callbackCalled {
			t.Error("Callback should not be called for nil cell")
		}
		if newCell == nil {
			t.Error("ResetWithCallback should return new cell even for nil input")
		}
	})
}

func TestOnceCellGetOrInitWithRetry(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		cell := NewOnceCell[string]()

		value, err := cell.GetOrInitWithRetry(func() (string, error) {
			return "success", nil
		}, 3, 10*time.Millisecond)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if value != "success" {
			t.Errorf("Expected 'success', got '%s'", value)
		}
		if !cell.IsInitialized() {
			t.Error("Cell should be initialized after successful GetOrInitWithRetry")
		}
	})

	t.Run("retry with eventual success", func(t *testing.T) {
		cell := NewOnceCell[string]()
		attempts := 0

		value, err := cell.GetOrInitWithRetry(func() (string, error) {
			attempts++
			if attempts < 3 {
				return "", fmt.Errorf("temporary failure %d", attempts)
			}
			return "success after retry", nil
		}, 3, 10*time.Millisecond)

		if err != nil {
			t.Errorf("Expected no error after retry, got %v", err)
		}
		if value != "success after retry" {
			t.Errorf("Expected 'success after retry', got '%s'", value)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("all retries fail", func(t *testing.T) {
		cell := NewOnceCell[string]()

		value, err := cell.GetOrInitWithRetry(func() (string, error) {
			return "", fmt.Errorf("permanent failure")
		}, 2, 10*time.Millisecond)

		if err == nil {
			t.Fatal("Expected error after all retries failed")
		}
		if err.Error() != "permanent failure" {
			t.Errorf("Expected 'permanent failure', got '%v'", err)
		}
		// Value should be empty (zero value) after failed retries
		if value != "" {
			t.Errorf("Expected empty value, got '%s'", value)
		}
		// Cell should be initialized even after failed retries
		if !cell.IsInitialized() {
			t.Error("Cell should be initialized even after failed retries")
		}
	})

	t.Run("zero retries", func(t *testing.T) {
		cell := NewOnceCell[string]()

		value, err := cell.GetOrInitWithRetry(func() (string, error) {
			return "", fmt.Errorf("failure")
		}, 0, 10*time.Millisecond)

		if err == nil {
			t.Fatal("Expected error with zero retries")
		}
		if value != "" {
			t.Errorf("Expected empty value, got '%s'", value)
		}
	})

	t.Run("already initialized", func(t *testing.T) {
		cell := NewOnceCell[string]()
		cell.Set("existing")

		var initCalled bool
		value, err := cell.GetOrInitWithRetry(func() (string, error) {
			initCalled = true
			return "new", nil
		}, 3, 10*time.Millisecond)

		if initCalled {
			t.Error("Init function should not be called for already initialized cell")
		}
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if value != "existing" {
			t.Errorf("Expected 'existing', got '%s'", value)
		}
	})

	t.Run("nil cell", func(t *testing.T) {
		var nilCell *OnceCell[string]

		value, err := nilCell.GetOrInitWithRetry(func() (string, error) {
			return "test", nil
		}, 3, 10*time.Millisecond)

		if err != nil {
			t.Errorf("Expected no error for nil cell, got %v", err)
		}
		if value != "" {
			t.Errorf("Expected empty value for nil cell, got '%s'", value)
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		cell := NewOnceCell[string]()
		const numGoroutines = 10
		var wg sync.WaitGroup
		results := make([]string, numGoroutines)
		errors := make([]error, numGoroutines)

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				value, err := cell.GetOrInitWithRetry(func() (string, error) {
					time.Sleep(1 * time.Millisecond) // Simulate work
					return fmt.Sprintf("value-%d", id), nil
				}, 1, 5*time.Millisecond)
				results[id] = value
				errors[id] = err
			}(i)
		}

		wg.Wait()

		// All should get the same value (first one to succeed)
		firstValue := results[0]
		for i := 1; i < numGoroutines; i++ {
			if results[i] != firstValue {
				t.Errorf("Expected all goroutines to get same value, got %s and %s", firstValue, results[i])
			}
			if errors[i] != nil {
				t.Errorf("Expected no error for goroutine %d, got %v", i, errors[i])
			}
		}
	})
}

func BenchmarkOnceCellResetWithCallback(b *testing.B) {
	cell := NewOnceCell[string]()
	cell.Set("test value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cell.ResetWithCallback(func(value string) {
			// Callback operation
		})
	}
}

func BenchmarkOnceCellGetOrInitWithRetry(b *testing.B) {
	cell := NewOnceCell[string]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := cell.GetOrInitWithRetry(func() (string, error) {
			return "success", nil
		}, 0, 1*time.Millisecond); err != nil {
			b.Fatal(err)
		}
	}
}

// Example tests for documentation
func ExampleNewOnceCell() {
	cell := NewOnceCell[string]()
	cell.Set("Hello, World!")

	value, ok := cell.Get()
	if ok {
		fmt.Println(value)
	}
	// Output: Hello, World!
}

func ExampleOnceCell_Set() {
	cell := NewOnceCell[int]()

	success1 := cell.Set(42)
	success2 := cell.Set(100)

	fmt.Println("First set:", success1)
	fmt.Println("Second set:", success2)

	value, _ := cell.Get()
	fmt.Println("Value:", value)

	// Output:
	// First set: true
	// Second set: false
	// Value: 42
}

func ExampleOnceCell_GetOrInit() {
	cell := NewOnceCell[string]()

	value := cell.GetOrInit(func() string {
		return "Lazy initialized value"
	})

	fmt.Println(value)
	// Output: Lazy initialized value
}

func ExampleOnceCell_IsInitialized() {
	cell := NewOnceCell[string]()

	fmt.Println("Before:", cell.IsInitialized())
	cell.Set("Hello")
	fmt.Println("After:", cell.IsInitialized())

	// Output:
	// Before: false
	// After: true
}
