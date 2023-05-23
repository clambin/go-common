package slackbot_test

import (
	"context"
	"github.com/clambin/go-common/slackbot"
	"github.com/slack-go/slack"
)

func Example() {
	b := slackbot.New("example", "my-slack-token", map[string]slackbot.CommandFunc{
		"hello": func(_ context.Context, _ ...string) []slack.Attachment {
			return []slack.Attachment{{
				Color: "good",
				Text:  "General Kenobi!",
			}}
		},
	}, nil)

	b.Run(context.Background())
}
