package metrics_test

import (
	"github.com/clambin/go-common/http/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewRequestMetrics(t *testing.T) {
	tests := []struct {
		name         string
		namespace    string
		subsystem    string
		constLabels  prometheus.Labels
		durationType metrics.DurationType
		buckets      []float64
		path         string
		want         string
	}{
		{
			name:         "all-in",
			namespace:    "foo",
			subsystem:    "bar",
			constLabels:  prometheus.Labels{"application": "app"},
			durationType: metrics.SummaryDuration,
			path:         "/",
			want: `
# HELP foo_bar_http_request_duration_seconds duration of http requests
# TYPE foo_bar_http_request_duration_seconds summary
foo_bar_http_request_duration_seconds_sum{application="app",code="200",method="GET",path="/"} 1
foo_bar_http_request_duration_seconds_count{application="app",code="200",method="GET",path="/"} 1

# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="app",code="200",method="GET",path="/"} 1
`,
		},
		{
			name:         "subst path",
			namespace:    "foo",
			subsystem:    "bar",
			constLabels:  prometheus.Labels{"application": "app"},
			durationType: metrics.SummaryDuration,
			path:         "",
			want: `
# HELP foo_bar_http_request_duration_seconds duration of http requests
# TYPE foo_bar_http_request_duration_seconds summary
foo_bar_http_request_duration_seconds_sum{application="app",code="200",method="GET",path="/"} 1
foo_bar_http_request_duration_seconds_count{application="app",code="200",method="GET",path="/"} 1

# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="app",code="200",method="GET",path="/"} 1
`,
		},
		{
			name: "no labels",
			path: "/",
			want: `
# HELP http_request_duration_seconds duration of http requests
# TYPE http_request_duration_seconds summary
http_request_duration_seconds_sum{code="200",method="GET",path="/"} 1
http_request_duration_seconds_count{code="200",method="GET",path="/"} 1

# HELP http_requests_total total number of http requests
# TYPE http_requests_total counter
http_requests_total{code="200",method="GET",path="/"} 1
`,
		},
		{
			name:         "all-in",
			namespace:    "foo",
			subsystem:    "bar",
			constLabels:  prometheus.Labels{"application": "app"},
			durationType: metrics.HistogramDuration,
			path:         "/",
			want: `
# HELP foo_bar_http_request_duration_seconds duration of http requests
# TYPE foo_bar_http_request_duration_seconds histogram
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.005"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.01"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.025"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.05"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.1"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.25"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.5"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="1"} 1
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="2.5"} 1
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="5"} 1
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="10"} 1
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="+Inf"} 1
foo_bar_http_request_duration_seconds_sum{application="app",code="200",method="GET",path="/"} 1
foo_bar_http_request_duration_seconds_count{application="app",code="200",method="GET",path="/"} 1
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="app",code="200",method="GET",path="/"} 1
`,
		},
		{
			name:         "subst path",
			namespace:    "foo",
			subsystem:    "bar",
			constLabels:  prometheus.Labels{"application": "app"},
			durationType: metrics.HistogramDuration,
			path:         "",
			want: `
# HELP foo_bar_http_request_duration_seconds duration of http requests
# TYPE foo_bar_http_request_duration_seconds histogram
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.005"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.01"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.025"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.05"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.1"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.25"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="0.5"} 0
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="1"} 1
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="2.5"} 1
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="5"} 1
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="10"} 1
foo_bar_http_request_duration_seconds_bucket{application="app",code="200",method="GET",path="/",le="+Inf"} 1
foo_bar_http_request_duration_seconds_sum{application="app",code="200",method="GET",path="/"} 1
foo_bar_http_request_duration_seconds_count{application="app",code="200",method="GET",path="/"} 1
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="app",code="200",method="GET",path="/"} 1
`,
		},
		{
			name:         "no labels",
			durationType: metrics.HistogramDuration,
			path:         "/",
			want: `
# HELP http_request_duration_seconds duration of http requests
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="0.005"} 0
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="0.01"} 0
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="0.025"} 0
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="0.05"} 0
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="0.1"} 0
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="0.25"} 0
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="0.5"} 0
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="1"} 1
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="2.5"} 1
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="5"} 1
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="10"} 1
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="+Inf"} 1
http_request_duration_seconds_sum{code="200",method="GET",path="/"} 1
http_request_duration_seconds_count{code="200",method="GET",path="/"} 1

# HELP http_requests_total total number of http requests
# TYPE http_requests_total counter
http_requests_total{code="200",method="GET",path="/"} 1
`,
		},
		{
			name:         "buckets",
			path:         "/",
			durationType: metrics.HistogramDuration,
			buckets:      []float64{0.1, 1, 2},
			want: `
# HELP http_request_duration_seconds duration of http requests
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="0.1"} 0
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="1"} 1
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="2"} 1
http_request_duration_seconds_bucket{code="200",method="GET",path="/",le="+Inf"} 1
http_request_duration_seconds_sum{code="200",method="GET",path="/"} 1
http_request_duration_seconds_count{code="200",method="GET",path="/"} 1

# HELP http_requests_total total number of http requests
# TYPE http_requests_total counter
http_requests_total{code="200",method="GET",path="/"} 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := metrics.NewRequestMetrics(metrics.Options{
				Namespace:    tt.namespace,
				Subsystem:    tt.subsystem,
				ConstLabels:  tt.constLabels,
				DurationType: tt.durationType,
				Buckets:      tt.buckets,
			})
			req := http.Request{Method: http.MethodGet, URL: &url.URL{Path: tt.path}}
			m.Measure(&req, http.StatusOK, time.Second)

			if err := testutil.CollectAndCompare(m, strings.NewReader(tt.want)); err != nil {
				t.Error(err)
			}
		})
	}
}
