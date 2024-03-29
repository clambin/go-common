package slackbot

import (
	"context"
	"fmt"
	slackclient "github.com/clambin/go-common/slackbot/internal/slack-client"
	"github.com/slack-go/slack"
	"log/slog"
	"regexp"
	"strings"
	"sync"
)

// SlackBot connects to Slack through Slack's Bot integration.
type SlackBot struct {
	client SlackClient
	Commands
	name   string
	logger *slog.Logger
}

type SlackClient interface {
	Run(context.Context)
	GetMessage() chan *slack.MessageEvent
	Send(channelID string, attachments []slack.Attachment) error
	GetUserID() (string, error)
	GetChannels() ([]string, error)
}

// New creates a new slackbot
func New(slackToken string, options ...Option) *SlackBot {
	b := &SlackBot{
		name:   "slackbot",
		logger: slog.Default(),
	}
	b.Commands = Commands{
		"help":    HandlerFunc(b.doHelp),
		"version": HandlerFunc(b.doVersion),
	}

	for _, option := range options {
		option(b)
	}

	b.client = slackclient.New(slackToken, b.logger.With("component", "slack-client"))

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

// Send posts attachments to Slack on the provided channel. If channel is blank, Send posts to  all channels that the bot has access to.
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

func (b *SlackBot) processMessage(ctx context.Context, message *slack.MessageEvent) error {
	if message.Text == "" {
		return nil
	}
	b.logger.Debug("message received",
		"user.id", message.User,
		"user.name", message.Username,
		"text", message.Text,
	)

	args, err := b.parseCommand(message.Text)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	if len(args) == 0 {
		return nil
	}

	b.logger.Debug("running command", "args", args)
	output := b.Commands.Handle(ctx, args...)

	return b.Send(message.Channel, output)
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

// TODO: this only works for top-level commands. Should this be part of Command functionality?
func (b *SlackBot) doHelp(_ context.Context, _ ...string) []slack.Attachment {
	return []slack.Attachment{{
		Color: "good",
		Title: "supported commands",
		Text:  strings.Join(b.GetCommands(), ", "),
	}}
}

func (b *SlackBot) doVersion(_ context.Context, _ ...string) []slack.Attachment {
	return []slack.Attachment{{Color: "good", Text: b.name}}
}
