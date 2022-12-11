package slackbot

import (
	"context"
	"github.com/clambin/go-common/slackbot/client"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestSlackBot_Run(t *testing.T) {
	b := New(t.Name(), "some-token", nil)
	c := newSlackClient("123", []string{"foo"})
	b.SlackClient = c

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, b.Run(ctx))
	}()

	c.fromSlack <- &slack.MessageEvent{Msg: slack.Msg{
		Name: "foo", User: "321", Channel: "foo", Text: "<@123> version",
	}}

	msg := <-c.toSlack
	require.Len(t, msg.attachments, 1)
	assert.Equal(t, t.Name(), msg.attachments[0].Text)

	cancel()
	wg.Wait()
}

func TestSlackBot_Send(t *testing.T) {
	channels := []string{"foo", "bar"}
	c := newSlackClient("123", channels)
	b := New(t.Name(), "some-token", nil)
	b.SlackClient = c

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() { defer wg.Done(); c.Run(ctx) }()

	err := b.Send("bar", []slack.Attachment{{
		Color: "good",
		Title: "hello",
		Text:  "world",
	}})

	sent := <-c.toSlack

	assert.NoError(t, err)
	assert.Equal(t, "bar", sent.channel)
	require.Len(t, sent.attachments, 1)
	assert.Equal(t, "good", sent.attachments[0].Color)
	assert.Equal(t, "hello", sent.attachments[0].Title)
	assert.Equal(t, "world", sent.attachments[0].Text)

	err = b.Send("", []slack.Attachment{{
		Color: "good",
		Title: "hello",
		Text:  "world",
	}})
	assert.NoError(t, err)

	for range channels {
		sent = <-c.toSlack
		assert.Contains(t, channels, sent.channel)
	}
	cancel()
	wg.Wait()
}

func TestSlackBot_Commands(t *testing.T) {
	b := New("command-test", "some-token", map[string]CommandFunc{
		"foo": func(_ context.Context, _ ...string) []slack.Attachment {
			return []slack.Attachment{{Color: "good", Title: "bar", Text: "snafu"}}
		},
		"bar": func(_ context.Context, _ ...string) []slack.Attachment {
			return []slack.Attachment{}
		},
	})
	c := newSlackClient("123", []string{"foo", "bar"})
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
			command: "invalid command",
			text:    "unrecognized command",
		},
		{
			command: "bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {

			c.fromSlack <- &slack.MessageEvent{Msg: slack.Msg{
				Name: "foo", User: "321", Channel: "foo", Text: "<@123> " + tt.command,
			}}

			msg := <-c.toSlack
			if tt.text != "" {
				require.Len(t, msg.attachments, 1)
				assert.Equal(t, tt.title, msg.attachments[0].Title)
				assert.Equal(t, tt.text, msg.attachments[0].Text)
			}
		})
	}

	cancel()
	wg.Wait()
}

type slackMessage struct {
	channel     string
	attachments []slack.Attachment
}

type slackClient struct {
	userId    string
	channels  []string
	fromSlack chan *slack.MessageEvent
	toSlack   chan *slackMessage
}

func newSlackClient(userId string, channels []string) *slackClient {
	return &slackClient{
		userId:    userId,
		channels:  channels,
		fromSlack: make(chan *slack.MessageEvent, 10),
		toSlack:   make(chan *slackMessage, 10),
	}
}

func (s slackClient) Run(ctx context.Context) {
	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		}
	}
}

func (s slackClient) Send(channel string, attachments []slack.Attachment) (err error) {
	s.toSlack <- &slackMessage{
		channel:     channel,
		attachments: attachments,
	}
	return nil
}

func (s slackClient) GetMessage() chan *slack.MessageEvent {
	return s.fromSlack
}

func (s slackClient) GetChannels() (channelIDs []string, err error) {
	return s.channels, nil
}

func (s slackClient) GetUserID() (string, error) {
	return s.userId, nil
}

var _ client.SlackClient = &slackClient{}
