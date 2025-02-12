package roundtripper_test

import (
	"bytes"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestWithCache(t *testing.T) {
	s := server{}
	o := roundtripper.CacheOptions{CacheTable: roundtripper.DefaultCacheTable}
	r := roundtripper.New(
		roundtripper.WithCache(o),
		roundtripper.WithRoundTripper(&s),
	)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	_, err := r.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}

	_, err = r.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.called.Load(); got != 1 {
		t.Errorf("called.Load() = %d, want 1", got)
	}
}

func TestWithInstrumentedCache(t *testing.T) {
	s := server{}
	m := roundtripper.NewCacheMetrics(roundtripper.CacheMetricsOptions{
		Namespace:   "foo",
		Subsystem:   "bar",
		ConstLabels: prometheus.Labels{"application": "snafu"},
	})
	o := roundtripper.CacheOptions{CacheTable: roundtripper.DefaultCacheTable, CacheMetrics: m}
	r := roundtripper.New(
		roundtripper.WithCache(o),
		roundtripper.WithRoundTripper(&s),
	)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	if _, err := r.RoundTrip(req); err != nil {
		t.Fatal(err)
	}

	if _, err := r.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if got := s.called.Load(); got != 1 {
		t.Errorf("called.Load() = %d, want 1", got)
	}

	if err := testutil.CollectAndCompare(m, bytes.NewBufferString(`
# HELP foo_bar_http_cache_hit_total Number of times the cache was used
# TYPE foo_bar_http_cache_hit_total counter
foo_bar_http_cache_hit_total{application="snafu",method="GET",path="/"} 1

# HELP foo_bar_http_cache_total Number of times the cache was consulted
# TYPE foo_bar_http_cache_total counter
foo_bar_http_cache_total{application="snafu",method="GET",path="/"} 2
`)); err != nil {
		t.Errorf("incorrect metrics: %s", err)
	}
}

// Current:
// BenchmarkWithCache-16    	 1000000	      1087 ns/op	    4516 B/op	       8 allocs/op
func BenchmarkWithCache(b *testing.B) {
	rt := roundtripper.New(
		roundtripper.WithCache(roundtripper.CacheOptions{CacheTable: roundtripper.DefaultCacheTable, DefaultExpiration: time.Minute}),
		roundtripper.WithRoundTripper(roundtripper.RoundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(strings.Repeat("hello\n", 10_000))),
			}, nil
		})),
	)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := rt.RoundTrip(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
