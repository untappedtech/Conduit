package cache_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/untappedtech/conduit/internal/cache"
)

func TestLRUCache_Basic(t *testing.T) {
	c := cache.NewLRU[string, int](2)

	if c.Len() != 0 {
		t.Fatalf("expected len 0, got %d", c.Len())
	}

	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 2 {
		t.Fatalf("expected len 2, got %d", c.Len())
	}

	val, ok := c.Get("a")
	if !ok || val != 1 {
		t.Fatalf("expected a=1, got %v (%v)", val, ok)
	}

	// Adding "c" should evict "b" since "a" was recently accessed
	c.Set("c", 3)
	if _, ok := c.Get("b"); ok {
		t.Fatalf("expected b to be evicted")
	}
	if val, ok := c.Get("a"); !ok || val != 1 {
		t.Fatalf("expected a=1 to remain")
	}
	if val, ok := c.Get("c"); !ok || val != 3 {
		t.Fatalf("expected c=3")
	}

	// Update existing key
	c.Set("a", 10)
	if val, ok := c.Get("a"); !ok || val != 10 {
		t.Fatalf("expected a=10, got %v", val)
	}

	// Remove
	if !c.Remove("a") {
		t.Fatalf("expected true removing a")
	}
	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected a to be gone")
	}
	if c.Remove("a") {
		t.Fatalf("expected false removing non-existent a")
	}

	// Clear
	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("expected len 0 after clear")
	}
}

func TestLRUCache_Concurrency(t *testing.T) {
	c := cache.NewLRU[string, string](50)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", (workerID+j)%70)
				c.Set(key, fmt.Sprintf("val-%d", workerID))
				c.Get(key)
				c.Peek(key)
				if j%10 == 0 {
					c.Remove(key)
				}
			}
		}(i)
	}

	wg.Wait()
}
