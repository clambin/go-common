package httpclient_test

import (
	"errors"
	"github.com/clambin/go-common/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestWithRequestMetrics(t *testing.T) {
	r := httpclient.NewRoundTripper(
		httpclient.WithMetrics("foo", "bar", "snafu"),
		httpclient.WithRoundTripper(httpclient.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		})),
	)

	req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
	_, err := r.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, 2, testutil.CollectAndCount(r))

	r = httpclient.NewRoundTripper(
		httpclient.WithMetrics("foo", "bar", "snafu"),
		httpclient.WithRoundTripper(httpclient.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
			return nil, errors.New("fail")
		})),
	)

	req, _ = http.NewRequest(http.MethodGet, "/foo", nil)
	_, err = r.RoundTrip(req)
	require.Error(t, err)

	assert.NoError(t, testutil.CollectAndCompare(r, strings.NewReader(`
# HELP foo_bar_api_errors_total Number of failed HTTP calls
# TYPE foo_bar_api_errors_total counter
foo_bar_api_errors_total{application="snafu",method="GET",path="/foo"} 1
`), "foo_bar_api_errors_total"))
}

func TestWithRequestMetrics_Custom(t *testing.T) {
	o := &customRequestMeasurer{
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName("foo", "bar", "api_errors_total"),
			Help:        "Number of failed HTTP calls",
			ConstLabels: map[string]string{"application": "snafu"},
		}, []string{"path", "method"}),
	}

	r := httpclient.NewRoundTripper(
		httpclient.WithCustomMetrics(o),
		httpclient.WithRoundTripper(httpclient.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
			return nil, errors.New("fail")
		})),
	)

	req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
	_, err := r.RoundTrip(req)
	require.Error(t, err)

	assert.NoError(t, testutil.CollectAndCompare(r, strings.NewReader(`
# HELP foo_bar_api_errors_total Number of failed HTTP calls
# TYPE foo_bar_api_errors_total counter
foo_bar_api_errors_total{application="snafu",method="GET",path="/foo"} 1
`), "foo_bar_api_errors_total"))
}

var _ httpclient.RequestMeasurer = customRequestMeasurer{}

type customRequestMeasurer struct {
	errors *prometheus.CounterVec
}

func (c customRequestMeasurer) MeasureRequest(req *http.Request, _ *http.Response, err error, _ time.Duration) {
	if err != nil {
		c.errors.WithLabelValues(req.URL.Path, req.Method).Add(1)
	}
}

func (c customRequestMeasurer) Describe(descs chan<- *prometheus.Desc) {
	c.errors.Describe(descs)
}

func (c customRequestMeasurer) Collect(metrics chan<- prometheus.Metric) {
	c.errors.Collect(metrics)
}
