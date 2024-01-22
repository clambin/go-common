package slackbot

import (
	"context"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	var rootCmd Command

	var roomCmd Command
	roomCmd.Add("set", func(ctx context.Context, args ...string) []slack.Attachment {
		return []slack.Attachment{{Text: "room set: " + strings.Join(args, ", ")}}
	})
	rootCmd.AddCommand("room", &roomCmd)

	rootCmd.Add("list", func(ctx context.Context, args ...string) []slack.Attachment {
		return []slack.Attachment{{Text: "list"}}
	})

	rootCmd.Add("set", func(ctx context.Context, args ...string) []slack.Attachment {
		return []slack.Attachment{{Text: "set: " + strings.Join(args, ", ")}}
	})

	tests := []struct {
		name string
		args []string
		want []slack.Attachment
	}{
		{
			name: "list",
			args: []string{"list"},
			want: []slack.Attachment{{Text: "list"}},
		},
		{
			name: "set",
			args: []string{"set", "a=b"},
			want: []slack.Attachment{{Text: "set: a=b"}},
		},
		{
			name: "set (no args)",
			args: []string{"set"},
			want: []slack.Attachment{{Text: "set: "}},
		},
		{
			name: "room set",
			args: []string{"room", "set", "key=val"},
			want: []slack.Attachment{{Text: "room set: key=val"}},
		},
		{
			name: "invalid command",
			args: []string{"`invalid"},
			want: []slack.Attachment{{
				Color: "bad",
				Title: "invalid command",
				Text:  "supported commands: list, room, set",
			}},
		},
		{
			name: "invalid command with junk",
			args: []string{"`invalid", "help"},
			want: []slack.Attachment{{
				Color: "bad",
				Title: "invalid command",
				Text:  "supported commands: list, room, set",
			}},
		},
		{
			name: "invalid subcommand",
			args: []string{"room", "invalid"},
			want: []slack.Attachment{{
				Color: "bad",
				Title: "invalid command",
				Text:  "supported commands: set",
			}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := rootCmd.handle(context.Background(), tt.args...)
			assert.Equal(t, tt.want, output)

		})
	}
}
