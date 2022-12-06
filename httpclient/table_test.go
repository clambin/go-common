package httpclient

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestCacheTable_ShouldCache(t *testing.T) {
	table := CacheTable{Table: []CacheTableEntry{
		{Endpoint: `/foo`},
		{Endpoint: `/foo/[\d+]`, IsRegExp: true},
		{Endpoint: `/bar/.*`, IsRegExp: true, Methods: []string{http.MethodGet}},
	}}

	type testcase struct {
		path   string
		method string
		expiry time.Duration
		match  bool
	}
	for _, tc := range []testcase{
		{path: "/foo", match: true},
		{path: "/foo/123", match: true},
		{path: "/foo/bar", match: false},
		{path: "/bar/get", method: http.MethodGet, match: true},
		{path: "/bar/post", method: http.MethodPost, match: false},
		{path: "/foobar", match: false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			found, expiry := table.shouldCache(req)
			assert.Equal(t, tc.match, found)
			assert.Equal(t, tc.expiry, expiry)
		})

	}

	assert.True(t, table.compiled)
	for _, entry := range table.Table {
		if entry.IsRegExp {
			assert.NotNil(t, entry.compiledRegExp)
		} else {
			assert.Nil(t, entry.compiledRegExp)
		}
	}
}

func TestCacheTable_CacheEverything(t *testing.T) {
	table := CacheTable{}

	found, _ := table.shouldCache(&http.Request{URL: &url.URL{Path: "/"}})
	assert.True(t, found)
}

func TestCacheTable_Invalid_Input(t *testing.T) {
	table := CacheTable{Table: []CacheTableEntry{
		{Endpoint: `/foo/[\d+`, IsRegExp: true},
	}}

	assert.Panics(t, func() { table.shouldCache(&http.Request{URL: &url.URL{Path: "/foo"}}) })
}
