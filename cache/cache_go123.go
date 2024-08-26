//go:build go1.23

package cache

import "iter"

// Iterate returns an iterator that yields all non-expired keys & they value.
//
// Note: the cache is locked while the iterator is running.
func (c *Cache[K, V]) Iterate() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		c.lock.RLock()
		defer c.lock.RUnlock()

		for k, e := range c.values {
			if !e.isExpired() {
				if !yield(k, e.value) {
					return
				}
			}
		}
	}
}
