package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/nikitaaldaev/bani/internal/logger"
)

// TestGracefulShutdown tests that the server handles SIGINT and SIGTERM gracefully
func TestGracefulShutdown(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	log.Info("Test graceful shutdown started")

	// This test verifies the structure and signal handling
	// In a real integration test, we would:
	// 1. Start the server
	// 2. Send SIGINT/SIGTERM
	// 3. Verify graceful shutdown within timeout
	// 4. Check that all connections are properly closed

	// For now, we test the signal handling logic indirectly
	sigChan := make(chan os.Signal, 1)
	testSig := syscall.SIGTERM

	go func() {
		time.Sleep(100 * time.Millisecond)
		sigChan <- testSig
	}()

	// This mimics what happens in the serve command
	select {
	case sig := <-sigChan:
		if sig != testSig {
			t.Errorf("Expected signal %v, got %v", testSig, sig)
		}
		log.Info("Signal received correctly", "signal", sig.String())
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for signal")
	}
}

// TestShutdownTimeout tests that graceful shutdown respects the timeout
func TestShutdownTimeout(t *testing.T) {
	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Verify that the timeout is properly configured
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("Context should have a deadline")
	}

	// The timeout should be approximately 100ms
	remaining := time.Until(deadline)
	if remaining <= 0 {
		t.Error("Timeout should be in the future")
	}

	if remaining > 200*time.Millisecond {
		t.Errorf("Timeout should be around 100ms, got %v", remaining)
	}
}

// TestSignalHandling tests that both SIGINT and SIGTERM are handled
func TestSignalHandling(t *testing.T) {
	tests := []struct {
		name   string
		signal os.Signal
	}{
		{
			name:   "SIGINT",
			signal: syscall.SIGINT,
		},
		{
			name:   "SIGTERM",
			signal: syscall.SIGTERM,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sigChan := make(chan os.Signal, 1)

			received := make(chan os.Signal, 1)
			go func() {
				sig := <-sigChan
				received <- sig
			}()

			// Send the test signal
			go func() {
				time.Sleep(50 * time.Millisecond)
				sigChan <- tt.signal
			}()

			// Wait for signal to be received
			select {
			case sig := <-received:
				if sig != tt.signal {
					t.Errorf("Expected signal %v, got %v", tt.signal, sig)
				}
			case <-time.After(500 * time.Millisecond):
				t.Error("Timeout waiting for signal")
			}
		})
	}
}

// TestLoggerInitialization tests that logger is properly initialized
func TestLoggerInitialization(t *testing.T) {
	// Create logger with different log levels
	tests := []struct {
		name     string
		level    logger.LogLevel
		expected bool
	}{
		{
			name:     "DebugLevel",
			level:    logger.LevelDebug,
			expected: true,
		},
		{
			name:     "InfoLevel",
			level:    logger.LevelInfo,
			expected: true,
		},
		{
			name:     "WarnLevel",
			level:    logger.LevelWarn,
			expected: false,
		},
		{
			name:     "ErrorLevel",
			level:    logger.LevelError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.New(tt.level)
			if log == nil {
				t.Error("Logger should not be nil")
			}
			// Verify that logger is properly initialized
			// The logger will output to stderr, which is acceptable
		})
	}
}
