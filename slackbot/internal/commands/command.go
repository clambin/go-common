package commands

import (
	"context"
	"github.com/slack-go/slack"
	"regexp"
	"slices"
	"strings"
	"sync"
)

type Doer interface {
	Do(context.Context, ...string) []slack.Attachment
}

type Action func(ctx context.Context, args ...string) []slack.Attachment
type ActionFunc Action

func (f ActionFunc) Do(ctx context.Context, args ...string) []slack.Attachment {
	return f(ctx, args...)
}

type Command struct {
	subCommands map[string]Doer
	doer        Doer
	lock        sync.RWMutex
}

func New() *Command {
	return &Command{
		subCommands: make(map[string]Doer),
	}
}

func (c *Command) Register(verb string, action Doer) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.subCommands[verb] = &Command{
		subCommands: make(map[string]Doer),
		doer:        action,
	}
}

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

func (c *Command) Do(ctx context.Context, args ...string) []slack.Attachment {
	c.lock.RLock()
	defer c.lock.RUnlock()

	subCmd, subArgs := split(args...)
	if subCmd != "" {
		if subCommand, ok := c.subCommands[subCmd]; ok {
			return subCommand.Do(ctx, subArgs...)
		}
	}

	if c.doer == nil {
		return c.invalidCommand()
	}
	return c.doer.Do(ctx, args...)
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

func TokenizeText(input string) []string {
	cleanInput := input
	for _, quote := range []string{"“", "”", "'"} {
		cleanInput = strings.ReplaceAll(cleanInput, quote, "\"")
	}
	r := regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
	output := r.FindAllString(cleanInput, -1)

	for index, word := range output {
		output[index] = strings.Trim(word, "\"")
	}
	return output
}
