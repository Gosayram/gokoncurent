package gokoncurent

import (
	"testing"
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
