package cache

import (
	"container/list"
	"sync"
	"time"
)

type Stats struct {
	Hits      int
	Misses    int
	Evictions int
}

type item struct {
	value   interface{}
	expires time.Time
	element *list.Element
}

type Cache struct {
	store   map[string]*item
	mu      sync.Mutex
	maxsize int
	order   *list.List
	stats   Stats
}

func New(maxsize int) *Cache {
	return &Cache{
		store:   make(map[string]*item),
		order:   list.New(),
		maxsize: maxsize,
		stats: Stats{
			Hits:      0,
			Misses:    0,
			Evictions: 0,
		},
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	//Eviction
	if len(c.store) >= c.maxsize {
		oldest := c.order.Back()
		keyRemove := oldest.Value.(string)
		delete(c.store, keyRemove)
		c.order.Remove(oldest)
		c.stats.Evictions += 1
	}

	elem := c.order.PushFront(key)
	c.store[key] = &item{
		value:   value,
		expires: time.Now().Add(ttl),
		element: elem,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.store[key]
	if !ok {
		c.stats.Misses += 1
		return nil, false
	}

	if time.Now().After(item.expires) {
		delete(c.store, key)
		c.stats.Misses += 1
		return nil, false
	}

	c.order.MoveToFront(item.element)
	c.stats.Hits += 1
	return item.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if it, ok := c.store[key]; ok {
		c.order.Remove(it.element)
		delete(c.store, key)
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[string]*item)
	c.order = list.New()
	c.stats = Stats{
		Hits:      0,
		Misses:    0,
		Evictions: 0,
	}
}

func (c *Cache) Stats() Stats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats
}
