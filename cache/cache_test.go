package cache_test

import (
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

	c.Add("foo", "bar")
	value, found := c.Get("foo")
	if !found {
		t.Error("foo was not found")
	}
	if value != "bar" {
		t.Error("value was not bar")
	}

	c.Add("foo", "foo")
	value, found = c.Get("foo")
	if !found {
		t.Error("foo was not found")
	}
	if value != "foo" {
		t.Error("value was not foo")
	}

	if c.Len() != 1 {
		t.Error("cache length should be 1")
	}
	if c.Size() != 1 {
		t.Error("cache size should be 1")
	}

	c.Remove("foo")
	if _, found = c.Get("foo"); found {
		t.Error("foo was found")
	}
}

func TestCacheExpiry(t *testing.T) {
	const shortExpiration = 100 * time.Millisecond
	c := cache.New[string, string](shortExpiration, 0)

	c.Add("foo", "bar")

	retries := 2
	found := true
	for found && retries > 0 {
		time.Sleep(shortExpiration)
		_, found = c.Get("foo")
		retries--
	}

	if found {
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
	keys := c.GetKeys()
	slices.Sort(keys)
	if !slices.Equal(keys, want) {
		t.Errorf("got keys %v, want %v", keys, want)
	}

	retries := 2
	found := true
	for found && retries > 0 {
		time.Sleep(shortExpiration)
		_, found = c.Get("foo")
		retries--
	}

	if found {
		t.Error("foo did not expire")
	}
}

func TestCacheScrubber(t *testing.T) {
	const shortExpiration = 100 * time.Millisecond
	c := cache.New[string, string](shortExpiration/2, shortExpiration)

	c.Add("foo", "bar")

	retries := 5
	var empty bool
	for !empty && retries > 0 {
		time.Sleep(shortExpiration)
		empty = c.Len() == 0
		retries--
	}

	if !empty {
		t.Error("foo did not expire")
	}

	// this forces the runtime.Finalizer to call stopScrubber
	c = nil
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
}

/*
func TestCache_Types(t *testing.T) {
	c1 := cache.New[int, string](0, 0)
	c1.Add(1, "hello")
	v1, found := c1.Get(1)
	assert.True(t, found)
	assert.Equal(t, "hello", v1)

	c2 := cache.New[string, string](0, 0)
	c2.Add("foo", "bar")
	v2, found2 := c2.Get("foo")
	assert.True(t, found2)
	assert.Equal(t, "bar", v2)

	type testStruct struct {
		value int
	}
	key := testStruct{value: 1}

	c3 := cache.New[testStruct, string](0, 0)
	c3.Add(key, "snafu")
	v3, found3 := c3.Get(key)
	assert.True(t, found3)
	assert.Equal(t, "snafu", v3)
}
*/

func BenchmarkCache_Get(b *testing.B) {
	c := cache.New[int, string](time.Hour, 0)
	for i := 0; i < 1e5; i++ {
		c.Add(i, strconv.Itoa(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := c.Get(1)
		if !ok {
			b.Fail()
		}
	}
}
