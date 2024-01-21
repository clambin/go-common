package slackbot

import (
	"context"
	"github.com/clambin/go-common/slackbot/internal/connector"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestSlackBot_Run(t *testing.T) {
	b := New("some-token")
	f := connector.NewFakeConnector()
	b.client.connector = f

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.Run(ctx)
	}()

	f.Connect()
	f.IncomingMessage("123", "<@123> version")

	msg := <-f.ToSlack
	assert.Equal(t, connector.PostedMessage{ChannelID: "123", Attachments: []slack.Attachment{{Color: "good", Text: "slackbot"}}}, msg)

	cancel()
	wg.Wait()
}

func TestSlackBot_Send(t *testing.T) {
	b := New("some-token")
	f := connector.NewFakeConnector()
	b.client.connector = f

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() { defer wg.Done(); b.Run(ctx) }()

	err := b.Send("bar", []slack.Attachment{{
		Color: "good",
		Title: "hello",
		Text:  "world",
	}})
	require.NoError(t, err)

	assert.Equal(t, connector.PostedMessage{
		ChannelID: "bar",
		Attachments: []slack.Attachment{{
			Color: "good",
			Title: "hello",
			Text:  "world",
		}},
	}, <-f.ToSlack)

	err = b.Send("", []slack.Attachment{{
		Color: "good",
		Title: "hello",
		Text:  "world",
	}})
	require.NoError(t, err)

	assert.Equal(t, connector.PostedMessage{
		ChannelID: "123",
		Attachments: []slack.Attachment{{
			Color: "good",
			Title: "hello",
			Text:  "world",
		}},
	}, <-f.ToSlack)

	cancel()
	wg.Wait()
}

func TestSlackBot_Commands(t *testing.T) {
	b := New("some-token",
		WithCommands(map[string]CommandFunc{
			"foo": func(_ context.Context, _ ...string) []slack.Attachment {
				return []slack.Attachment{{Color: "good", Title: "bar", Text: "snafu"}}
			},
		}),
		WithName("command-test"),
	)
	b.Register("bar", func(_ context.Context, _ ...string) []slack.Attachment {
		return []slack.Attachment{}
	})

	f := connector.NewFakeConnector()
	b.client.connector = f

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.Run(ctx)
	}()

	f.Connect()

	tests := []struct {
		command string
		title   string
		text    string
	}{
		{
			command: "version",
			text:    "command-test",
		},
		{
			command: "help",
			title:   "supported commands",
			text:    "bar, foo, help, version",
		},
		{
			command: "foo",
			title:   "bar",
			text:    "snafu",
		},
		{
			command: "foo bar",
			title:   "bar",
			text:    "snafu",
		},
		{
			command: "bar",
		},
		{
			command: "invalid command",
			text:    "unrecognized command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {

			f.IncomingMessage("foo", "<@123> "+tt.command)

			msg := <-f.ToSlack
			if tt.text != "" {
				require.Len(t, msg.Attachments, 1)
				assert.Equal(t, tt.title, msg.Attachments[0].Title)
				assert.Equal(t, tt.text, msg.Attachments[0].Text)
			}
		})
	}

	cancel()
	wg.Wait()
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
