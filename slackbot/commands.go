package slackbot

import (
	"context"
	"github.com/slack-go/slack"
	"sort"
	"sync"
)

type commandRunner struct {
	commands map[string]CommandFunc
	lock     sync.RWMutex
}

func newCommandRunner() *commandRunner {
	return &commandRunner{
		commands: make(map[string]CommandFunc),
	}
}

func (c *commandRunner) Register(command string, callBack CommandFunc) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.commands[command] = callBack
}

func (c *commandRunner) Do(ctx context.Context, command string, args ...string) []slack.Attachment {
	c.lock.RLock()
	defer c.lock.RUnlock()
	callBack, ok := c.commands[command]
	if !ok {
		return []slack.Attachment{{
			Color: "red",
			Text:  "unrecognized command",
		}}
	}
	return callBack(ctx, args...)
}
func (c *commandRunner) GetCommands() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	cmds := make([]string, 0, len(c.commands))
	for command := range c.commands {
		cmds = append(cmds, command)
	}
	sort.Strings(cmds)
	return cmds
}
