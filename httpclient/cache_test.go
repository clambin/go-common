package httpclient_test

import (
	"bytes"
	"github.com/clambin/go-common/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestCacher_Put_Get(t *testing.T) {
	//c := httpclient.Cache{Table: httpclient.CacheTable{{
	//	Path: "/foo",
	//}}}

	table := httpclient.CacheTable{{Path: "/foo"}}
	c := httpclient.NewCache(table, time.Minute, 0)

	tests := []struct {
		name  string
		url   string
		found bool
	}{
		{
			name: "first call",
			url:  "/foo",
		},
		{
			name:  "second call",
			url:   "/foo",
			found: true,
		},
		{
			name: "do not cache",
			url:  "/bar",
		},
		{
			name: "do not cache - second call",
			url:  "/bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
			key, _, found, err := c.Get(req)
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

				err = c.Put(key, req, resp)
				require.NoError(t, err)
			}
		})
	}
}
