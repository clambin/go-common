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
	Commands *Command
	client   *slackClient
	name     string
	logger   *slog.Logger
}

// New creates a new slackbot
func New(slackToken string, options ...Option) *SlackBot {
	b := &SlackBot{
		name:     "slackbot",
		Commands: &Command{},
		logger:   slog.Default(),
	}
	b.Commands.Add("help", b.doHelp)
	b.Commands.Add("version", b.doVersion)

	for _, option := range options {
		option(b)
	}

	b.client = newSlackClient(slackToken, b.logger)

	return b
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
	output := b.Commands.handle(ctx, args...)

	err = b.Send(message.Channel, output)
	if err != nil {
		err = fmt.Errorf("send to Slack failed: %w", err)
	}
	return err
}

func (b *SlackBot) parseCommand(input string) ([]string, error) {
	userID, err := b.client.GetUserID()
	if err != nil {
		return nil, err
	}
	words := tokenizeText(input)
	if len(words) == 0 || words[0] != "<@"+userID+">" {
		return nil, nil
	}
	if len(words) == 1 {
		return nil, nil
	}
	return words[1:], nil
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

// Send posts the provided messages to Slack on the provided channel. If channel is blank,
// the messages will be posted to all channels that the bot has access to.
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

// TODO: this only works for top-level commands. Should this be part of Command functionality?
func (b *SlackBot) doHelp(_ context.Context, _ ...string) []slack.Attachment {
	return []slack.Attachment{{
		Color: "good",
		Title: "supported commands",
		Text:  strings.Join(b.Commands.GetCommands(), ", "),
	}}
}

func (b *SlackBot) doVersion(_ context.Context, _ ...string) []slack.Attachment {
	return []slack.Attachment{{Color: "good", Text: b.name}}
}
