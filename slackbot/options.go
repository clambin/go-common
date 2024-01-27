package slackbot

import (
	"log/slog"
)

// Option is the function signature for any options for New().
type Option func(*SlackBot)

// WithName sets the name of the SlackBot. The name is currently only used in the version Command.
func WithName(name string) Option {
	return func(b *SlackBot) {
		b.name = name
	}
}

// WithCommands adds a set of provided commands to the SlackBot. For more complex Command structures, use AddCommand & AddCommandGroup.
// and Command.AddCommandGroup() after creating the SlackBot with New().
func WithCommands(commands Commands) Option {
	return func(b *SlackBot) {
		b.Add(commands)
	}
}

// WithLogger sets the slog Logger.  By default, SlackBot uses slog.Default().
func WithLogger(logger *slog.Logger) Option {
	return func(b *SlackBot) {
		b.logger = logger
	}
}
