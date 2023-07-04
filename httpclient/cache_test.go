package httpclient_test

import (
	"bytes"
	"github.com/clambin/go-common/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestWithCache(t *testing.T) {
	s := stubbedServer{}
	r := httpclient.NewRoundTripper(
		httpclient.WithCache(httpclient.DefaultCacheTable, 0, 0),
		httpclient.WithRoundTripper(&s),
	)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	_, err := r.RoundTrip(req)
	require.NoError(t, err)

	_, err = r.RoundTrip(req)
	assert.NoError(t, err)

	assert.Equal(t, 1, s.called)
}

func TestWithInstrumentedCache(t *testing.T) {
	s := stubbedServer{}
	r := httpclient.NewRoundTripper(
		httpclient.WithInstrumentedCache(httpclient.DefaultCacheTable, 0, 0, "foo", "bar", "test"),
		httpclient.WithRoundTripper(&s),
	)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	_, err := r.RoundTrip(req)
	require.NoError(t, err)

	_, err = r.RoundTrip(req)
	assert.NoError(t, err)

	assert.Equal(t, 1, s.called)

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(r)

	assert.NoError(t, testutil.GatherAndCompare(reg, bytes.NewBufferString(`
# HELP foo_bar_api_cache_hit_total Number of times the cache was used
# TYPE foo_bar_api_cache_hit_total counter
foo_bar_api_cache_hit_total{application="test",method="GET",path="/"} 1
# HELP foo_bar_api_cache_total Number of times the cache was consulted
# TYPE foo_bar_api_cache_total counter
foo_bar_api_cache_total{application="test",method="GET",path="/"} 2
`)))
}
