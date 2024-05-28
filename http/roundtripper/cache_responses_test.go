package roundtripper

import (
	"bytes"
	"github.com/clambin/go-common/cache"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test_responseCache(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		method string
		found  bool
	}{
		{
			name:   "first call",
			url:    "/foo",
			method: http.MethodGet,
			found:  false,
		},
		{
			name:   "second call",
			url:    "/foo",
			method: http.MethodGet,
			found:  true,
		},
		{
			name:   "cache per method",
			url:    "/foo",
			method: http.MethodPost,
			found:  false,
		},
		{
			name:   "cache per method - second call",
			url:    "/foo",
			method: http.MethodPost,
			found:  true,
		},
		{
			name:   "method mismatch",
			url:    "/foo",
			method: http.MethodHead,
			found:  false,
		},
		{
			name:   "method mismatch - second call",
			url:    "/foo",
			method: http.MethodHead,
			found:  false,
		},
		{
			name:   "do not cache",
			url:    "/bar",
			method: http.MethodGet,
			found:  false,
		},
		{
			name:   "do not cache - second call",
			url:    "/bar",
			method: http.MethodGet,
			found:  false,
		},
	}

	c := responseCache{
		table: CacheTable{{Path: "/foo", Methods: []string{http.MethodGet, http.MethodPost}}},
		cache: cache.New[string, []byte](time.Minute, 0),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			key := req.Method + "|" + req.URL.Path
			_, found, err := c.get(key, req)
			if err != nil {
				t.Fatal(err)
			}
			if found != tt.found {
				t.Errorf("got %v, want %v", found, tt.found)
			}
			if key == "" {
				t.Errorf("got empty key")
			}

			if !tt.found {
				resp := &http.Response{
					Status:        "OK",
					StatusCode:    http.StatusOK,
					Body:          io.NopCloser(bytes.NewBufferString("Hello")),
					ContentLength: 5,
					Request:       req,
				}

				err = c.put(key, req, resp)
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkCachePut(b *testing.B) {
	c := responseCache{
		table: DefaultCacheTable,
		cache: cache.New[string, []byte](time.Minute, 0),
	}
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBufferString("this is a request"))
	resp := http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("this is a response")),
		Request:    req,
	}
	key := req.Method + "|" + req.URL.Path

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.put(key, req, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	c := responseCache{
		table: DefaultCacheTable,
		cache: cache.New[string, []byte](time.Minute, 0),
	}
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBufferString("this is a request"))
	resp := http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("this is a response")),
		Request:    req,
	}
	key := req.Method + "|" + req.URL.Path
	_ = c.put(key, req, &resp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok, err := c.get(key, req)
		if err != nil {
			b.Fatal(err)
		}
		if !ok {
			b.Fatal("response not found in cache???")
		}
	}
}
