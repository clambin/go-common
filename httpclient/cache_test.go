package httpclient

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestCacher_Put_Get(t *testing.T) {
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

	table := CacheTable{{Path: "/foo", Methods: []string{http.MethodGet, http.MethodPost}}}
	c := newCache(table, time.Minute, 0)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			key, _, found, err := c.get(req)
			require.NoError(t, err)
			assert.Equal(t, tt.found, found)
			assert.NotEmpty(t, key)

			if !tt.found {
				resp := &http.Response{
					Status:        "OK",
					StatusCode:    http.StatusOK,
					Body:          io.NopCloser(bytes.NewBufferString("Hello")),
					ContentLength: 5,
					Request:       req,
				}

				err = c.put(key, req, resp)
				require.NoError(t, err)
			}
		})
	}
}
