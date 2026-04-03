package utils

import (
	"bytes"
	"log/slog"
	"os"
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
		})
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	Info("test message", "key", "value")

	if !strings.Contains(buf.String(), "test message") {
		t.Error("Info log should contain message")
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))
	slog.SetDefault(logger)

	Error("test error", nil)

	if !strings.Contains(buf.String(), "test error") {
		t.Error("Error log should contain message")
	}
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	slog.SetDefault(logger)

	Warn("test warning")

	if !strings.Contains(buf.String(), "test warning") {
		t.Error("Warn log should contain message")
	}
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	Debug("test debug")

	if !strings.Contains(buf.String(), "test debug") {
		t.Error("Debug log should contain message")
	}
}

func TestFatal(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		Fatal("test fatal", nil)
		return
	}
}
