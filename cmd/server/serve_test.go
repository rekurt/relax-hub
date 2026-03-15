package main

import (
	"testing"

	"github.com/nikitaaldaev/bani/internal/logger"
)

// TestWithAdminFlag tests that --with-admin flag is registered on serve command
func TestWithAdminFlag(t *testing.T) {
	f := serveCmd.Flags().Lookup("with-admin")
	if f == nil {
		t.Fatal("expected --with-admin flag to be registered on serve command")
	}
	if f.DefValue != "false" {
		t.Errorf("expected default value false, got %s", f.DefValue)
	}
}

// TestWithAdminFlagSetsVar tests that --with-admin flag sets the withAdmin variable
func TestWithAdminFlagSetsVar(t *testing.T) {
	// Reset
	withAdmin = false
	if err := serveCmd.Flags().Set("with-admin", "true"); err != nil {
		t.Fatalf("failed to set --with-admin flag: %v", err)
	}
	if !withAdmin {
		t.Error("expected withAdmin to be true after setting flag")
	}
	// Cleanup
	withAdmin = false
	_ = serveCmd.Flags().Set("with-admin", "false")
}

// TestLoggerInitialization tests that logger can be created at each log level
func TestLoggerInitialization(t *testing.T) {
	levels := []struct {
		name  string
		level logger.LogLevel
	}{
		{"DebugLevel", logger.LevelDebug},
		{"InfoLevel", logger.LevelInfo},
		{"WarnLevel", logger.LevelWarn},
		{"ErrorLevel", logger.LevelError},
	}

	for _, tt := range levels {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.New(tt.level)
			if log == nil {
				t.Error("Logger should not be nil")
			}
		})
	}
}
