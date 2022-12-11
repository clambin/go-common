package httpserver_test

import (
	"errors"
	"github.com/clambin/go-common/httpserver"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

func Example() {
	s, err := httpserver.New(
		httpserver.WithPort{Port: 8080},
		httpserver.WithPrometheus{},
		httpserver.WithMetrics{Application: "example", MetricsType: httpserver.Summary},
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
		panic(err)
	}
	prometheus.MustRegister(s)
	err = s.Serve()
	if !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
