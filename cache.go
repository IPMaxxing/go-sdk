package ipmax

import (
	"container/list"
	"sync"
	"time"
)

type cacheEntry struct {
	key     string
	value   any
	expires time.Time
}

type lruCache struct {
	mu       sync.Mutex
	capacity int
	ttl      time.Duration
	order    *list.List
	entries  map[string]*list.Element
}

func newLRUCache(capacity int, ttl time.Duration) *lruCache {
	return &lruCache{capacity: capacity, ttl: ttl, order: list.New(), entries: map[string]*list.Element{}}
}

func (c *lruCache) get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	entry := element.Value.(*cacheEntry)
	if !time.Now().Before(entry.expires) {
		c.order.Remove(element)
		delete(c.entries, key)
		return nil, false
	}
	c.order.MoveToFront(element)
	return entry.value, true
}

func (c *lruCache) set(key string, value any) {
	if c.capacity <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := &cacheEntry{key: key, value: value, expires: time.Now().Add(c.ttl)}
	if element, ok := c.entries[key]; ok {
		element.Value = entry
		c.order.MoveToFront(element)
		return
	}
	c.entries[key] = c.order.PushFront(entry)
	for c.order.Len() > c.capacity {
		oldest := c.order.Back()
		c.order.Remove(oldest)
		delete(c.entries, oldest.Value.(*cacheEntry).key)
	}
}
