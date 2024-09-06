package utils

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      string
	Expiration int64
}

type Cache struct {
	mu    sync.Mutex
	items map[string]CacheItem
	ttl   time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
		ttl:   ttl,
	}
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(c.ttl).Unix(),
	}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, found := c.items[key]
	if !found || time.Now().Unix() > item.Expiration {
		delete(c.items, key)
		return "", false
	}
	return item.Value, true
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheItem)
}
