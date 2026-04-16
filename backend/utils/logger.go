package utils

import (
	"log/slog"
	"os"
	"sync"
)

var (
	Logger     *slog.Logger
	loggerOnce sync.Once
	loggerMu   sync.RWMutex
)

func InitLogger(level string) {
	loggerOnce.Do(func() {
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
	})
}

func getLogger() *slog.Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	if Logger == nil {
		InitLogger("info")
	}
	return Logger
}

func Info(msg string, args ...any) {
	getLogger().Info(msg, args...)
}

func Debug(msg string, args ...any) {
	getLogger().Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	getLogger().Warn(msg, args...)
}

func Error(msg string, err error, args ...any) {
	if err != nil {
		args = append(args, "error", err)
	}
	getLogger().Error(msg, args...)
}

func Fatal(msg string, err error, args ...any) {
	Error(msg, err, args...)
	os.Exit(1)
}

func WithError(err error) *slog.Logger {
	return getLogger().With("error", err)
}
