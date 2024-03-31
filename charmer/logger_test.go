package charmer_test

import (
	"bytes"
	"context"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/testutils"
	"github.com/spf13/cobra"
	"log/slog"
	"testing"
)

func TestSetLogger_With_Handler(t *testing.T) {
	tests := []struct {
		name          string
		setter        func(*cobra.Command, bool)
		isTextHandler bool
		isJSONHandler bool
	}{
		{
			name:          "JSON",
			setter:        charmer.SetJSONLogger,
			isJSONHandler: true,
		},
		{
			name:          "Text",
			setter:        charmer.SetTextLogger,
			isTextHandler: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd cobra.Command
			tt.setter(&cmd, true)
			l := charmer.GetLogger(&cmd)
			l.Debug("test")
			if ok := l.Handler().Enabled(context.Background(), slog.LevelDebug); !ok {
				t.Errorf("logger should have been enabled")
			}
			if tt.isTextHandler {
				if _, ok := l.Handler().(*slog.TextHandler); !ok {
					t.Errorf("logger should have been a TextHandler")
				}
			}
			if tt.isJSONHandler {
				if _, ok := l.Handler().(*slog.JSONHandler); !ok {
					t.Errorf("logger should have been a JSONHandler")
				}
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	var cmd cobra.Command

	logger := charmer.GetLogger(&cmd)
	if logger != slog.Default() {
		t.Errorf("logger should return default logger")
	}

	cmd.SetContext(context.WithValue(context.Background(), "logger", nil)) //nolint:all
	if logger != slog.Default() {
		t.Errorf("logger should return default logger")
	}

	var buffer bytes.Buffer
	l := testutils.NewTextLogger(&buffer, slog.LevelDebug)
	charmer.SetLogger(&cmd, l)

	logger = charmer.GetLogger(&cmd)
	if logger != l {
		t.Errorf("logger should return new logger")
	}

	l.Debug("test")
	if got := buffer.String(); got != "level=DEBUG msg=test\n" {
		t.Errorf("incorrect log output: %s", got)
	}
}
