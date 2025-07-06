package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"github.com/Gosayram/gokoncurent/pkg/arcmutex"
)

// This example demonstrates how ArcMutex[T] can be used to safely share and
// mutate a map[string][]int across multiple goroutines. The program spawns a
// number of workers that concurrently append random integers to slices stored
// inside the shared map under random keys. Access is synchronized via
// ArcMutex.WithLock, ensuring data races are avoided while retaining the
// convenience of reference counting for easy sharing.
func main() {
	// Shared map from string to slice of ints.
	shared := arcmutex.NewArcMutex(make(map[string][]int))
	defer shared.Drop()

	const (
		workers  = 10
		inserts  = 100
		keySpace = 5
	)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			_ = id // use id to avoid unused variable error
			defer wg.Done()
			for j := 0; j < inserts; j++ {
				keyID, _ := rand.Int(rand.Reader, big.NewInt(int64(keySpace)))
				key := fmt.Sprintf("key-%d", keyID.Int64())
				valueBig, _ := rand.Int(rand.Reader, big.NewInt(1000))
				value := int(valueBig.Int64())

				// Append value to slice under the key.
				shared.WithLock(func(m *map[string][]int) {
					(*m)[key] = append((*m)[key], value)
				})
			}
		}(i)
	}

	wg.Wait()

	// Read final state with read-only access.
	result := shared.WithLockResult(func(m *map[string][]int) interface{} {
		// Copy data to avoid holding lock while printing.
		copyMap := make(map[string][]int, len(*m))
		for k, v := range *m {
			copySlice := append([]int(nil), v...)
			copyMap[k] = copySlice
		}
		return copyMap
	}).(map[string][]int)

	// Print aggregated lengths per key.
	fmt.Println("Final number of items per key:")
	for k, v := range result {
		fmt.Printf("  %s: %d items\n", k, len(v))
	}
}
