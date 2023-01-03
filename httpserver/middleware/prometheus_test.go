package middleware_test

import (
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServerMetrics(t *testing.T) {
	tests := []struct {
		name       string
		metricType middleware.PrometheusMetricsDurationType
	}{
		{name: "summary", metricType: middleware.Summary},
		{name: "histogram", metricType: middleware.Histogram},
	}

	h := func(w http.ResponseWriter, _ *http.Request) {
		//w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := middleware.NewPrometheusMetrics(middleware.PrometheusMetricsOptions{
				Namespace:   "foo",
				Subsystem:   "bar",
				Application: "snafu",
				MetricsType: tt.metricType,
			})
			reg := prometheus.NewPedanticRegistry()
			reg.MustRegister(m)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			m.Handle(http.HandlerFunc(h)).ServeHTTP(w, r)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "OK", w.Body.String())

			metrics, err := reg.Gather()
			require.NoError(t, err)
			assert.Len(t, metrics, 2)

			for _, metric := range metrics {
				switch metric.GetName() {
				case "foo_bar_http_requests_total":
					require.Len(t, metric.GetMetric(), 1)
					assert.Equal(t, 1.0, metric.GetMetric()[0].GetCounter().GetValue())
				case "foo_bar_http_requests_duration_seconds":
					require.Len(t, metric.GetMetric(), 1)
					if tt.metricType == middleware.Summary {
						assert.Equal(t, uint64(1), metric.GetMetric()[0].GetSummary().GetSampleCount())
					} else if tt.metricType == middleware.Histogram {
						assert.Equal(t, uint64(1), metric.GetMetric()[0].GetHistogram().GetSampleCount())
					}
				default:
					t.Fatalf("unepxected metric: %s", metric.GetName())
				}
			}
		})
	}
}
