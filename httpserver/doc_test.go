package httpserver_test

import (
	"github.com/clambin/go-common/httpserver"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

func Example() {
	metrics := httpserver.NewAvgMetrics("example")
	prometheus.MustRegister(metrics)
	s, err := httpserver.New(
		httpserver.WithPort{Port: 8080},
		httpserver.WithPrometheus{},
		httpserver.WithMetrics{Metrics: metrics},
		httpserver.WithHandlers{Handlers: []httpserver.Handler{{
			Path:    "/",
			Methods: []string{http.MethodGet},
			Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("Hello world"))
			}),
		}},
		},
	)

	if err != nil {
		err = s.Run()
	}

	if err != nil {
		panic(err)
	}
}
