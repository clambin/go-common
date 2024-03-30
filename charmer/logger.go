package charmer

import (
	"context"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

type logCtxType string

var logCtx logCtxType = "logger"

func SetTextLogger(cmd *cobra.Command, debug bool) {
	var opts slog.HandlerOptions
	if debug {
		opts.Level = slog.LevelDebug
	}
	SetLogger(cmd, slog.New(slog.NewTextHandler(os.Stderr, &opts)))
}

func SetJSONLogger(cmd *cobra.Command, debug bool) {
	var opts slog.HandlerOptions
	if debug {
		opts.Level = slog.LevelDebug
	}
	SetLogger(cmd, slog.New(slog.NewJSONHandler(os.Stderr, &opts)))
}

func SetLogger(cmd *cobra.Command, logger *slog.Logger) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, logCtx, logger))
}

func GetLogger(cmd *cobra.Command) *slog.Logger {
	if ctx := cmd.Context(); ctx != nil {
		if l := ctx.Value(logCtx); l != nil {
			return l.(*slog.Logger)
		}
	}
	return slog.Default()
}
