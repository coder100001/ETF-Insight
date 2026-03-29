package utils

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitLogger(tt.level)
			if Logger == nil {
				t.Error("Logger should not be nil after InitLogger")
			}
		})
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	Logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	Info("test message", "key", "value")

	if !strings.Contains(buf.String(), "test message") {
		t.Error("Info log should contain message")
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	Logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	Error("test error", nil)

	if !strings.Contains(buf.String(), "test error") {
		t.Error("Error log should contain message")
	}
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	Logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	Warn("test warning")

	if !strings.Contains(buf.String(), "test warning") {
		t.Error("Warn log should contain message")
	}
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	Logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	Debug("test debug")

	if !strings.Contains(buf.String(), "test debug") {
		t.Error("Debug log should contain message")
	}
}

func TestWithError(t *testing.T) {
	InitLogger("info")

	logger := WithError(nil)
	if logger == nil {
		t.Error("WithError should return logger")
	}
}
