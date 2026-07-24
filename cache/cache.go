package cache

import (
	"container/list"
	"sync"
	"time"
)

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
}

func New(maxsize int) *Cache {
	return &Cache{
		store:   make(map[string]*item),
		order:   list.New(),
		maxsize: maxsize,
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.store) >= c.maxsize {
		oldest := c.order.Back()
		keyRemove := oldest.Value.(string)
		delete(c.store, keyRemove)
		c.order.Remove(oldest)
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
		return nil, false
	}

	if time.Now().After(item.expires) {
		delete(c.store, key)
		return nil, false
	}

	c.order.MoveToFront(item.element)
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
