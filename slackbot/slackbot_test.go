package slackbot

import (
	"context"
	"github.com/clambin/go-common/slackbot/internal/mocks"
	slack_client "github.com/clambin/go-common/slackbot/internal/slack-client"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSlackBot_Run(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	ev := make(chan *slack.MessageEvent)

	f := mocks.NewSlackClient(t)
	f.EXPECT().GetMessage().Return(ev)
	f.EXPECT().Run(ctx)
	f.EXPECT().GetUserID().Return("123", nil)

	b := New("some-token")
	b.client = f

	ch := make(chan error)
	go func() {
		ch <- b.Run(ctx)
	}()

	ev <- &slack.MessageEvent{
		Msg: slack.Msg{Text: "help"},
	}

	time.Sleep(time.Second)

	cancel()
	assert.NoError(t, <-ch)
}

func TestSlackBot_Send(t *testing.T) {
	tests := []struct {
		name        string
		channel     string
		channelsErr error
		message     []slack.Attachment
		sendErr     error
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:        "single channel",
			channel:     "foo",
			channelsErr: nil,
			message: []slack.Attachment{{
				Color: "good",
				Title: "hello",
				Text:  "world",
			}},
			sendErr: nil,
			wantErr: assert.NoError,
		},
		{
			name:        "broadcast",
			channel:     "",
			channelsErr: nil,
			message: []slack.Attachment{{
				Color: "good",
				Title: "hello",
				Text:  "world",
			}},
			sendErr: nil,
			wantErr: assert.NoError,
		},
		{
			name:        "GetChannels fails",
			channel:     "",
			channelsErr: slack_client.ErrNotConnected,
			message: []slack.Attachment{{
				Color: "good",
				Title: "hello",
				Text:  "world",
			}},
			sendErr: nil,
			wantErr: assert.Error,
		},
		{
			name:        "send fails",
			channel:     "foo",
			channelsErr: nil,
			message: []slack.Attachment{{
				Color: "good",
				Title: "hello",
				Text:  "world",
			}},
			sendErr: slack_client.ErrNotConnected,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := mocks.NewSlackClient(t)
			f.EXPECT().GetChannels().Return([]string{"foo"}, tt.channelsErr).Maybe()

			b := New("some-token")
			b.client = f

			f.EXPECT().Send("foo", tt.message).Return(tt.sendErr).Maybe()

			err := b.Send(tt.channel, tt.message)
			tt.wantErr(t, err)
		})
	}
}

func TestSlackBot_processMessage(t *testing.T) {
	tests := []struct {
		name      string
		message   *slack.MessageEvent
		userIDErr error
		want      []slack.Attachment
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name:    "empty",
			message: &slack.MessageEvent{Msg: slack.Msg{Text: ""}},
			wantErr: assert.NoError,
		},
		{
			name:    "chatter",
			message: &slack.MessageEvent{Msg: slack.Msg{Text: "chatter"}},
			wantErr: assert.NoError,
		},
		{
			name:    "version",
			message: &slack.MessageEvent{Msg: slack.Msg{Text: "<@123> version"}},
			want:    []slack.Attachment{{Color: "good", Text: "name"}},
			wantErr: assert.NoError,
		},
		{
			name:    "help",
			message: &slack.MessageEvent{Msg: slack.Msg{Text: "<@123> help"}},
			want:    []slack.Attachment{{Color: "good", Title: "supported commands", Text: "foo, help, version"}},
			wantErr: assert.NoError,
		},
		{
			name:    "Command",
			message: &slack.MessageEvent{Msg: slack.Msg{Text: "<@123> foo"}},
			want:    []slack.Attachment{{Color: "good", Title: "bar", Text: "snafu"}},
			wantErr: assert.NoError,
		},
		{
			name:      "error",
			message:   &slack.MessageEvent{Msg: slack.Msg{Text: "<@123> version"}},
			userIDErr: slack_client.ErrNotConnected,
			//want:      []slack.Attachment{{Color: "good", Text: "name"}},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			c := mocks.NewSlackClient(t)
			c.EXPECT().GetUserID().Return("123", tt.userIDErr).Maybe()
			c.EXPECT().GetChannels().Return([]string{"foo"}, nil).Maybe()
			if tt.want != nil {
				c.EXPECT().Send("foo", tt.want).Return(nil)
			}

			b := New("some-token", WithName("name"), WithCommands(map[string]Handler{
				"foo": HandlerFunc(func(_ context.Context, _ ...string) []slack.Attachment {
					return []slack.Attachment{{Color: "good", Title: "bar", Text: "snafu"}}
				}),
			}),
			)
			b.client = c

			tt.wantErr(t, b.processMessage(ctx, tt.message))

		})
	}
}

func TestSlackbot_parseCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		userIdError error
		want        []string
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:        "not connected",
			input:       "<@123>",
			userIdError: slack_client.ErrNotConnected,
			wantErr:     assert.Error,
		},
		{
			name:    "empty string",
			wantErr: assert.NoError,
		},
		{
			name:    "chatter",
			input:   "hello world",
			wantErr: assert.NoError,
		},
		{
			name:    "no Command",
			input:   "<@123>",
			wantErr: assert.NoError,
		},
		{
			name:    "single Command",
			input:   "<@123> version",
			want:    []string{"version"},
			wantErr: assert.NoError,
		},
		{
			name:    "Command arguments",
			input:   "<@123> foo bar snafu",
			want:    []string{"foo", "bar", "snafu"},
			wantErr: assert.NoError,
		},
		{
			name:    "arguments with quotes",
			input:   `<@123> foo "bar snafu"`,
			want:    []string{"foo", "bar snafu"},
			wantErr: assert.NoError,
		},
		{
			name:    "fancy quotes",
			input:   `<@123> foo “bar snafu“ foobar`,
			want:    []string{"foo", "bar snafu", "foobar"},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := mocks.NewSlackClient(t)
			c.EXPECT().GetUserID().Return("123", tt.userIdError)

			b := New("some-token")
			b.client = c

			args, err := b.parseCommand(tt.input)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, args)
		})
	}
}

func Test_tokenizeText(t *testing.T) {
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

			assert.Equal(t, tt.want, tokenizeText(tt.input))
		})
	}
}
