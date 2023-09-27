package tabulator

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLessThan(t *testing.T) {
	assert.Equal(t, -1, compare("A", "B"))

	ts := time.Now()
	assert.Equal(t, -1, compare(ts, ts.Add(time.Hour)))
	assert.Equal(t, 0, compare(ts, ts))
	assert.Equal(t, 1, compare(ts, ts.Add(-time.Hour)))
}
