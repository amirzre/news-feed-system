package logger

import (
	"fmt"
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

// HTTP request logging helpers
func (l *Logger) LogHTTPRequest(method, path string, statusCode int, duration int64) {
	l.Info(
		"HTTP request",
		"method", method,
		"path", path,
		"status", statusCode,
		"durationMS", duration,
	)
}

// Database operation logging helpers
func (l *Logger) LogDBOperation(operation, table string, duration int64, err error) {
	if err != nil {
		l.Error(
			"Database operation failed",
			"operation", operation,
			"table", table,
			"durationMs", duration,
			"error", err.Error(),
		)
	} else {
		l.Debug(
			"Database operation completed",
			"operation", operation,
			"table", table,
			"durationMS", duration,
		)
	}
}

// Service operation logging helpers
func (l *Logger) LogServiceOperation(service, operation string, success bool, duration int64) {
	fields := []any{
		"service", service,
		"operation", operation,
		"success", success,
		"durationMS", duration,
	}

	if success {
		l.Info("Service operation completed", fields...)
	} else {
		l.Warn("Service operation failed", fields...)
	}
}

// Cache operation logging helpers
func (l *Logger) LogCacheOperation(operation, key string, hit bool) {
	l.Debug(
		"Cache operation",
		"operation", operation,
		"key", key,
		"hit", hit,
	)
}

// Startup logging helper
func (l *Logger) LogStartup(service string, version string, port int) {
	l.Info(fmt.Sprintf("%s started successfully", service),
		"version", version,
		"port", port,
	)
}

// Shutdown logging helper
func (l *Logger) LogShutdown(service string, reason string) {
	l.Info(fmt.Sprintf("%s shutting down", service),
		"reason", reason,
	)
}
