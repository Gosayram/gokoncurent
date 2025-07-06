package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Gosayram/gokoncurent/pkg/oncecell"
)

// randInt returns a cryptographically secure random integer in [0, max).
func randInt(max int64) (int64, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return nBig.Int64(), nil
}

// Simulated external resource initialization that may fail.
func unreliableInit() (string, error) {
	// 50% chance to fail using crypto/rand.
	v, err := randInt(100)
	if err != nil {
		return "", err
	}
	if v < 50 { // 50% probability
		return "", errors.New("transient initialization failure")
	}

	id, err := randInt(1000)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("resource-%d", id), nil
}

func main() {
	// Create OnceCell for a string resource.
	cell := oncecell.NewOnceCell[string]()

	// Attempt to initialize with up to 5 retries and exponential backoff.
	value, err := cell.GetOrInitWithRetry(
		func() (string, error) {
			return unreliableInit()
		},
		5,                   // maxRetries
		50*time.Millisecond, // initial backoff
	)

	if err != nil {
		fmt.Println("Failed to initialize resource after retries:", err)
		return
	}

	fmt.Println("Successfully initialized resource:", value)

	// Subsequent calls will return immediately without reinitializing.
	cached, ok := cell.Get()
	if ok {
		fmt.Println("Cached resource value:", cached)
	}
}
