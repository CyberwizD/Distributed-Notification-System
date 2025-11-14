package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New builds a slog logger with the requested verbosity level.
func New(level string) *slog.Logger {
	lvl := parseLevel(level)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})
	return slog.New(handler)
}

func parseLevel(v string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
