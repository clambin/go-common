package slackbot

import (
	"context"
	"github.com/slack-go/slack"
	"slices"
	"strings"
)

// A Handler executes a command and returns messages to be posted to Slack.
type Handler interface {
	Handle(context.Context, ...string) []slack.Attachment
}

// HandlerFunc is an adapter that allows a function to be used as a Handler
type HandlerFunc func(context.Context, ...string) []slack.Attachment

// Handle calls f(ctx, args)
func (f HandlerFunc) Handle(ctx context.Context, args ...string) []slack.Attachment {
	return f(ctx, args...)
}

var _ Handler = &Commands{}

// Commands is a map of verb/Handler pairs.
//
// Note that Commands itself implements the Handler interface. This allows nested command structures to be built:
//
//	Commands
//	"foo"    -> handler
//	"bar"    -> Commands
//	            "snafu"    -> handler
//
// This will support the commands "foo" and "bar snafu"
type Commands map[string]Handler

// Handle processes the incoming command. The first arg is considered the verb. If it matches a supported command, its
// corresponding handler is called, passing the remaining arguments.
//
// If the verb is not supported, An attachment is reported will all supported commands
func (c Commands) Handle(ctx context.Context, args ...string) []slack.Attachment {
	if subCmd, subArgs := split(args...); subCmd != "" {
		if subCommand, ok := c[subCmd]; ok {
			return subCommand.Handle(ctx, subArgs...)
		}
	}

	return []slack.Attachment{{
		Title: "invalid command",
		Color: "bad",
		Text:  "supported commands: " + strings.Join(c.GetCommands(), ", "),
	}}
}

// GetCommands returns a sorted list of all supported commands.
func (c Commands) GetCommands() []string {
	commands := make([]string, 0, len(c))
	for verb := range c {
		commands = append(commands, verb)
	}
	slices.Sort(commands)
	return commands
}

// Add adds one or more commands.
func (c Commands) Add(commands Commands) {
	for verb, handler := range commands {
		c[verb] = handler
	}
}

func split(args ...string) (string, []string) {
	if len(args) == 0 {
		return "", nil
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return args[0], args[1:]
}
