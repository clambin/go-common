package roundtripper_test

import (
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
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
foo_bar_http_requests_total{code="",method="GET",path="/foo"} 1
`,
		},
		{
			name:        "failure - with application",
			pass:        false,
			application: "snafu",
			want: `
# HELP foo_bar_http_requests_total total number of http requests
# TYPE foo_bar_http_requests_total counter
foo_bar_http_requests_total{application="snafu",code="",method="GET",path="/foo"} 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := server{fail: !tt.pass}
			m := roundtripper.NewDefaultRoundTripMetrics("foo", "bar", tt.application)
			r := roundtripper.New(
				roundtripper.WithInstrumentedRoundTripper(m),
				roundtripper.WithRoundTripper(&s),
			)

			req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
			_, err := r.RoundTrip(req)
			if tt.pass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			assert.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(tt.want), "foo_bar_http_requests_total"))
			assert.Equal(t, 1, testutil.CollectAndCount(m, "foo_bar_http_request_duration_seconds"))
		})
	}
}
