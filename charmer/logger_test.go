package charmer_test

import (
	"context"
	"github.com/clambin/go-common/charmer"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
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

	if charmer.GetLogger(&cmd) != slog.Default() {
		t.Errorf("logger should return default logger")
	}

	l := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	charmer.SetLogger(&cmd, l)
	if charmer.GetLogger(&cmd) != l {
		t.Errorf("logger should return custom logger")
	}

	charmer.SetLogger(&cmd, nil)
	if charmer.GetLogger(&cmd) != slog.Default() {
		t.Errorf("logger should return default logger")
	}
}
