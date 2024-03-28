package metrics_test

import (
	"github.com/clambin/go-common/http/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"strings"
	"testing"
)

func TestInflightMetric_Collect(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		subsystem   string
		constLabels map[string]string
		want        string
	}{
		{
			name: "default",
			want: `
# HELP inflight_requests number of requests currently in flight
# TYPE inflight_requests gauge
inflight_requests 1

# HELP inflight_requests_max highest number of in flight requests
# TYPE inflight_requests_max gauge
inflight_requests_max 2
`,
		},
		{
			name:      "prefixed",
			namespace: "foo",
			subsystem: "bar",
			want: `
 # HELP foo_bar_inflight_requests number of requests currently in flight
# TYPE foo_bar_inflight_requests gauge
foo_bar_inflight_requests 1

# HELP foo_bar_inflight_requests_max highest number of in flight requests
# TYPE foo_bar_inflight_requests_max gauge
foo_bar_inflight_requests_max 2
`,
		},
		{
			name:        "labels",
			constLabels: map[string]string{"application": "snafu"},
			want: `
 # HELP inflight_requests number of requests currently in flight
# TYPE inflight_requests gauge
inflight_requests{application="snafu"} 1

# HELP inflight_requests_max highest number of in flight requests
# TYPE inflight_requests_max gauge
inflight_requests_max{application="snafu"} 2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := metrics.NewInflightMetric(tt.namespace, tt.subsystem, tt.constLabels)
			m.Inc()
			m.Inc()
			m.Dec()

			if err := testutil.CollectAndCompare(m, strings.NewReader(tt.want)); err != nil {
				t.Error(err)
			}
		})
	}
}
