package utils

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func InitLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

func Error(msg string, err error, args ...any) {
	if err != nil {
		args = append(args, "error", err)
	}
	Logger.Error(msg, args...)
}

func Fatal(msg string, err error, args ...any) {
	Error(msg, err, args...)
	os.Exit(1)
}

func WithError(err error) *slog.Logger {
	return Logger.With("error", err)
}
