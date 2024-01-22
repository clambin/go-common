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

func TestTokenizeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "one word",
			input: `do`,
			want:  []string{"do"},
		},
		{
			name:  "multiple words",
			input: `a b c `,
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "single-quoted words",
			input: `a 'b c'`,
			want:  []string{"a", "b c"},
		},
		{
			name:  "double-quoted words",
			input: `a "b c"`,
			want:  []string{"a", "b c"},
		},
		{
			name:  "inverse-quoted words",
			input: `a “b c"“`,
			want:  []string{"a", "b c"},
		},
		{
			name:  "empty",
			input: ``,
			want:  nil,
		},
		{
			name:  "empty quote",
			input: `""`,
			want:  []string{""},
		},
		{
			name:  "mismatched quotes",
			input: `"foo`,
			want:  []string{"foo"},
		},
		{
			name:  "empty mismatched quote",
			input: `"`,
			want:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, TokenizeText(tt.input))
		})
	}
}
