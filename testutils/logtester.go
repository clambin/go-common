package testutils

import (
	"io"
	"log/slog"
)

func NewTextLogger(output io.Writer, level slog.Level) *slog.Logger {
	opts := getOptions(level)
	return slog.New(slog.NewTextHandler(output, &opts))
}

func NewJSONLogger(output io.Writer, level slog.Level) *slog.Logger {
	opts := getOptions(level)
	return slog.New(slog.NewJSONHandler(output, &opts))
}

func getOptions(level slog.Level) slog.HandlerOptions {
	return slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
}
