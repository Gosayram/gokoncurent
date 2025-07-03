package arc

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

func TestNewArc(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		arc := NewArc("hello")
		if arc == nil {
			t.Fatal("NewArc should not return nil")
		}
		if arc.Get() == nil {
			t.Fatal("Arc.Get() should not return nil")
		}
		if *arc.Get() != "hello" {
			t.Errorf("Expected 'hello', got '%s'", *arc.Get())
		}
		if arc.RefCount() != 1 {
			t.Errorf("Expected reference count 1, got %d", arc.RefCount())
		}
	})

	t.Run("int value", func(t *testing.T) {
		arc := NewArc(42)
		if *arc.Get() != 42 {
			t.Errorf("Expected 42, got %d", *arc.Get())
		}
	})

	t.Run("struct value", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		value := TestStruct{Name: "Alice", Age: 30}
		arc := NewArc(value)
		retrieved := arc.Get()
		if retrieved.Name != "Alice" || retrieved.Age != 30 {
			t.Errorf("Expected {Alice 30}, got %+v", *retrieved)
		}
	})
}

func TestArcClone(t *testing.T) {
	original := NewArc("test")

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
		if original.Get() != clone.Get() {
			t.Error("Clone should share the same data pointer")
		}
		if *original.Get() != *clone.Get() {
			t.Error("Clone should have the same value")
		}
	})

	t.Run("clone of nil returns nil", func(t *testing.T) {
		var nilArc *Arc[string]
		clone := nilArc.Clone()
		if clone != nil {
			t.Error("Clone of nil should return nil")
		}
	})
}

func TestArcGet(t *testing.T) {
	t.Run("valid arc", func(t *testing.T) {
		arc := NewArc(123)
		value := arc.Get()
		if value == nil {
			t.Fatal("Get should not return nil for valid arc")
		}
		if *value != 123 {
			t.Errorf("Expected 123, got %d", *value)
		}
	})

	t.Run("nil arc", func(t *testing.T) {
		var nilArc *Arc[int]
		value := nilArc.Get()
		if value != nil {
			t.Error("Get should return nil for nil arc")
		}
	})
}

func TestArcRefCount(t *testing.T) {
	arc := NewArc("test")

	t.Run("initial count", func(t *testing.T) {
		if arc.RefCount() != 1 {
			t.Errorf("Expected initial reference count 1, got %d", arc.RefCount())
		}
	})

	t.Run("after cloning", func(t *testing.T) {
		clone1 := arc.Clone()
		if arc.RefCount() != 2 {
			t.Errorf("Expected reference count 2, got %d", arc.RefCount())
		}

		clone2 := arc.Clone()
		if arc.RefCount() != 3 {
			t.Errorf("Expected reference count 3, got %d", arc.RefCount())
		}

		// Clean up
		clone1.Drop()
		clone2.Drop()
	})

	t.Run("nil arc", func(t *testing.T) {
		var nilArc *Arc[string]
		if nilArc.RefCount() != 0 {
			t.Error("Nil arc should have reference count 0")
		}
	})
}

func TestArcDrop(t *testing.T) {
	t.Run("drop single reference", func(t *testing.T) {
		arc := NewArc("test")
		if arc.RefCount() != 1 {
			t.Errorf("Expected reference count 1, got %d", arc.RefCount())
		}

		wasLast := arc.Drop()
		if !wasLast {
			t.Error("Drop should return true when dropping the last reference")
		}
		if arc.RefCount() != 0 {
			t.Errorf("Expected reference count 0 after drop, got %d", arc.RefCount())
		}
	})

	t.Run("drop with multiple references", func(t *testing.T) {
		arc := NewArc("test")
		clone := arc.Clone()

		wasLast := arc.Drop()
		if wasLast {
			t.Error("Drop should return false when there are still references")
		}
		if clone.RefCount() != 1 {
			t.Errorf("Expected reference count 1 after dropping one reference, got %d", clone.RefCount())
		}

		wasLast = clone.Drop()
		if !wasLast {
			t.Error("Drop should return true when dropping the last reference")
		}
	})

	t.Run("drop nil arc", func(t *testing.T) {
		var nilArc *Arc[string]
		wasLast := nilArc.Drop()
		if wasLast {
			t.Error("Drop of nil arc should return false")
		}
	})
}

