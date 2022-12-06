package slackbot_test

import (
	"context"
	"github.com/clambin/go-common/slackbot"
	"github.com/clambin/go-common/slackbot/client/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestSlackBot_Run(t *testing.T) {
	c := mocks.NewSlackClient(t)
	c.On("GetChannels").Return([]string{"foo", "bar"}, nil)
	ch := make(chan *slack.MessageEvent)
	c.On("GetMessage").Return(ch)
	c.On("GetUserID").Return("123", nil)

	var runWg sync.WaitGroup
	runWg.Add(1)
	c.On("Run", mock.AnythingOfType("*context.cancelCtx")).Run(func(args mock.Arguments) {
		runWg.Done()
	}).Return(nil)

	b := slackbot.New(t.Name(), "some-token", nil)
	b.SlackClient = c

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, b.Run(ctx))
	}()

	var wg2 sync.WaitGroup
	wg2.Add(1)
	c.On("Send", "foo", mock.AnythingOfType("[]slack.Attachment")).Run(func(args mock.Arguments) {
		defer wg2.Done()
		assertSlackMessage(t, args, "foo", "", t.Name())
	}).Return(nil).Once()

	ch <- &slack.MessageEvent{Msg: slack.Msg{
		Name: "foo", User: "321", Channel: "foo", Text: "<@123> version",
	}}

	wg2.Wait()
	runWg.Wait()
	cancel()
	wg.Wait()
}

func TestSlackBot_Send(t *testing.T) {
	c := mocks.NewSlackClient(t)
	b := slackbot.New(t.Name(), "some-token", nil)
	b.SlackClient = c

	var wg2 sync.WaitGroup
	wg2.Add(1)
	c.On("Send", "bar", mock.AnythingOfType("[]slack.Attachment")).Run(func(args mock.Arguments) {
		defer wg2.Done()
		assertSlackMessage(t, args, "bar", "hello", "world")
	}).Return(nil).Once()

	err := b.Send("bar", []slack.Attachment{{
		Color: "good",
		Title: "hello",
		Text:  "world",
	}})
	assert.NoError(t, err)

	wg2.Wait()
}

func TestSlackBot_Commands(t *testing.T) {
	c := mocks.NewSlackClient(t)
	c.On("GetChannels").Return([]string{"foo", "bar"}, nil)
	ch := make(chan *slack.MessageEvent)
	c.On("GetMessage").Return(ch)
	c.On("GetUserID").Return("123", nil)
	c.On("Run", mock.AnythingOfType("*context.cancelCtx")).Return(nil).Maybe()

	b := slackbot.New("command-test", "some-token", map[string]slackbot.CommandFunc{
		"foo": func(_ context.Context, _ ...string) []slack.Attachment {
			return []slack.Attachment{{Color: "good", Title: "bar", Text: "snafu"}}
		},
		"bar": func(_ context.Context, _ ...string) []slack.Attachment {
			return []slack.Attachment{}
		},
	})
	b.SlackClient = c

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, b.Run(ctx))
	}()

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

			var wg2 sync.WaitGroup
			wg2.Add(1)
			c.On("Send", "foo", mock.AnythingOfType("[]slack.Attachment")).Run(func(args mock.Arguments) {
				defer wg2.Done()
				assertSlackMessage(t, args, "foo", tt.title, tt.text)
			}).Return(nil).Once()

			ch <- &slack.MessageEvent{Msg: slack.Msg{
				Name: "foo", User: "321", Channel: "foo", Text: "<@123> " + tt.command,
			}}

			wg2.Wait()
		})
	}

	cancel()
	wg.Wait()
}

func assertSlackMessage(t *testing.T, args mock.Arguments, channel, title, text string) {
	t.Helper()
	require.Len(t, args, 2)
	ch, ok := args[0].(string)
	require.True(t, ok)
	assert.Equal(t, channel, ch)
	if text != "" {
		attachments, ok := args[1].([]slack.Attachment)
		require.True(t, ok)
		require.Len(t, attachments, 1)
		assert.Equal(t, title, attachments[0].Title)
		assert.Equal(t, text, attachments[0].Text)
	}
}
