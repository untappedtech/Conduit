package cache

import (
	"container/list"
	"sync"
)

type cacheEntry[K comparable, V any] struct {
	key   K
	value V
}

// LRUCache is a thread-safe generic LRU cache.
type LRUCache[K comparable, V any] struct {
	mutex     sync.RWMutex
	capacity  int
	items     map[K]*list.Element
	evictList *list.List
}

// NewLRU creates a new LRUCache with the specified capacity.
func NewLRU[K comparable, V any](capacity int) *LRUCache[K, V] {
	if capacity <= 0 {
		capacity = 256
	}
	return &LRUCache[K, V]{
		capacity:  capacity,
		items:     make(map[K]*list.Element),
		evictList: list.New(),
	}
}

// Get looks up a key's value from the cache, updating its LRU position.
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		return elem.Value.(*cacheEntry[K, V]).value, true
	}

	var zero V
	return zero, false
}

// Peek returns a key's value without updating its LRU position.
func (c *LRUCache[K, V]) Peek(key K) (V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if elem, ok := c.items[key]; ok {
		return elem.Value.(*cacheEntry[K, V]).value, true
	}

	var zero V
	return zero, false
}

// Set adds a value to the cache, evicting the oldest element if at capacity.
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		elem.Value.(*cacheEntry[K, V]).value = value
		return
	}

	if c.evictList.Len() >= c.capacity {
		oldest := c.evictList.Back()
		if oldest != nil {
			c.evictList.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry[K, V]).key)
		}
	}

	entry := &cacheEntry[K, V]{key: key, value: value}
	elem := c.evictList.PushFront(entry)
	c.items[key] = elem
}

// Remove removes the provided key from the cache.
func (c *LRUCache[K, V]) Remove(key K) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.items[key]; ok {
		c.evictList.Remove(elem)
		delete(c.items, key)
		return true
	}
	return false
}

// Clear purges all stored entries from the cache.
func (c *LRUCache[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[K]*list.Element)
	c.evictList.Init()
}

// Len returns the current number of items in the cache.
func (c *LRUCache[K, V]) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

// Capacity returns the maximum capacity of the cache.
func (c *LRUCache[K, V]) Capacity() int {
	return c.capacity
}
