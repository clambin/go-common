package cache

import (
	"maps"
	"runtime"
	"sync"
	"time"
)

// Cache implements a generic cache with expiration timers for entries, with optional removal of expired items.
type Cache[K comparable, V any] struct {
	*realCache[K, V]
	*scrubber[K, V]
}

type realCache[K comparable, V any] struct {
	values     map[K]entry[V]
	expiration time.Duration
	lock       sync.RWMutex
}

type entry[V any] struct {
	value  V
	expiry time.Time
}

func (e entry[K]) isExpired() bool {
	return !e.expiry.IsZero() && time.Until(e.expiry) < 0
}

// New creates a new Cache for the specified key and value types. The expiration parameter specifies the default time
// an entry can live in the cache before expiring. The cleanup parameters specifies how often the cache will remove
// expired items.
func New[K comparable, V any](expiration, cleanup time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		realCache: &realCache[K, V]{
			values:     make(map[K]entry[V]),
			expiration: expiration,
		},
	}
	if cleanup > 0 {
		c.scrubber = &scrubber[K, V]{
			period: cleanup,
			halt:   make(chan struct{}),
			cache:  c.realCache,
		}
		go c.scrubber.run()
		runtime.SetFinalizer(c, stopScrubber[K, V])
	}
	return c
}

// Add adds a key/value pair to the cache, using the default expiry time
func (c *Cache[K, V]) Add(key K, value V) {
	c.AddWithExpiry(key, value, c.expiration)
}

// AddWithExpiry adds a key/value pair to the cache with a specified expiration timer
func (c *Cache[K, V]) AddWithExpiry(key K, value V, expiry time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var e time.Time
	if expiry != 0 {
		e = time.Now().Add(expiry)
	}

	c.values[key] = entry[V]{
		value:  value,
		expiry: e,
	}
}

// Get returns the value from the cache for the provided key. If the item is not found, or expired, found will be false
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	e, found := c.values[key]
	if found {
		found = !e.isExpired()
	}
	return e.value, found
}

// GetKeys returns all keys in the cache.
func (c *Cache[K, V]) GetKeys() (keys []K) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for key := range c.values {
		keys = append(keys, key)
	}
	return
}

// Size returns the current size of the cache. Expired items are counted
func (c *Cache[K, V]) Size() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.values)
}

// Len returns the number of non-expired items in the case
func (c *Cache[K, V]) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var count int
	for _, e := range c.values {
		if !e.isExpired() {
			count++
		}
	}
	return count
}

// GetDefaultExpiration returns the default expiration time of the cache
func (c *Cache[K, V]) GetDefaultExpiration() time.Duration {
	return c.expiration
}

func (c *realCache[K, V]) scrub() {
	c.lock.Lock()
	defer c.lock.Unlock()

	maps.DeleteFunc(c.values, func(k K, e entry[V]) bool {
		return time.Now().After(e.expiry)
	})
}
