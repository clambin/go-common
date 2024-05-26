package middleware_test

import (
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/middleware"
	"github.com/clambin/go-common/http/pkg/testutils"
	"github.com/prometheus/client_golang/prometheus/testutil"
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
			m := metrics.NewRequestMetrics(metrics.Options{
				Namespace:    "foo",
				Subsystem:    "bar",
				ConstLabels:  labels,
				DurationType: metrics.SummaryDuration,
			})
			h := middleware.WithRequestMetrics(m)(
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(tt.code)
				}),
			)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			h.ServeHTTP(w, r)

			if w.Code != tt.code {
				t.Errorf("got %d, want %d", w.Code, tt.code)
			}
			if err := testutil.CollectAndCompare(m, strings.NewReader(tt.want), "foo_bar_http_requests_total"); err != nil {
				t.Error(err)
			}
			if want := testutil.CollectAndCount(m, "foo_bar_http_request_duration_seconds"); want != 1 {
				t.Errorf("got %d, want 1", want)
			}
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
			m := metrics.NewRequestMetrics(metrics.Options{
				Namespace:    "foo",
				Subsystem:    "bar",
				ConstLabels:  labels,
				DurationType: metrics.HistogramDuration,
			})

			h := middleware.WithRequestMetrics(m)(
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(tt.code)
				}),
			)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			h.ServeHTTP(w, r)

			if w.Code != tt.code {
				t.Errorf("got %d, want %d", w.Code, tt.code)
			}

			if err := testutil.CollectAndCompare(m, strings.NewReader(tt.want), "foo_bar_http_requests_total"); err != nil {
				t.Error(err)
			}
			if got := testutil.CollectAndCount(m, "foo_bar_http_request_duration_seconds"); got != 1 {
				t.Errorf("got %d, want 1", got)
			}
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
		ch <- struct{}{}
	}()

	if ok := testutils.Eventually(func() bool {
		return s.counter.Load() > 0
	}, time.Second, time.Millisecond); !ok {
		t.Fatal("condition never satisfied")
	}

	if err := testutil.CollectAndCompare(m, strings.NewReader(`
# HELP foo_bar_inflight_requests number of requests currently in flight
# TYPE foo_bar_inflight_requests gauge
foo_bar_inflight_requests{application="snafu"} 1

# HELP foo_bar_inflight_requests_max highest number of in flight requests
# TYPE foo_bar_inflight_requests_max gauge
foo_bar_inflight_requests_max{application="snafu"} 1
`)); err != nil {
		t.Error(err)
	}

	<-ch

	if err := testutil.CollectAndCompare(m, strings.NewReader(`
# HELP foo_bar_inflight_requests number of requests currently in flight
# TYPE foo_bar_inflight_requests gauge
foo_bar_inflight_requests{application="snafu"} 0

# HELP foo_bar_inflight_requests_max highest number of in flight requests
# TYPE foo_bar_inflight_requests_max gauge
foo_bar_inflight_requests_max{application="snafu"} 1
`)); err != nil {
		t.Error(err)
	}
}

type server struct {
	wait    time.Duration
	counter atomic.Int32
}

func (s *server) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	s.counter.Add(1)
	time.Sleep(s.wait)
}
