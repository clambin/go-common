package slackbot

import (
	"context"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	handler := func(text string) Handler {
		return HandlerFunc(func(_ context.Context, args ...string) []slack.Attachment {
			if len(args) > 0 {
				text += ": " + strings.Join(args, ", ")
			}
			return []slack.Attachment{{Text: text}}
		})
	}

	tests := []struct {
		name     string
		commands Commands
		args     []string
		want     []slack.Attachment
	}{
		{
			name:     "single command",
			commands: Commands{"foo": handler("foo")},
			args:     []string{"foo"},
			want:     []slack.Attachment{{Text: "foo"}},
		},
		{
			name:     "single command with args",
			commands: Commands{"foo": handler("foo")},
			args:     []string{"foo", "a=b"},
			want:     []slack.Attachment{{Text: "foo: a=b"}},
		},
		{
			name:     "empty",
			commands: Commands{"foo": handler("foo")},
			args:     nil,
			want:     []slack.Attachment{{Color: "bad", Title: "invalid command", Text: "supported commands: foo"}},
		},
		{
			name:     "invalid command",
			commands: Commands{"foo": handler("foo")},
			args:     []string{"bar"},
			want:     []slack.Attachment{{Color: "bad", Title: "invalid command", Text: "supported commands: foo"}},
		},
		{
			name:     "nested command",
			commands: Commands{"foo": NewCommandGroup(Commands{"bar": handler("bar")})},
			args:     []string{"foo", "bar"},
			want:     []slack.Attachment{{Text: "bar"}},
		},
		{
			name:     "invalid nested command",
			commands: Commands{"foo": NewCommandGroup(Commands{"bar": handler("bar")})},
			args:     []string{"foo", "foo"},
			want:     []slack.Attachment{{Color: "bad", Title: "invalid command", Text: "supported commands: bar"}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var c CommandGroup
			c.Add(tt.commands)

			output := c.Handle(context.Background(), tt.args...)
			assert.Equal(t, tt.want, output)

		})
	}
}
