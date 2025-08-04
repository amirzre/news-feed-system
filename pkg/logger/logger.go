package logger

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
)

type Logger struct {
	*slog.Logger
}

// New creates a new logger instance
func New(cfg *config.Config) *Logger {
	var handler slog.Handler

	level := parseLogLevel(cfg.App.LogLevel)

	if cfg.IsDevelopment() {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{
						Key:   a.Key,
						Value: slog.StringValue(a.Value.Time().Format(time.TimeOnly)),
					}
				}
				return a
			},
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	logger := slog.New(handler)

	return &Logger{
		Logger: logger,
	}
}

// parseLogLevel converts string log level to slog level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
