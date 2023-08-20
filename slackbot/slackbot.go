package slackbot

import (
	"context"
	"fmt"
	"github.com/slack-go/slack"
	"log/slog"
	"regexp"
	"strings"
	"sync"
)

// SlackBot connects to Slack through Slack's Bot integration.
type SlackBot struct {
	client        *slackClient
	name          string
	commands      map[string]CommandFunc
	commandRunner *commandRunner
	logger        *slog.Logger
}

// CommandFunc signature for command callback functions
type CommandFunc func(ctx context.Context, args ...string) []slack.Attachment

// New creates a new slackbot
func New(slackToken string, options ...Option) *SlackBot {
	b := &SlackBot{
		name:     "slackbot",
		commands: make(map[string]CommandFunc),
		logger:   slog.Default(),
	}
	b.commands["help"] = b.doHelp
	b.commands["version"] = b.doVersion

	for _, option := range options {
		option(b)
	}

	b.client = newSlackClient(slackToken, b.logger)
	b.commandRunner = &commandRunner{commands: b.commands}

	return b
}

// Register adds a new command
func (b *SlackBot) Register(name string, command CommandFunc) {
	b.commandRunner.Register(name, command)
}

// Run the slackbot
func (b *SlackBot) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); b.client.Run(ctx) }()
	for {
		select {
		case message := <-b.client.GetMessage():
			if err := b.processMessage(ctx, message); err != nil {
				b.logger.Error("failed to process message", "err", err)
			}
		case <-ctx.Done():
			wg.Wait()
			return nil
		}
	}
}

func (b *SlackBot) processMessage(ctx context.Context, message *slack.MessageEvent) error {
	b.logger.Debug("message received",
		"user.id", message.User,
		"user.name", message.Username,
		"text", message.Text,
	)

	command, args, err := b.parseCommand(message.Text)
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	if command == "" {
		return nil
	}

	b.logger.Debug("running command", "command", command, "args", args)
	output := b.commandRunner.Do(ctx, command, args...)

	err = b.Send(message.Channel, output)
	if err != nil {
		err = fmt.Errorf("post to Slack failed: %w", err)
	}
	return err
}

func (b *SlackBot) Send(channel string, attachments []slack.Attachment) error {
	channelIDs := []string{channel}
	if channel == "" {
		var err error
		if channelIDs, err = b.client.GetChannels(); err != nil {
			return err
		}
	}

	for _, c := range channelIDs {
		if err := b.client.Send(c, attachments); err != nil {
			return err
		}
	}
	return nil
}

func (b *SlackBot) parseCommand(input string) (string, []string, error) {
	userID, err := b.client.GetUserID()
	if err != nil {
		return "", nil, err
	}
	words := tokenizeText(input)
	if len(words) == 0 || words[0] != "<@"+userID+">" {
		return "", nil, nil
	}
	return words[1], words[2:], nil
}

func tokenizeText(input string) []string {
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

func (b *SlackBot) doHelp(_ context.Context, _ ...string) []slack.Attachment {
	return []slack.Attachment{{
		Color: "good",
		Title: "supported commands",
		Text:  strings.Join(b.commandRunner.GetCommands(), ", "),
	}}
}

func (b *SlackBot) doVersion(_ context.Context, _ ...string) []slack.Attachment {
	return []slack.Attachment{{Color: "good", Text: b.name}}
}
