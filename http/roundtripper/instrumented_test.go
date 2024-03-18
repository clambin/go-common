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
	s := server{}
	m := roundtripper.NewDefaultRoundTripMetrics("foo", "bar")
	r := roundtripper.WithInstrumentedRoundTripper(m)(&s)

	req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
	_, err := r.RoundTrip(req)
	require.NoError(t, err)

	assert.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(`
# HELP foo_bar_requests_total total number of requests
# TYPE foo_bar_requests_total counter
foo_bar_requests_total{code="200",method="GET",path="/foo"} 1
`), "foo_bar_requests_total"))
	assert.Equal(t, 1, testutil.CollectAndCount(m, "foo_bar_latency"))

	s.fail = true

	req, _ = http.NewRequest(http.MethodGet, "/foo", nil)
	_, err = r.RoundTrip(req)
	require.Error(t, err)

	assert.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(`
# HELP foo_bar_requests_total total number of requests
# TYPE foo_bar_requests_total counter
foo_bar_requests_total{code="",method="GET",path="/foo"} 1
foo_bar_requests_total{code="200",method="GET",path="/foo"} 1
`), "foo_bar_requests_total"))
	assert.Equal(t, 2, testutil.CollectAndCount(m, "foo_bar_latency"))
}
