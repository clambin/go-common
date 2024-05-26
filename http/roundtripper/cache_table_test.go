package roundtripper

import (
	"github.com/clambin/go-common/http/pkg/testutils"
	"net/http"
	"testing"
	"time"
)

func TestCacheTable_shouldCache(t *testing.T) {
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
			Path:    "",
			Methods: []string{http.MethodPost},
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
			name:     "empty path matches",
			method:   http.MethodPost,
			url:      "/snafu",
			cache:    true,
			duration: time.Minute,
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
			if cache != tt.cache {
				t.Errorf("shouldCache() cache = %v, want %v", cache, tt.cache)
			}
			if tt.cache {
				if duration != tt.duration {
					t.Errorf("shouldCache() duration = %v, want %v", duration, tt.duration)
				}
			}
		})
	}
}

func TestCacheTable_DefaultCacheTable(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	cache, _ := DefaultCacheTable.shouldCache(req)
	if cache != true {
		t.Errorf("shouldCache() cache = %v, want %v", cache, true)
	}
}

func TestCacheTable_BadCacheTable(t *testing.T) {
	table := CacheTable{{
		Path:     "/snafu/[a-",
		IsRegExp: true,
	}}

	if ok := testutils.Panics(table.mustCompile); !ok {
		t.Error("function did not panic")
	}
}
