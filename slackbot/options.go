package slackbot

import (
	"log/slog"
)

type Option func(*SlackBot)

func WithName(name string) Option {
	return func(b *SlackBot) {
		b.name = name
	}
}

func WithCommands(commands map[string]Handler) Option {
	return func(b *SlackBot) {
		for name, command := range commands {
			b.commands.Add(name, command)
		}
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(b *SlackBot) {
		b.logger = logger
	}
}
