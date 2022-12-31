package slackbot

import (
	"context"
	"github.com/clambin/go-common/slackbot/client"
	"github.com/slack-go/slack"
	"golang.org/x/exp/slog"
	"regexp"
	"strings"
	"sync"
)

// SlackBot connects to Slack through Slack's Bot integration.
type SlackBot struct {
	client.SlackClient
	name     string
	commands *commands
	lock     sync.RWMutex
}

// CommandFunc signature for command callback functions
type CommandFunc func(ctx context.Context, args ...string) []slack.Attachment

// New creates a new slackbot
func New(name string, slackToken string, commands map[string]CommandFunc) *SlackBot {
	b := SlackBot{
		name:        name,
		commands:    newCommands(),
		SlackClient: client.New(slackToken),
	}

	b.Register("help", b.doHelp)
	b.Register("version", b.doVersion)

	for cmd, callbackFunction := range commands {
		b.commands.Register(cmd, callbackFunction)
	}

	return &b
}

// Register adds a new command
func (b *SlackBot) Register(name string, command CommandFunc) {
	b.commands.Register(name, command)
}

// Run the slackbot
func (b *SlackBot) Run(ctx context.Context) (err error) {
	go b.SlackClient.Run(ctx)

	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case message := <-b.GetMessage():
			command, args := b.parseCommand(message.Text)
			if command != "" {
				err = b.Send(message.Channel, b.commands.Do(ctx, command, args...))
			}
		}
		if err != nil {
			slog.Error("failed to post message on Slack", err)
		}
	}
	return
}

func (b *SlackBot) Send(channel string, attachments []slack.Attachment) (err error) {
	channels := []string{channel}
	if channel == "" {
		if channels, err = b.GetChannels(); err != nil {
			return err
		}
	}

	for _, c := range channels {
		if err = b.SlackClient.Send(c, attachments); err != nil {
			break
		}
	}
	return err
}

func (b *SlackBot) parseCommand(input string) (command string, args []string) {
	userID, err := b.SlackClient.GetUserID()
	if err != nil {
		panic("we received a message but user ID not yet set up? " + err.Error())
	}
	words := tokenizeText(input)
	if len(words) > 1 && words[0] == "<@"+userID+">" {
		command = words[1]
		args = words[2:]
	}
	return
}

func tokenizeText(input string) (output []string) {
	cleanInput := input
	for _, quote := range []string{"“", "”", "'"} {
		cleanInput = strings.ReplaceAll(cleanInput, quote, "\"")
	}
	r := regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
	output = r.FindAllString(cleanInput, -1)

	for index, word := range output {
		output[index] = strings.Trim(word, "\"")
	}
	return
}

func (b *SlackBot) doHelp(_ context.Context, _ ...string) []slack.Attachment {
	return []slack.Attachment{{
		Color: "good",
		Title: "supported commands",
		Text:  strings.Join(b.commands.GetCommands(), ", "),
	}}
}

func (b *SlackBot) doVersion(_ context.Context, _ ...string) []slack.Attachment {
	return []slack.Attachment{{Color: "good", Text: b.name}}
}
