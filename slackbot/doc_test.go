package slackbot_test

import (
	"context"
	"github.com/clambin/go-common/slackbot"
	"github.com/slack-go/slack"
)

func Example() {
	b := slackbot.New("my-slack-token", slackbot.WithCommands(map[string]slackbot.Handler{
		"hello": func(_ context.Context, _ ...string) []slack.Attachment {
			return []slack.Attachment{{Color: "good", Text: "General Kenobi!"}}
		},
	}))

	_ = b.Run(context.Background())
}

func ExampleCommand_AddCommand() {
	b := slackbot.New("my-slack-token")
	b.Commands = &slackbot.Command{}
	b.Commands.Add("hello", func(ctx context.Context, s ...string) []slack.Attachment {
		return []slack.Attachment{{Color: "good", Text: "General Kenobi!"}}
	})

	subCmd := &slackbot.Command{}
	subCmd.Add("foo", func(ctx context.Context, s ...string) []slack.Attachment {
		return []slack.Attachment{{Text: "foo"}}
	})
	subCmd.Add("bar", func(ctx context.Context, s ...string) []slack.Attachment {
		return []slack.Attachment{{Text: "bar"}}
	})
	b.Commands.AddCommand("say", subCmd)

	_ = b.Run(context.Background())
}
