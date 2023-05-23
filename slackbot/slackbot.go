package slackbot

import (
	"context"
	"github.com/slack-go/slack"
	"golang.org/x/exp/slog"
	"regexp"
	"strings"
	"sync"
)

// SlackBot connects to Slack through Slack's Bot integration.
type SlackBot struct {
	client   *slackClient
	name     string
	commands *commandRunner
	logger   *slog.Logger
}

// CommandFunc signature for command callback functions
type CommandFunc func(ctx context.Context, args ...string) []slack.Attachment

// New creates a new slackbot
func New(name string, slackToken string, commands map[string]CommandFunc, logger *slog.Logger) *SlackBot {
	if logger == nil {
		logger = slog.Default()
	}
	b := SlackBot{
		name:     name,
		commands: newCommandRunner(),
		client:   newSlackClient(slackToken, logger),
		logger:   logger,
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
func (b *SlackBot) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); b.client.Run(ctx) }()
	for {
		select {
		case message := <-b.client.GetMessage():
			b.processMessage(ctx, message)
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

func (b *SlackBot) processMessage(ctx context.Context, message *slack.MessageEvent) {
	b.logger.Debug("message received",
		"user.id", message.User,
		"user.name", message.Username,
		"text", message.Text,
	)

	command, args, err := b.parseCommand(message.Text)
	if err != nil {
		b.logger.Error("failed to parse command", "err", err)
		return
	}
	if command == "" {
		return
	}

	b.logger.Debug("running command", "command", command, "args", args)
	output := b.commands.Do(ctx, command, args...)

	err = b.Send(message.Channel, output)
	if err != nil {
		b.logger.Error("failed to post to Slack", "err", err)
	}
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
