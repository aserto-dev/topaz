package tsync_test

import (
	"sync"
	"testing"

	"github.com/aserto-dev/topaz/internal/tsync"
)

// TestTypedSyncMapBasic validates fundamental CRUD methods.
func TestTypedSyncMapBasic(t *testing.T) {
	m := &tsync.Map[string, int]{}

	// Test Store and Load
	m.Store("key1", 42)

	val, ok := m.Load("key1")
	if !ok || val != 42 {
		t.Errorf("Expected 42, got %v (ok: %v)", val, ok)
	}

	// Test Load missing key
	_, ok = m.Load("missing")
	if ok {
		t.Error("Expected ok=false for missing key")
	}

	// Test LoadOrStore
	actual, loaded := m.LoadOrStore("key1", 99)
	if !loaded || actual != 42 {
		t.Errorf("LoadOrStore failed: expected loaded=true, actual=42; got loaded=%v, actual=%v", loaded, actual)
	}

	actual, loaded = m.LoadOrStore("key2", 100)
	if loaded || actual != 100 {
		t.Errorf("LoadOrStore failed: expected loaded=false, actual=100; got loaded=%v, actual=%v", loaded, actual)
	}

	// Test Delete
	m.Delete("key1")

	_, ok = m.Load("key1")
	if ok {
		t.Error("Key should have been deleted")
	}
}

// TestTypedSyncMapConcurrent runs multiple goroutines to verify race-free execution.
func TestTypedSyncMapConcurrent(t *testing.T) {
	m := &tsync.Map[int, int]{}

	var wg sync.WaitGroup

	workers := 10
	iterations := 1000

	// Concurrent Writers.
	for i := range workers {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for j := range iterations {
				m.Store(workerID*iterations+j, j)
			}
		}(i)
	}

	// Concurrent Readers.
	for i := range workers {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for j := range iterations {
				m.Load(workerID*iterations + j)
			}
		}(i)
	}

	wg.Wait()
}

// TestTypedSyncMapRange validates iteration and early termination behavior.
func TestTypedSyncMapRange(t *testing.T) {
	m := &tsync.Map[string, int]{}

	// Prepare sample dataset
	expectedData := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	for k, v := range expectedData {
		m.Store(k, v)
	}

	// 1. Test Full Iteration
	visited := make(map[string]int)

	m.Range(func(key string, value int) bool {
		visited[key] = value
		return true // Continue iteration
	})

	if len(visited) != len(expectedData) {
		t.Errorf("Expected %d elements, but visited %d", len(expectedData), len(visited))
	}

	for k, expectedVal := range expectedData {
		if gotVal, exists := visited[k]; !exists || gotVal != expectedVal {
			t.Errorf("Mismatch for key %s: expected %d, got %d", k, expectedVal, gotVal)
		}
	}

	// 2. Test Early Termination (returning false)
	count := 0

	m.Range(func(key string, value int) bool {
		count++
		return false // Stop immediately after the first element
	})

	if count != 1 {
		t.Errorf("Expected Range to stop after 1 iteration, but ran %d times", count)
	}
}

// BenchmarkTypedSyncMap_ReadHeavy simulates standard OPA Rego caching behavior
// (highly optimized for read-heavy scenarios with stable keys).
func BenchmarkTypedSyncMap_ReadHeavy(b *testing.B) {
	m := &tsync.Map[string, int]{}
	m.Store("policy_1", 100)
	m.Store("policy_2", 200)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = m.Load("policy_1")
			_, _ = m.Load("policy_2")
		}
	})
}

// BenchmarkTypedSyncMap_WriteHeavy simulates frequent updates/churn.
func BenchmarkTypedSyncMap_WriteHeavy(b *testing.B) {
	m := &tsync.Map[int, int]{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Store(i, i)
			i++
		}
	})
}
