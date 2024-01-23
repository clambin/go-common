package slackbot

import (
	"context"
	slack_client "github.com/clambin/go-common/slackbot/internal/slack-client"
	"github.com/clambin/go-common/slackbot/mocks"
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
	/*
		ctx, cancel := context.WithCancel(context.Background())
		ev := make(chan *slack.MessageEvent)

		f := mocks.NewSlackClient(t)
		f.EXPECT().GetMessage().Return(ev)
		f.EXPECT().Run(ctx)
		f.EXPECT().GetChannels().Return([]string{"foo", "bar"}, nil)

		b := New("some-token")
		b.client = f

		ch := make(chan struct{})
		go func() { _ = b.Run(ctx); ch <- struct{}{} }()

		f.EXPECT().Send("bar", []slack.Attachment{{
			Color: "good",
			Title: "hello",
			Text:  "world",
		}}).Return(nil).Twice()

		err := b.Send("bar", []slack.Attachment{{
			Color: "good",
			Title: "hello",
			Text:  "world",
		}})
		require.NoError(t, err)

		err = b.Send("", []slack.Attachment{{
			Color: "good",
			Title: "hello",
			Text:  "world",
		}})
		require.NoError(t, err)

		cancel()
		<-ch

	*/
}

func TestSlackBot_Commands(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ev := make(chan *slack.MessageEvent)

	f := mocks.NewSlackClient(t)
	f.EXPECT().GetUserID().Return("123", nil)
	f.EXPECT().GetMessage().Return(ev)
	f.EXPECT().Run(ctx)
	f.EXPECT().GetChannels().Return([]string{"bar"}, nil)

	b := New("some-token",
		WithCommands(map[string]Handler{
			"foo": func(_ context.Context, _ ...string) []slack.Attachment {
				return []slack.Attachment{{Color: "good", Title: "bar", Text: "snafu"}}
			},
		}),
		WithName("command-test"),
	)
	b.Commands.Add("bar", func(_ context.Context, _ ...string) []slack.Attachment {
		return []slack.Attachment{}
	})
	b.client = f

	ch := make(chan error)
	go func() {
		ch <- b.Run(ctx)
	}()

	tests := []struct {
		command string
		want    []slack.Attachment
		title   string
		text    string
	}{
		{
			command: "version",
			want: []slack.Attachment{{
				Color: "good",
				Text:  "command-test",
			}},
		},
		{
			command: "help",
			want: []slack.Attachment{{
				Color: "good",
				Title: "supported commands",
				Text:  "bar, foo, help, version",
			}},
		},
		{
			command: "foo",
			want: []slack.Attachment{{
				Color: "good",
				Title: "bar",
				Text:  "snafu",
			}},
		},
		{
			command: "foo bar",
			want: []slack.Attachment{{
				Color: "good",
				Title: "bar",
				Text:  "snafu",
			}},
		},
		{
			command: "bar",
			want:    []slack.Attachment{},
		},
		{
			command: "invalid command",
			want: []slack.Attachment{{
				Color: "bad",
				Title: "invalid command",
				Text:  "supported commands: bar, foo, help, version",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			f.EXPECT().Send("bar", tt.want).Return(nil)
			ev <- &slack.MessageEvent{Msg: slack.Msg{Text: "<@123> " + tt.command}}

		})
	}

	cancel()
	assert.NoError(t, <-ch)
}

func TestParseCommand(t *testing.T) {
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
			name:    "no command",
			input:   "<@123>",
			wantErr: assert.NoError,
		},
		{
			name:    "single command",
			input:   "<@123> version",
			want:    []string{"version"},
			wantErr: assert.NoError,
		},
		{
			name:    "command arguments",
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
