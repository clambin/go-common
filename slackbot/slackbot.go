package slackbot

import (
	"context"
	"fmt"
	"github.com/slack-go/slack"
	"log/slog"
	"strings"
	"sync"
)

// SlackBot connects to Slack through Slack's Bot integration.
type SlackBot struct {
	client   *slackClient
	name     string
	commands *Command
	logger   *slog.Logger
}

// New creates a new slackbot
func New(slackToken string, options ...Option) *SlackBot {
	b := &SlackBot{
		name:     "slackbot",
		commands: &Command{},
		logger:   slog.Default(),
	}
	b.commands.Add("help", b.doHelp)
	b.commands.Add("version", b.doVersion)

	for _, option := range options {
		option(b)
	}

	b.client = newSlackClient(slackToken, b.logger)

	return b
}

// Register adds a new command
func (b *SlackBot) Register(name string, command Handler) {
	b.commands.Add(name, command)
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

	args, err := b.parseCommand(message.Text)
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	if len(args) == 0 {
		return nil
	}

	b.logger.Debug("running command", "args", args)
	output := b.commands.handle(ctx, args...)

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

func (b *SlackBot) parseCommand(input string) ([]string, error) {
	userID, err := b.client.GetUserID()
	if err != nil {
		return nil, err
	}
	words := TokenizeText(input)
	if len(words) == 0 || words[0] != "<@"+userID+">" {
		return nil, nil
	}
	if len(words) == 1 {
		return nil, nil
	}
	return words[1:], nil
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
