package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(2 * time.Second)

	cache.Set("key1", "value1")

	value, found := cache.Get("key1")
	if !found || value != "value1" {
		t.Errorf("expected value 'value1', got '%v', found: %v", value, found)
	}

	time.Sleep(3 * time.Second)
	_, found = cache.Get("key1")
	if found {
		t.Error("expected key to expire, but it was found")
	}
}

func TestCache_GetNonExistentKey(t *testing.T) {
	cache := NewCache(2 * time.Second)

	_, found := cache.Get("nonexistent")
	if found {
		t.Error("expected key to not be found, but it was")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(2 * time.Second)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	cache.Clear()

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	if found1 || found2 {
		t.Error("expected cache to be cleared, but keys were found")
	}
}

func TestCache_SetOverwrite(t *testing.T) {
	cache := NewCache(2 * time.Second)

	cache.Set("key1", "value1")
	cache.Set("key1", "value2")

	value, found := cache.Get("key1")
	if !found || value != "value2" {
		t.Errorf("expected value 'value2', got '%v', found: %v", value, found)
	}
}

func BenchmarkCache_ConcurrentGetSet(b *testing.B) {
	cache := NewCache(10 * time.Second)

	const goroutines = 100
	var wg sync.WaitGroup

	worker := func(id int, iterations int) {
		defer wg.Done()
		key := fmt.Sprintf("key-%d", id)
		for i := 0; i < iterations; i++ {
			value := fmt.Sprintf("value-%d-%d", id, i)
			cache.Set(key, value)
			_, _ = cache.Get(key)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(goroutines)
		for j := 0; j < goroutines; j++ {
			go worker(j, 10)
		}
		wg.Wait()
	}
}
