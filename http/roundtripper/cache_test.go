package roundtripper_test

import (
	"bytes"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestWithCache(t *testing.T) {
	s := server{}
	r := roundtripper.New(
		roundtripper.WithCache(roundtripper.DefaultCacheTable, 0, 0),
		roundtripper.WithRoundTripper(&s),
	)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	_, err := r.RoundTrip(req)
	require.NoError(t, err)

	_, err = r.RoundTrip(req)
	assert.NoError(t, err)

	assert.Equal(t, 1, int(s.called.Load()))
}

func TestWithInstrumentedCache(t *testing.T) {
	s := server{}
	m := roundtripper.NewCacheMetrics("foo", "bar", "snafu")
	r := roundtripper.New(
		roundtripper.WithInstrumentedCache(roundtripper.DefaultCacheTable, 0, 0, m),
		roundtripper.WithRoundTripper(&s),
	)

	req, _ := http.NewRequest(http.MethodGet, "", nil)
	_, err := r.RoundTrip(req)
	require.NoError(t, err)

	_, err = r.RoundTrip(req)
	assert.NoError(t, err)

	assert.Equal(t, 1, int(s.called.Load()))

	assert.NoError(t, testutil.CollectAndCompare(m, bytes.NewBufferString(`
# HELP foo_bar_http_cache_hit_total Number of times the cache was used
# TYPE foo_bar_http_cache_hit_total counter
foo_bar_http_cache_hit_total{application="snafu",method="GET",path="/"} 1

# HELP foo_bar_http_cache_total Number of times the cache was consulted
# TYPE foo_bar_http_cache_total counter
foo_bar_http_cache_total{application="snafu",method="GET",path="/"} 2
`)))
}

func BenchmarkWithCache(b *testing.B) {
	var body bytes.Buffer
	for i := 0; i < 10000; i++ {
		body.WriteString("hello\n")
	}
	rt := roundtripper.New(
		roundtripper.WithCache(roundtripper.DefaultCacheTable, time.Minute, 0),
		roundtripper.WithRoundTripper(roundtripper.RoundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(&body)}, nil
		})),
	)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	for i := 0; i < b.N; i++ {
		_, err := rt.RoundTrip(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
