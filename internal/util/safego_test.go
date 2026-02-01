package util

import (
	"sync"
	"testing"
	"time"
)

func TestSafeGo(t *testing.T) {
	var wg sync.WaitGroup
	executed := false

	wg.Add(1)
	SafeGo(func() {
		defer wg.Done()
		executed = true
	})

	wg.Wait()

	if !executed {
		t.Error("SafeGo did not execute the function")
	}
}

func TestSafeGoWithPanic(t *testing.T) {
	var wg sync.WaitGroup
	panicked := make(chan bool, 1)

	wg.Add(1)
	SafeGo(func() {
		defer wg.Done()
		defer func() {
			// This will be called after the panic is recovered
			panicked <- true
		}()
		panic("test panic")
	})

	// Wait for the goroutine with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - goroutine completed without crashing
	case <-time.After(2 * time.Second):
		t.Error("SafeGo did not complete in time")
	}
}

func TestSafeGoWithName(t *testing.T) {
	var wg sync.WaitGroup
	executed := false

	wg.Add(1)
	SafeGoWithName("test-goroutine", func() {
		defer wg.Done()
		executed = true
	})

	wg.Wait()

	if !executed {
		t.Error("SafeGoWithName did not execute the function")
	}
}

func TestSafeGoWithNameAndPanic(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	SafeGoWithName("test-panic-goroutine", func() {
		defer wg.Done()
		panic("test named panic")
	})

	// Wait for the goroutine with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - goroutine completed without crashing
	case <-time.After(2 * time.Second):
		t.Error("SafeGoWithName did not complete in time after panic")
	}
}

func TestSafeGoConcurrent(t *testing.T) {
	const numGoroutines = 100
	var wg sync.WaitGroup
	counter := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		SafeGo(func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		})
	}

	wg.Wait()

	if counter != numGoroutines {
		t.Errorf("Expected counter to be %d, got %d", numGoroutines, counter)
	}
}
