package cache

import (
	"time"
)

type scrubber[K comparable, V any] struct {
	period time.Duration
	halt   chan struct{}
	cache  *realCache[K, V]
}

func (s *scrubber[K, V]) run() {
	ticker := time.NewTicker(s.period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cache.scrub()
		case <-s.halt:
			return
		}
	}
}

func stopScrubber[K comparable, V any](c *Cache[K, V]) {
	c.scrubber.halt <- struct{}{}
}
