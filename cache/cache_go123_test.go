//go:build go1.23

package cache_test

import (
	"github.com/clambin/go-common/cache"
	"slices"
	"testing"
	"time"
)

func TestCache_Iterate(t *testing.T) {
	c := cache.New[string, string](time.Hour, 0)
	c.Add("foo", "foo")
	c.Add("bar", "bar")
	c.AddWithExpiry("snafu", "snafu", -time.Hour)

	// just doing this for code coverage
	for _, _ = range c.Iterate() {
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
