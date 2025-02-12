package testutils_test

import (
	"github.com/clambin/go-common/testutils"
	"testing"
	"time"
)

func TestPanics(t *testing.T) {
	if ok := testutils.Panics(func() {
		panic("error")
	}); !ok {
		t.Error("expected panic")
	}

	if ok := testutils.Panics(func() {
		return
	}); ok {
		t.Error("unexpected panic")
	}
}

func TestEventually(t *testing.T) {
	if ok := testutils.Eventually(func() bool { return true }, 100*time.Millisecond, time.Millisecond); !ok {
		t.Error("expected true")
	}

	if ok := testutils.Eventually(func() bool { return false }, 100*time.Millisecond, time.Millisecond); ok {
		t.Error("expected false")
	}
}
