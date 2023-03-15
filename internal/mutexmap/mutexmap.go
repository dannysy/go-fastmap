package mutexmap

import (
	"sync"
)

type MutexMap struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

func New() *MutexMap {
	c := MutexMap{
		mu:    sync.RWMutex{},
		items: map[string]interface{}{},
	}
	return &c
}

func (c *MutexMap) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	val, ok := c.items[key]
	c.mu.RUnlock()
	return val, ok
}

func (c *MutexMap) Set(key string, value interface{}) {
	c.mu.Lock()
	c.items[key] = value
	c.mu.Unlock()
}

func (c *MutexMap) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}
