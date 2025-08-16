package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type cacheItem struct {
	value      []byte
	expiration time.Time
}

type LRUCache struct {
	cache *lru.Cache[string, *cacheItem]
	mutex sync.RWMutex
}

func NewLRUCache(size int) (*LRUCache, error) {
	cache, err := lru.New[string, *cacheItem](size)
	if err != nil {
		return nil, err
	}

	return &LRUCache{
		cache: cache,
	}, nil
}

func (l *LRUCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	l.cache.Add(key, &cacheItem{
		value:      data,
		expiration: exp,
	})

	return nil
}

func (l *LRUCache) Get(ctx context.Context, key string, dest any) error {
	l.mutex.RLock()
	item, ok := l.cache.Get(key)
	l.mutex.RUnlock()

	if !ok {
		return ErrKeyNotFound
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		l.mutex.Lock()
		l.cache.Remove(key)
		l.mutex.Unlock()
		return ErrKeyNotFound
	}

	return json.Unmarshal(item.value, dest)
}

func (l *LRUCache) Delete(ctx context.Context, key string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.cache.Remove(key)
	return nil
}

func (l *LRUCache) Exists(ctx context.Context, key string) (bool, error) {
	l.mutex.RLock()
	item, ok := l.cache.Get(key)
	l.mutex.RUnlock()

	if !ok {
		return false, nil
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		l.mutex.Lock()
		l.cache.Remove(key)
		l.mutex.Unlock()
		return false, nil
	}

	return true, nil
}
