package slackbot_test

import (
	"context"
	"fmt"
	"github.com/clambin/go-common/slackbot"
	"github.com/slack-go/slack"
	"strings"
)

func Example_simple() {
	b := slackbot.New("my-slack-token", slackbot.WithCommands(slackbot.Commands{
		"hello": slackbot.HandlerFunc(func(_ context.Context, _ ...string) []slack.Attachment {
			return []slack.Attachment{{Color: "good", Text: "General Kenobi!"}}
		}),
	}))

	fmt.Println("Commands: " + strings.Join(b.GetCommands(), ", "))
	// Output: Commands: hello, help, version
}

func Example_nested() {
	b := slackbot.New("my-slack-token", slackbot.WithCommands(slackbot.Commands{
		"hello": slackbot.HandlerFunc(func(ctx context.Context, s ...string) []slack.Attachment {
			return []slack.Attachment{{Color: "good", Text: "General Kenobi!"}}
		}),
		"say": &slackbot.Commands{
			"foo": slackbot.HandlerFunc(func(ctx context.Context, s ...string) []slack.Attachment {
				return []slack.Attachment{{Text: "foo"}}
			}),
			"bar": slackbot.HandlerFunc(func(ctx context.Context, s ...string) []slack.Attachment {
				return []slack.Attachment{{Text: "bar"}}
			}),
		},
	}))

	fmt.Println("Commands: " + strings.Join(b.GetCommands(), ", "))
	// Output: Commands: hello, help, say, version
}
