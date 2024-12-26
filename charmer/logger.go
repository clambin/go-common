package charmer

import (
	"context"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

type logCtxType string

var logCtx logCtxType = "logger"

// SetTextLogger creates a slog.Logger with a slog.TextHandler and adds it to the command's context.
// If debug is true, the handler's log level is set to slog.LevelDebug.
func SetTextLogger(cmd *cobra.Command, debug bool) {
	var opts slog.HandlerOptions
	if debug {
		opts.Level = slog.LevelDebug
	}
	SetLogger(cmd, slog.New(slog.NewTextHandler(os.Stderr, &opts)))
}

// SetJSONLogger creates a slog.Logger with a slog.JSONHandler and adds it to the command's context.
// If debug is true, the handler's log level is set to slog.LevelDebug.
func SetJSONLogger(cmd *cobra.Command, debug bool) {
	var opts slog.HandlerOptions
	if debug {
		opts.Level = slog.LevelDebug
	}
	SetLogger(cmd, slog.New(slog.NewJSONHandler(os.Stderr, &opts)))
}

// SetLogger adds a slog.Logger to the command's context.
func SetLogger(cmd *cobra.Command, logger *slog.Logger) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, logCtx, logger))
}

// GetLogger returns the logger from the command's context. If no logger was set, it returns slog.Default().
func GetLogger(cmd *cobra.Command) *slog.Logger {
	if ctx := cmd.Context(); ctx != nil {
		if v := ctx.Value(logCtx); v != nil {
			if l, ok := v.(*slog.Logger); ok && l != nil {
				return l
			}
		}
	}
	return slog.Default()
}
