package middleware_test

import (
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/middleware"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultServerSummaryMetrics(t *testing.T) {
	tests := []struct {
		name        string
		application string
		code        int
		want        string
	}{
		{
			name:        "success",
			application: "",
			code:        http.StatusOK,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{code="200",method="GET",path="/"} 1
`,
		},
		{
			name:        "success - with application",
			application: "snafu",
			code:        http.StatusOK,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="snafu",code="200",method="GET",path="/"} 1
`,
		},
		{
			name:        "failure",
			application: "",
			code:        http.StatusNotFound,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{code="404",method="GET",path="/"} 1
`,
		},
		{
			name:        "success - with application",
			application: "snafu",
			code:        http.StatusNotFound,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="snafu",code="404",method="GET",path="/"} 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var labels map[string]string
			if tt.application != "" {
				labels = map[string]string{"application": tt.application}
			}
			m := metrics.NewRequestSummaryMetrics("foo", "bar", labels)

			h := middleware.WithRequestMetrics(m)(
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(tt.code)
				}),
			)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			h.ServeHTTP(w, r)
			assert.Equal(t, tt.code, w.Code)
			assert.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(tt.want), "foo_bar_http_requests_total"))
			assert.Equal(t, 2, testutil.CollectAndCount(m), "foo_bar_http_request_duration_seconds")
		})
	}
}

func TestDefaultServerHistogramMetrics(t *testing.T) {
	tests := []struct {
		name        string
		application string
		code        int
		want        string
	}{
		{
			name:        "success",
			application: "",
			code:        http.StatusOK,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{code="200",method="GET",path="/"} 1
`,
		},
		{
			name:        "success - with application",
			application: "snafu",
			code:        http.StatusOK,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="snafu",code="200",method="GET",path="/"} 1
`,
		},
		{
			name:        "failure",
			application: "",
			code:        http.StatusNotFound,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{code="404",method="GET",path="/"} 1
`,
		},
		{
			name:        "success - with application",
			application: "snafu",
			code:        http.StatusNotFound,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="snafu",code="404",method="GET",path="/"} 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var labels map[string]string
			if tt.application != "" {
				labels = map[string]string{"application": tt.application}
			}
			m := metrics.NewRequestHistogramMetrics("foo", "bar", labels)

			h := middleware.WithRequestMetrics(m)(
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(tt.code)
				}),
			)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			h.ServeHTTP(w, r)
			assert.Equal(t, tt.code, w.Code)

			assert.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(tt.want), "foo_bar_http_requests_total"))
			// TODO: why is this only 2 for histograms?
			assert.Equal(t, 2, testutil.CollectAndCount(m), "foo_bar_http_request_duration_seconds")
		})
	}
}

func TestWithInflightMetrics(t *testing.T) {
	s := server{wait: 500 * time.Millisecond}
	m := metrics.NewInflightMetric("foo", "bar", map[string]string{"application": "snafu"})
	h := middleware.WithInflightMetrics(m)(&s)

	ch := make(chan struct{})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	go func() {
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		ch <- struct{}{}
	}()

	assert.Eventually(t, func() bool {
		return s.counter.Load() > 0
	}, time.Second, time.Millisecond)

	assert.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(`
# HELP foo_bar_inflight_requests number of requests currently in flight
# TYPE foo_bar_inflight_requests gauge
foo_bar_inflight_requests{application="snafu"} 1

# HELP foo_bar_inflight_requests_max highest number of in flight requests
# TYPE foo_bar_inflight_requests_max gauge
foo_bar_inflight_requests_max{application="snafu"} 1
`)))

	<-ch

	assert.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(`
# HELP foo_bar_inflight_requests number of requests currently in flight
# TYPE foo_bar_inflight_requests gauge
foo_bar_inflight_requests{application="snafu"} 0

# HELP foo_bar_inflight_requests_max highest number of in flight requests
# TYPE foo_bar_inflight_requests_max gauge
foo_bar_inflight_requests_max{application="snafu"} 1
`)))
}

type server struct {
	wait    time.Duration
	counter atomic.Int32
}

func (s *server) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	s.counter.Add(1)
	time.Sleep(s.wait)
}
