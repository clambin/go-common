package slackbot

import (
	"log/slog"
)

// Option is the function signature for any options for New().
type Option func(*SlackBot)

// WithName sets the name of the SlackBot. The name is currently only used in the version command.
func WithName(name string) Option {
	return func(b *SlackBot) {
		b.name = name
	}
}

// WithCommands adds a set of provided commands to the SlackBot. For more complex command structures, use Command.Add)
// and Command.AddCommand() after creating the SlackBot with New().
func WithCommands(commands map[string]Handler) Option {
	return func(b *SlackBot) {
		for name, command := range commands {
			b.Commands.Add(name, command)
		}
	}
}

// WithLogger sets the slog Logger.  By default, SlackBot uses slog.Default().
func WithLogger(logger *slog.Logger) Option {
	return func(b *SlackBot) {
		b.logger = logger
	}
}
