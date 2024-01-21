package commands

import (
	"context"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	rootCmd := New()
	helpCmd := New()
	rootCmd.Register("help", helpCmd)

	rootCmd.Register("list", ActionFunc(func(ctx context.Context, args ...string) []slack.Attachment {
		return []slack.Attachment{{Text: "list run"}}
	}))
	rootCmd.Register("set", ActionFunc(func(ctx context.Context, args ...string) []slack.Attachment {
		return []slack.Attachment{{Text: "value set: " + strings.Join(args, ", ")}}
	}))
	helpCmd.Register("set", ActionFunc(func(ctx context.Context, args ...string) []slack.Attachment {
		return []slack.Attachment{{Text: "set key=val"}}
	}))

	tests := []struct {
		name string
		args []string
		want []slack.Attachment
	}{
		{
			name: "command with args",
			args: []string{"set", "a=b"},
			want: []slack.Attachment{{Text: "value set: a=b"}},
		},
		{
			name: "command without args",
			args: []string{"set"},
			want: []slack.Attachment{{Text: "value set: "}},
		},
		{
			name: "nested command",
			args: []string{"help", "set"},
			want: []slack.Attachment{{Text: "set key=val"}},
		},
		{
			name: "invalid command",
			args: []string{"`invalid"},
			want: []slack.Attachment{{
				Color: "bad",
				Title: "invalid command",
				Text:  "supported commands: help, list, set",
			}},
		},
		{
			name: "invalid command with junk",
			args: []string{"`invalid", "help"},
			want: []slack.Attachment{{
				Color: "bad",
				Title: "invalid command",
				Text:  "supported commands: help, list, set",
			}},
		},
		{
			name: "invalid subcommand",
			args: []string{"help", "invalid"},
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

			output := rootCmd.Do(context.Background(), tt.args...)
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
