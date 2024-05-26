package roundtripper_test

import (
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/pkg/testutils"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestWithInstrumentedRoundTripper(t *testing.T) {
	tests := []struct {
		name        string
		pass        bool
		application string
		want        string
	}{
		{
			name: "success",
			pass: true,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{code="200",method="GET",path="/foo"} 1
`,
		},
		{
			name:        "success - with application",
			pass:        true,
			application: "snafu",
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="snafu",code="200",method="GET",path="/foo"} 1
`,
		},
		{
			name: "failure",
			pass: false,
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{code="0",method="GET",path="/foo"} 1
`,
		},
		{
			name:        "failure - with application",
			pass:        false,
			application: "snafu",
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="snafu",code="0",method="GET",path="/foo"} 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := server{fail: !tt.pass}

			var labels map[string]string
			if tt.application != "" {
				labels = map[string]string{"application": tt.application}
			}
			m := metrics.NewRequestMetrics(metrics.Options{
				Namespace:   "foo",
				Subsystem:   "bar",
				ConstLabels: labels,
			})
			r := roundtripper.New(
				roundtripper.WithRequestMetrics(m),
				roundtripper.WithRoundTripper(&s),
			)

			req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
			_, err := r.RoundTrip(req)
			if tt.pass && err != nil {
				t.Errorf("round tripper.RoundTrip() error = %v, want pass %v", err, tt.pass)
			} else if !tt.pass && err == nil {
				t.Errorf("round tripper.RoundTrip() error = nil, want pass %v", tt.pass)
			}

			if err = testutil.CollectAndCompare(m, strings.NewReader(tt.want), "foo_bar_http_requests_total"); err != nil {
				t.Errorf("invalid metrics: %v", err)
			}
			if got := testutil.CollectAndCount(m, "foo_bar_http_request_duration_seconds"); got != 1 {
				t.Errorf("invalid metrics: got %d, want %d", got, 1)
			}
		})
	}
}

func TestWithInflightMetrics(t *testing.T) {
	s := server{delay: 500 * time.Millisecond}
	m := metrics.NewInflightMetric("foo", "bar", map[string]string{"application": "snafu"})
	r := roundtripper.New(
		roundtripper.WithInflightMetrics(m),
		roundtripper.WithRoundTripper(&s),
	)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	go func() {
		_, _ = r.RoundTrip(req)
	}()

	if ok := testutils.Eventually(func() bool { return s.inFlight.Load() > 0 }, time.Second, 10*time.Millisecond); !ok {
		t.Error("condition never satisfied")
	}

	if err := testutil.CollectAndCompare(m, strings.NewReader(`
# HELP foo_bar_inflight_requests number of requests currently in flight
# TYPE foo_bar_inflight_requests gauge
foo_bar_inflight_requests{application="snafu"} 1

# HELP foo_bar_inflight_requests_max highest number of in flight requests
# TYPE foo_bar_inflight_requests_max gauge
foo_bar_inflight_requests_max{application="snafu"} 1
`)); err != nil {
		t.Errorf("invalid metrics: %v", err)
	}
}
