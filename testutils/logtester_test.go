package testutils_test

import (
	"bytes"
	"errors"
	"github.com/clambin/go-common/testutils"
	"io"
	"log/slog"
	"testing"
)

func TestLogTester(t *testing.T) {
	tests := []struct {
		name   string
		logger func(io.Writer, slog.Level) *slog.Logger
		want   string
	}{
		{
			name:   "NewTextLogger",
			logger: testutils.NewTextLogger,
			want: `level=ERROR msg=test err=error
`,
		},
		{
			name:   "NewJSONLogger",
			logger: testutils.NewJSONLogger,
			want: `{"level":"ERROR","msg":"test","err":"error"}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			l := tt.logger(&output, slog.LevelInfo)
			l.Error("test", "err", errors.New("error"))

			if got := output.String(); got != tt.want {
				t.Errorf("LogTester() = %v, want %v", got, tt.want)
			}

		})
	}
}
