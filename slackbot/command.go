package slackbot

import (
	"context"
	"github.com/slack-go/slack"
	"slices"
	"strings"
	"sync"
)

// A Handler executes a command and returns messages to be posted to Slack.
type Handler func(context.Context, ...string) []slack.Attachment

// A Command holds the set of commands supported by a SlackBot.
//
// In its simplest form, Command contains a set of command names, with a corresponding Handler that executes the command:
//
//	Command:
//	  "foo" -> Handler
//	  "bar" -> Handler
//
// This supports two commands "foo" and "bar".
//
// More complex command structures can be created by using AddCommand to nest Command structures:
//
//	Command:
//	  "foo" -> Command:
//	             "bar"   -> Handler
//	             "snafu" -> Handler
//
// This supports two compound commands, "foo bar" and "foo snafu", each with their own handlers.
type Command struct {
	subCommands map[string]*Command
	handler     Handler
	lock        sync.RWMutex
}

// Add a new command
func (c *Command) Add(verb string, handler Handler) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.subCommands == nil {
		c.subCommands = make(map[string]*Command)
	}
	c.subCommands[verb] = &Command{handler: handler}
}

// AddCommand adds a compound command (i.e. a command that holds several subcommands).
func (c *Command) AddCommand(verb string, command *Command) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.subCommands == nil {
		c.subCommands = make(map[string]*Command)
	}
	c.subCommands[verb] = command
}

// GetCommands returns all commands supported by the Command.
func (c *Command) GetCommands() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	commands := make([]string, 0, len(c.subCommands))
	for verb := range c.subCommands {
		commands = append(commands, verb)
	}
	slices.Sort(commands)
	return commands
}

func (c *Command) handle(ctx context.Context, args ...string) []slack.Attachment {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if subCmd, subArgs := split(args...); subCmd != "" {
		if subCommand, ok := c.subCommands[subCmd]; ok {
			return subCommand.handle(ctx, subArgs...)
		}
	}

	if c.handler == nil {
		return c.invalidCommand()
	}
	return c.handler(ctx, args...)
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

func (c *Command) invalidCommand() []slack.Attachment {
	return []slack.Attachment{{
		Title: "invalid command",
		Color: "bad",
		Text:  "supported commands: " + strings.Join(c.GetCommands(), ", "),
	}}
}
