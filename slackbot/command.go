package slackbot

import (
	"context"
	"github.com/slack-go/slack"
	"slices"
	"strings"
	"sync"
)

// A Handler executes a Command and returns messages to be posted to Slack.
type Handler interface {
	Handle(context.Context, ...string) []slack.Attachment
}
type HandlerFunc func(context.Context, ...string) []slack.Attachment

func (f HandlerFunc) Handle(ctx context.Context, args ...string) []slack.Attachment {
	return f(ctx, args...)
}

type Commands map[string]Handler

var _ Handler = &CommandGroup{}

type CommandGroup struct {
	subCommands Commands
	lock        sync.RWMutex
}

func NewCommandGroup(commands Commands) *CommandGroup {
	var g CommandGroup
	g.Add(commands)
	return &g
}

func (c *CommandGroup) Add(commands Commands) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.subCommands == nil {
		c.subCommands = commands
		return
	}

	for verb, handler := range commands {
		c.subCommands[verb] = handler
	}
}

// GetCommands returns all commands supported by the Command.
func (c *CommandGroup) GetCommands() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	commands := make([]string, 0, len(c.subCommands))
	for verb := range c.subCommands {
		commands = append(commands, verb)
	}
	slices.Sort(commands)
	return commands
}

func (c *CommandGroup) Handle(ctx context.Context, args ...string) []slack.Attachment {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if subCmd, subArgs := split(args...); subCmd != "" {
		if subCommand, ok := c.subCommands[subCmd]; ok {
			return subCommand.Handle(ctx, subArgs...)
		}
	}

	return []slack.Attachment{{
		Title: "invalid command",
		Color: "bad",
		Text:  "supported commands: " + strings.Join(c.GetCommands(), ", "),
	}}
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
