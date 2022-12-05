package slackbot

import (
	"context"
	"github.com/clambin/go-common/slackbot/client"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"regexp"
	"strings"
)

// SlackBot connects to Slack through Slack's Bot integration.
type SlackBot struct {
	client.SlackClient
	name     string
	channels []string
	commands *commands
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

	b.commands.Register("help", b.doHelp)
	b.commands.Register("version", b.doVersion)

	for cmd, callbackFunction := range commands {
		b.commands.Register(cmd, callbackFunction)
	}

	return &b
}

// Run the slackbot
func (b *SlackBot) Run(ctx context.Context) (err error) {
	if b.channels, err = b.SlackClient.GetChannels(); err != nil {
		return err
	}

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
			log.WithError(err).Warning("failed to post message on Slack")
		}
	}
	return
}

func (b *SlackBot) Send(channel string, attachments []slack.Attachment) error {
	var channels = b.channels
	if channel != "" {
		channels = []string{channel}
	}

	var err error
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
	cleanInput := strings.Replace(input, "“", "\"", -1)
	// TODO: why are we doing this twice?
	cleanInput = strings.Replace(cleanInput, "”", "\"", -1)
	r := regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
	output = r.FindAllString(cleanInput, -1)

	log.WithField("parsed", output).Debug("parsed slack input")
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
