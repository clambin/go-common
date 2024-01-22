package slackbot_test

import (
	"context"
	"github.com/clambin/go-common/slackbot"
	"github.com/slack-go/slack"
)

func Example() {
	b := slackbot.New("my-slack-token", slackbot.WithCommands(map[string]slackbot.Handler{
		"hello": func(_ context.Context, _ ...string) []slack.Attachment {
			return []slack.Attachment{{
				Color: "good",
				Text:  "General Kenobi!",
			}}
		},
	}))

	_ = b.Run(context.Background())
}
