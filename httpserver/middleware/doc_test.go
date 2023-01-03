package middleware_test

import (
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

func ExamplePrometheusMetrics() {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello"))
	}

	// This will create the following metrics:
	//   foo_bar_http_requests_total
	//   foo_bar_http_requests_duration_seconds
	//
	// Both will have a label "handler" set to "example".
	// The duration metrics will be a Summary (i.e. the default)
	m := middleware.NewPrometheusMetrics(middleware.PrometheusMetricsOptions{
		Namespace:   "foo",
		Subsystem:   "bar",
		Application: "example",
	})
	prometheus.MustRegister(m)

	s := &http.Server{
		Addr:    ":8080",
		Handler: m.Handle(http.HandlerFunc(handler)),
	}

	_ = s.ListenAndServe()
}
