package logger

import (
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogLevel
	}{
		{"debug", "debug", LevelDebug},
		{"info", "info", LevelInfo},
		{"warn", "warn", LevelWarn},
		{"error", "error", LevelError},
		{"default", "unknown", LevelInfo},
		{"empty", "", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLogLevel(tt.input)
			if got != tt.expected {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLoggerNew(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
	}{
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"error", LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level)
			if logger == nil {
				t.Errorf("New(%v) returned nil", tt.level)
				return
			}
			if logger.level != tt.level {
				t.Errorf("New(%v) set level to %v", tt.level, logger.level)
			}
		})
	}
}

func TestLoggerDebug(t *testing.T) {
	logger := New(LevelDebug)
	// Should not panic
	logger.Debug("test message")
	logger.Debug("test message with fields", "key", "value")
}

func TestLoggerInfo(t *testing.T) {
	logger := New(LevelInfo)
	// Should not panic
	logger.Info("test message")
	logger.Info("test message with fields", "key", "value")
}

func TestLoggerWarn(t *testing.T) {
	logger := New(LevelWarn)
	// Should not panic
	logger.Warn("test message")
	logger.Warn("test message with fields", "key", "value")
}

func TestLoggerError(t *testing.T) {
	logger := New(LevelError)
	// Should not panic
	logger.Error("test message")
	logger.Error("test message with fields", "key", "value")
}

func TestLoggerLevelFiltering(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
		// We can't easily test output, but we can test that different levels don't panic
		fn func(*Logger)
	}{
		{
			name:  "error level logs errors",
			level: LevelError,
			fn:    func(l *Logger) { l.Error("msg") },
		},
		{
			name:  "warn level logs warn and error",
			level: LevelWarn,
			fn:    func(l *Logger) { l.Warn("msg"); l.Error("msg") },
		},
		{
			name:  "info level logs info and above",
			level: LevelInfo,
			fn:    func(l *Logger) { l.Info("msg"); l.Warn("msg"); l.Error("msg") },
		},
		{
			name:  "debug level logs everything",
			level: LevelDebug,
			fn:    func(l *Logger) { l.Debug("msg"); l.Info("msg"); l.Warn("msg"); l.Error("msg") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level)
			// Should not panic
			tt.fn(logger)
		})
	}
}
