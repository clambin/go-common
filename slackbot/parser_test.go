package slackbot

import (
	"github.com/clambin/go-common/slackbot/client/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseText(t *testing.T) {
	var output []string

	output = tokenizeText("Hello world")
	if assert.Len(t, output, 2) {
		assert.Equal(t, "Hello", output[0])
		assert.Equal(t, "world", output[1])
	}

	output = tokenizeText("He said \"Hello world\"")
	if assert.Len(t, output, 3) {
		assert.Equal(t, "He", output[0])
		assert.Equal(t, "said", output[1])
		assert.Equal(t, "Hello world", output[2])
	}

	output = tokenizeText("")
	assert.Len(t, output, 0)

	output = tokenizeText("\"Hello world\"")
	if assert.Len(t, output, 1) {
		assert.Equal(t, "Hello world", output[0])
	}

	output = tokenizeText("\"\"")
	if assert.Len(t, output, 1) {
		assert.Equal(t, "", output[0])
	}

	output = tokenizeText("\"")
	assert.Len(t, output, 0)
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

	c := mocks.NewSlackClient(t)
	c.On("GetUserID").Return("123", nil)
	b := SlackBot{SlackClient: c}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, args := b.parseCommand(tt.input)
			assert.Equal(t, tt.command, command)
			assert.Equal(t, len(tt.args), len(args))
		})
	}
}
