package cache_test

import (
	"context"
	"github.com/clambin/go-common/cache"
	"runtime"
	"slices"
	"strconv"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	c := cache.New[string, string](time.Hour, 0)

	if c.GetDefaultExpiration() != time.Hour {
		t.Error("default expiration was incorrect")
	}

	// Add a value
	c.Add("foo", "bar")
	value, found := c.Get("foo")
	if !found {
		t.Error("foo was not found")
	}
	if value != "bar" {
		t.Error("value was not bar")
	}

	// Overwrite an existing value
	c.Add("foo", "foo")
	value, found = c.Get("foo")
	if !found {
		t.Error("foo was not found")
	}
	if value != "foo" {
		t.Error("value was not foo")
	}

	// Verify Len & Size
	if c.Len() != 1 {
		t.Error("cache length should be 1")
	}
	if c.Size() != 1 {
		t.Error("cache size should be 1")
	}

	// Remove a value
	c.Remove("foo")
	if _, found = c.Get("foo"); found {
		t.Error("foo was found")
	}

	// GetAndRemove a value
	c.Add("foo", "bar")
	value, found = c.GetAndRemove("foo")
	if !found {
		t.Error("foo was not found")
	}
	if value != "bar" {
		t.Error("value was not bar")
	}
	_, found = c.Get("foo")
	if found {
		t.Error("foo was found")
	}
}

func TestCacheExpiry(t *testing.T) {
	const shortExpiration = 100 * time.Millisecond
	c := cache.New[string, string](shortExpiration, 0)

	c.Add("foo", "bar")

	expired := eventually(func() bool {
		_, found := c.Get("foo")
		return !found
	}, time.Second, shortExpiration)

	if !expired {
		t.Fatal("foo did not expire")
	}
	if c.Len() != 0 {
		t.Error("cache length should be 0")
	}
	if c.Size() != 1 {
		t.Error("cache size should be 1")
	}
}

func TestCache_AddWithExpiry(t *testing.T) {
	const shortExpiration = 100 * time.Millisecond
	c := cache.New[string, int](shortExpiration, 0)

	c.Add("foo", 1)
	c.AddWithExpiry("bar", 2, time.Hour)

	want := []string{"bar", "foo"}
	keys := c.Keys()
	slices.Sort(keys)
	if !slices.Equal(keys, want) {
		t.Errorf("got keys %v, want %v", keys, want)
	}

	expired := eventually(func() bool {
		_, found := c.Get("foo")
		return !found
	}, time.Second, shortExpiration)

	if !expired {
		t.Error("foo did not expire")
	}
}

func TestCache_Iterate(t *testing.T) {
	c := cache.New[string, string](time.Hour, 0)
	c.Add("foo", "foo")
	c.Add("bar", "bar")
	c.AddWithExpiry("snafu", "snafu", -time.Hour)

	// just doing this for code coverage
	for range c.Iterate() {
		break
	}

	var keys []string
	for k, v := range c.Iterate() {
		if k != v {
			t.Errorf("value %q does not match key %q", v, k)
		}
		keys = append(keys, k)
	}
	slices.Sort(keys)
	if len(keys) != 2 || keys[0] != "bar" || keys[1] != "foo" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestCacheScrubber(t *testing.T) {
	const shortExpiration = 100 * time.Millisecond
	c := cache.New[string, string](shortExpiration/2, shortExpiration)
	c.Add("foo", "bar")

	if scrubbed := eventually(func() bool {
		return c.Size() == 0
	}, time.Second, shortExpiration); !scrubbed {
		t.Error("cache was not scrubbed")
	}

	// force cleanup to be run and stop the scrubber
	c = nil
	runtime.GC()
}

func BenchmarkCache_Get(b *testing.B) {
	c := cache.New[int, string](time.Hour, 0)
	for i := range 100_0000 {
		c.Add(i, strconv.Itoa(i))
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		if _, ok := c.Get(1); !ok {
			b.Fail()
		}
	}
}

func eventually(f func() bool, timeout time.Duration, interval time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ch := make(chan bool)
	go func(ctx context.Context, f func() bool) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				ch <- false
				return
			case <-ticker.C:
				if ok := f(); ok {
					ch <- true
					return
				}
			}
		}
	}(ctx, f)

	return <-ch
}
