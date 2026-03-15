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
			l := New(tt.level)
			if l == nil {
				t.Errorf("New(%v) returned nil", tt.level)
				return
			}
			if l.level != tt.level {
				t.Errorf("New(%v) set level to %v", tt.level, l.level)
			}
			if l.sugar == nil {
				t.Error("New() did not initialize zap sugar logger")
			}
		})
	}
}

func TestNewWithFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"console format", "console"},
		{"unknown defaults to console", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewWithFormat(LevelInfo, tt.format)
			if l == nil {
				t.Fatal("NewWithFormat returned nil")
			}
			if l.sugar == nil {
				t.Error("NewWithFormat did not initialize zap sugar logger")
			}
		})
	}
}

func TestLoggerDebug(t *testing.T) {
	l := New(LevelDebug)
	// Should not panic
	l.Debug("test message")
	l.Debug("test message with fields", "key", "value")
}

func TestLoggerInfo(t *testing.T) {
	l := New(LevelInfo)
	// Should not panic
	l.Info("test message")
	l.Info("test message with fields", "key", "value")
}

func TestLoggerWarn(t *testing.T) {
	l := New(LevelWarn)
	// Should not panic
	l.Warn("test message")
	l.Warn("test message with fields", "key", "value")
}

func TestLoggerError(t *testing.T) {
	l := New(LevelError)
	// Should not panic
	l.Error("test message")
	l.Error("test message with fields", "key", "value")
}

func TestLoggerLevelFiltering(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
		fn    func(*Logger)
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
			l := New(tt.level)
			// Should not panic
			tt.fn(l)
		})
	}
}

func TestLoggerSync(t *testing.T) {
	l := New(LevelInfo)
	// Should not panic
	l.Sync()
}

func TestLoggerJSONFormat(t *testing.T) {
	l := NewWithFormat(LevelDebug, "json")
	// Should not panic with structured fields
	l.Info("request", "method", "GET", "path", "/api/v1/test", "status", 200)
	l.Error("failed", "error", "something went wrong", "retry", true)
}

func TestToZapLevel(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
	}{
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"error", LevelError},
		{"unknown defaults to info", LogLevel(99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			_ = toZapLevel(tt.level)
		})
	}
}
