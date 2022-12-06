package httpclient

import (
	"net/http"
)

// InstrumentedClient implements the Caller interface. If provided by Options, it will collect performance metrics of the API calls
// and record them for Prometheus to scrape.
type InstrumentedClient struct {
	BaseClient
	Options     Options
	Application string
}

var _ Caller = &InstrumentedClient{}

// Options contains options to alter InstrumentedClient behaviour
type Options struct {
	PrometheusMetrics *Metrics // Prometheus metric to record API performance metrics
}

// Do sends the request and records performance metrics of the call.
// Currently, it records the request's duration (i.e. latency) and error rate.
func (c *InstrumentedClient) Do(req *http.Request) (resp *http.Response, err error) {
	endpoint := req.URL.Path
	timer := c.Options.PrometheusMetrics.makeLatencyTimer(c.Application, endpoint, req.Method)

	resp, err = c.BaseClient.Do(req)

	if timer != nil {
		timer.ObserveDuration()
	}
	c.Options.PrometheusMetrics.reportErrors(err, c.Application, endpoint, req.Method)
	return
}
