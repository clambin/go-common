package prometheus_test

import (
	"context"
	"fmt"
	"github.com/clambin/go-common/taskmanager"
	"github.com/clambin/go-common/taskmanager/prometheus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var testcases = []struct {
		name    string
		options []prometheus.Option
		port    int
		path    string
	}{
		{
			name: "default",
			port: 9090,
			path: "/metrics",
		},
		{
			name:    "alt address",
			options: []prometheus.Option{prometheus.WithAddr(":9091")},
			port:    9091,
			path:    "/metrics",
		},
		{
			name:    "alt path",
			options: []prometheus.Option{prometheus.WithPath("/alt")},
			port:    9090,
			path:    "/alt",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			mgr := taskmanager.New(prometheus.New(tt.options...))

			ctx, cancel := context.WithCancel(context.Background())
			ch := make(chan error)
			go func() {
				ch <- mgr.Run(ctx)
			}()

			assert.Eventually(t, func() bool {
				resp, err := http.Get(fmt.Sprintf("http://localhost:%d/%s", tt.port, tt.path))
				if err != nil {
					return false
				}
				_ = resp.Body.Close()
				return resp.StatusCode == http.StatusOK
			}, time.Second, time.Millisecond)

			cancel()
		})
	}
}
