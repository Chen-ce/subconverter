package cache

import (
	"strings"
	"sync"
	"time"
)

type item struct {
	value      interface{}
	expiration int64
}

// Cache 简单的内存缓存
type Cache struct {
	items map[string]item
	mu    sync.RWMutex
}

// New 创建一个新的缓存
func New() *Cache {
	return &Cache{
		items: make(map[string]item),
	}
}

// Set 设置缓存
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items[key] = item{
		value:      value,
		expiration: expiration,
	}
}

// Get 获取缓存
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	it, found := c.items[key]
	if !found {
		return nil, false
	}
	
	if it.expiration > 0 && time.Now().UnixNano() > it.expiration {
		return nil, false
	}
	
	return it.value, true
}

// Delete 删除缓存
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// DeletePrefix 删除匹配前缀的所有缓存
func (c *Cache) DeletePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.items {
		if strings.HasPrefix(key, prefix) {
			delete(c.items, key)
		}
	}
}
