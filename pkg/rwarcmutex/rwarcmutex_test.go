// Package rwarcmutex provides tests for RWArcMutex[T] concurrency primitive.
package rwarcmutex

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewRWArcMutex_Basic(t *testing.T) {
	m := NewRWArcMutex(42)
	require.NotNil(t, m)
	require.Equal(t, int64(1), m.RefCount())

	m.WithRLock(func(v *int) {
		require.Equal(t, 42, *v)
	})

	m.WithLock(func(v *int) {
		*v = 100
	})

	m.WithRLock(func(v *int) {
		require.Equal(t, 100, *v)
	})
}

func TestRWArcMutex_CloneAndDrop(t *testing.T) {
	m := NewRWArcMutex("hello")
	clone := m.Clone()
	require.Equal(t, int64(2), m.RefCount())

	clone2 := m.Clone()
	require.Equal(t, int64(3), m.RefCount())

	clone.Drop()
	require.Equal(t, int64(2), m.RefCount())
	clone2.Drop()
	require.Equal(t, int64(1), m.RefCount())

	m.Drop()
	require.Equal(t, int64(0), m.RefCount())
}

func TestRWArcMutex_ConcurrentAccess(t *testing.T) {
	m := NewRWArcMutex(0)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				m.WithLock(func(v *int) {
					*v++
				})
			}
		}()
	}
	wg.Wait()
	m.WithRLock(func(v *int) {
		require.Equal(t, 10000, *v)
	})
}

func TestRWArcMutex_RaceCloneDrop(t *testing.T) {
	m := NewRWArcMutex(1)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := m.Clone()
			time.Sleep(time.Millisecond)
			if c != nil {
				c.Drop()
			}
		}()
	}
	wg.Wait()
	m.Drop()
	require.Equal(t, int64(0), m.RefCount())
}

func TestRWArcMutex_NilAndClosed(t *testing.T) {
	var m *RWArcMutex[int]
	m.Clone() // should not panic
	m.Drop()  // should not panic
	m.WithRLock(func(_ *int) { t.Fail() })
	m.WithLock(func(_ *int) { t.Fail() })

	m2 := NewRWArcMutex(5)
	m2.Drop()
	m2.WithRLock(func(_ *int) { t.Fail() })
	m2.WithLock(func(_ *int) { t.Fail() })
}