func TestArcIsValid(t *testing.T) {
	t.Run("valid arc", func(t *testing.T) {
		arc := NewArc("test")
		if !arc.IsValid() {
			t.Error("New arc should be valid")
		}
	})

	t.Run("nil arc", func(t *testing.T) {
		var nilArc *Arc[string]
		if nilArc.IsValid() {
			t.Error("Nil arc should not be valid")
		}
	})

	t.Run("dropped arc", func(t *testing.T) {
		arc := NewArc("test")
		arc.Drop()
		if arc.IsValid() {
			t.Error("Dropped arc should not be valid")
		}
	})
}

func TestArcConcurrency(t *testing.T) {
	t.Run("concurrent cloning", func(t *testing.T) {
		arc := NewArc("concurrent test")
		const numGoroutines = 100
		var wg sync.WaitGroup
		clones := make([]*Arc[string], numGoroutines)

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				defer wg.Done()
				clones[index] = arc.Clone()
			}(i)
		}

		wg.Wait()

		// Check that all clones are valid and reference count is correct
		expectedCount := int64(numGoroutines + 1) // +1 for original
		if arc.RefCount() != expectedCount {
			t.Errorf("Expected reference count %d, got %d", expectedCount, arc.RefCount())
		}

		for i, clone := range clones {
			if clone == nil {
				t.Errorf("Clone %d should not be nil", i)
				continue
			}
			if !clone.IsValid() {
				t.Errorf("Clone %d should be valid", i)
			}
			if *clone.Get() != "concurrent test" {
				t.Errorf("Clone %d has wrong value: %s", i, *clone.Get())
			}
		}

		// Clean up
		for _, clone := range clones {
			if clone != nil {
				clone.Drop()
			}
		}
		arc.Drop()
	})

	t.Run("concurrent access", func(t *testing.T) {
		arc := NewArc(42)
		const numGoroutines = 50
		const numIterations = 100
		var wg sync.WaitGroup

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < numIterations; j++ {
					// Test concurrent read access
					value := arc.Get()
					if value == nil || *value != 42 {
						t.Errorf("Unexpected value during concurrent access")
						return
					}

					// Test concurrent cloning and dropping
					clone := arc.Clone()
					if clone == nil {
						t.Error("Clone should not be nil during concurrent access")
						return
					}

					// Brief pause to increase chance of race conditions
					runtime.Gosched()

					clone.Drop()
				}
			}()
		}

		wg.Wait()

		// Original should still be valid
		if !arc.IsValid() {
			t.Error("Original arc should still be valid after concurrent access")
		}
		if arc.RefCount() != 1 {
			t.Errorf("Expected reference count 1, got %d", arc.RefCount())
		}
	})
}

func TestArcRaceConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	t.Run("clone and drop race", func(t *testing.T) {
		for iteration := 0; iteration < 100; iteration++ {
			arc := NewArc("race test")
			var wg sync.WaitGroup

			// Start multiple goroutines that clone and immediately drop
			wg.Add(10)
			for i := 0; i < 10; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						clone := arc.Clone()
						if clone != nil {
							clone.Drop()
						}
					}
				}()
			}

			wg.Wait()

			// Original should still be valid
			if !arc.IsValid() {
				t.Error("Original arc should still be valid after race test")
			}

			arc.Drop()
		}
	})
}

// Benchmark tests
func BenchmarkArcNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arc := NewArc(i)
		_ = arc
	}
}

func BenchmarkArcClone(b *testing.B) {
	arc := NewArc("benchmark")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clone := arc.Clone()
		clone.Drop()
	}
}

func BenchmarkArcGet(b *testing.B) {
	arc := NewArc("benchmark")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arc.Get()
	}
}

func BenchmarkArcConcurrentClone(b *testing.B) {
	arc := NewArc("concurrent benchmark")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			clone := arc.Clone()
			clone.Drop()
		}
	})
}

// Example tests for documentation
func ExampleNewArc() {
	arc := NewArc("Hello, World!")
	fmt.Println(*arc.Get())
	// Output: Hello, World!
}

func ExampleArc_Clone() {
	original := NewArc(42)
	clone := original.Clone()

	fmt.Println(*original.Get())
	fmt.Println(*clone.Get())
	fmt.Println(original.RefCount())

	// Output:
	// 42
	// 42
	// 2
}
