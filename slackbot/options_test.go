package slackbot

import (
	"context"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
)

func TestWithName(t *testing.T) {
	b := New("123")
	assert.Equal(t, "slackbot", b.name)

	b = New("123", WithName("foo"))
	assert.Equal(t, "foo", b.name)
}

func TestWithLogger(t *testing.T) {
	b := New("123")
	assert.Equal(t, slog.Default(), b.logger)

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	b = New("123", WithLogger(l))
	assert.Equal(t, l, b.logger)
}

func TestWithCommands(t *testing.T) {
	b := New("123")
	assert.Equal(t, []string{"help", "version"}, b.commandRunner.GetCommands())

	b = New("123", WithCommands(map[string]CommandFunc{
		"test": func(_ context.Context, args ...string) []slack.Attachment {
			return nil
		},
	}))
	assert.Equal(t, []string{"help", "test", "version"}, b.commandRunner.GetCommands())
}
