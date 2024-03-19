package middleware_test

import (
	"github.com/clambin/go-common/http/middleware"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
# HELP foo_bar_http_server_requests_total total number of http server requests
# TYPE foo_bar_http_server_requests_total counter
foo_bar_http_server_requests_total{code="200",method="GET",path="/"} 1
`,
		},
		{
			name:        "success - with application",
			application: "snafu",
			code:        http.StatusOK,
			want: `
# HELP foo_bar_http_server_requests_total total number of http server requests
# TYPE foo_bar_http_server_requests_total counter
foo_bar_http_server_requests_total{application="snafu",code="200",method="GET",path="/"} 1
`,
		},
		{
			name:        "failure",
			application: "",
			code:        http.StatusNotFound,
			want: `
# HELP foo_bar_http_server_requests_total total number of http server requests
# TYPE foo_bar_http_server_requests_total counter
foo_bar_http_server_requests_total{code="404",method="GET",path="/"} 1
`,
		},
		{
			name:        "success - with application",
			application: "snafu",
			code:        http.StatusNotFound,
			want: `
# HELP foo_bar_http_server_requests_total total number of http server requests
# TYPE foo_bar_http_server_requests_total counter
foo_bar_http_server_requests_total{application="snafu",code="404",method="GET",path="/"} 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := middleware.NewDefaultServerSummaryMetrics("foo", "bar", tt.application)

			h := middleware.WithServerMetrics(metrics)(
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(tt.code)
				}),
			)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			h.ServeHTTP(w, r)
			assert.Equal(t, tt.code, w.Code)
			assert.NoError(t, testutil.CollectAndCompare(metrics, strings.NewReader(tt.want), "foo_bar_http_server_requests_total"))
			assert.Equal(t, 2, testutil.CollectAndCount(metrics), "foo_bar_http_server_request_duration_seconds")
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
# HELP foo_bar_http_server_requests_total total number of http server requests
# TYPE foo_bar_http_server_requests_total counter
foo_bar_http_server_requests_total{code="200",method="GET",path="/"} 1
`,
		},
		{
			name:        "success - with application",
			application: "snafu",
			code:        http.StatusOK,
			want: `
# HELP foo_bar_http_server_requests_total total number of http server requests
# TYPE foo_bar_http_server_requests_total counter
foo_bar_http_server_requests_total{application="snafu",code="200",method="GET",path="/"} 1
`,
		},
		{
			name:        "failure",
			application: "",
			code:        http.StatusNotFound,
			want: `
# HELP foo_bar_http_server_requests_total total number of http server requests
# TYPE foo_bar_http_server_requests_total counter
foo_bar_http_server_requests_total{code="404",method="GET",path="/"} 1
`,
		},
		{
			name:        "success - with application",
			application: "snafu",
			code:        http.StatusNotFound,
			want: `
# HELP foo_bar_http_server_requests_total total number of http server requests
# TYPE foo_bar_http_server_requests_total counter
foo_bar_http_server_requests_total{application="snafu",code="404",method="GET",path="/"} 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := middleware.NewDefaultServerHistogramMetrics("foo", "bar", tt.application)

			h := middleware.WithServerMetrics(metrics)(
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(tt.code)
				}),
			)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			h.ServeHTTP(w, r)
			assert.Equal(t, tt.code, w.Code)

			assert.NoError(t, testutil.CollectAndCompare(metrics, strings.NewReader(tt.want), "foo_bar_http_server_requests_total"))
			// TODO: why is this only 2 for histograms?
			assert.Equal(t, 2, testutil.CollectAndCount(metrics), "foo_bar_http_server_request_duration_seconds")
		})
	}
}
