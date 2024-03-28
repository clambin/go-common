package roundtripper_test

import (
	"bytes"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/prometheus/client_golang/prometheus/testutil"
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
	m := roundtripper.NewCacheMetrics("foo", "bar", "snafu")
	r := roundtripper.New(
		roundtripper.WithInstrumentedCache(roundtripper.DefaultCacheTable, 0, 0, m),
		roundtripper.WithRoundTripper(&s),
	)

	req, _ := http.NewRequest(http.MethodGet, "", nil)
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
