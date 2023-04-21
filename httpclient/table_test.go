package httpclient

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestCacheTable_ShouldCache(t *testing.T) {
	table := CacheTable{
		{
			Path:   "/foo",
			Expiry: time.Second,
		},
		{
			Path:    "/bar",
			Methods: []string{http.MethodGet},
			Expiry:  time.Minute,
		},
		{
			Path:     "/snafu/[a-z]+",
			IsRegExp: true,
			Expiry:   time.Hour,
		},
	}

	tests := []struct {
		name     string
		method   string
		url      string
		cache    bool
		duration time.Duration
	}{
		{
			name:     "no method",
			method:   http.MethodGet,
			url:      "/foo",
			cache:    true,
			duration: time.Second,
		},
		{
			name:     "method match",
			method:   http.MethodGet,
			url:      "/bar",
			cache:    true,
			duration: time.Minute,
		},
		{
			name:   "method mismatch",
			method: http.MethodPut,
			url:    "/bar",
			cache:  false,
		},
		{
			name:     "regexp",
			method:   http.MethodGet,
			url:      "/snafu/foobar",
			cache:    true,
			duration: time.Hour,
		},
		{
			name:   "regexp mismatch",
			method: http.MethodGet,
			url:    "/snafu/123",
			cache:  false,
		},
		{
			name:   "url mismatch",
			method: http.MethodGet,
			url:    "/snafu",
			cache:  false,
		},
	}

	table.mustCompile()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			cache, duration := table.shouldCache(req)
			assert.Equal(t, tt.cache, cache)
			if tt.cache {
				assert.Equal(t, tt.duration, duration)
			}
		})
	}
}

func TestCacheTable_DefaultCacheTable(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	cache, _ := DefaultCacheTable.shouldCache(req)
	assert.True(t, cache)
}

func TestCacheTable_BadCacheTable(t *testing.T) {
	table := CacheTable{{
		Path:     "/snafu/[a-",
		IsRegExp: true,
	}}
	assert.Panics(t, func() {
		table.mustCompile()
	})
}
