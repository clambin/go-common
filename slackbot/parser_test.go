package slackbot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseText(t *testing.T) {
	tests := []struct {
		input  string
		output []string
	}{
		{input: `Hello world`, output: []string{"Hello", "world"}},
		{input: `He said "Hello world"`, output: []string{"He", "said", "Hello world"}},
		{input: `"Hello world"`, output: []string{"Hello world"}},
		{input: `”Hello world”`, output: []string{"Hello world"}},
		{input: `""`, output: []string{""}},
		{input: `"`},
		{input: ``},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			output := tokenizeText(tt.input)
			assert.Equal(t, tt.output, output)
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		command string
		args    []string
	}{
		{
			name: "empty string",
		},
		{
			name:  "chatter",
			input: "hello world",
		},
		{
			name:    "single command",
			input:   "<@123> version",
			command: "version",
		},
		{
			name:    "command arguments",
			input:   "<@123> foo bar snafu",
			command: "foo",
			args:    []string{"bar", "snafu"},
		},
		{
			name:    "arguments with quotes",
			input:   `<@123> foo "bar snafu"`,
			command: "foo",
			args:    []string{"bar snafu"},
		},
		{
			name:    "fancy quotes",
			input:   `<@123> foo “bar snafu“ foobar`,
			command: "foo",
			args:    []string{"bar snafu", "foobar"},
		},
	}

	b := New("some-token")
	b.client.userID = "123"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, args, err := b.parseCommand(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.command, command)
			assert.Equal(t, len(tt.args), len(args))
		})
	}
}
