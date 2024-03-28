package logtester_test

import (
	"bytes"
	"errors"
	"github.com/clambin/solaredge-monitor/pkg/logtester"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestLogTester(t *testing.T) {
	var output bytes.Buffer
	l := logtester.New(&output, slog.LevelInfo)

	l.Error("test", "err", errors.New("error"))

	assert.Equal(t, "level=ERROR msg=test err=error\n", output.String())
}
