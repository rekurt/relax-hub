package notification

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rekurt/relax-hub/internal/logger"
)

func TestParseFrom(t *testing.T) {
	cases := []struct {
		name        string
		raw         string
		wantHeader  string
		wantEnv     string
		expectError bool
	}{
		{
			name:       "bare address",
			raw:        "partners@relax-hub.ru",
			wantHeader: "<partners@relax-hub.ru>",
			wantEnv:    "partners@relax-hub.ru",
		},
		{
			name:       "display name",
			raw:        "RelaxHUB <partners@relax-hub.ru>",
			wantHeader: `"RelaxHUB" <partners@relax-hub.ru>`,
			wantEnv:    "partners@relax-hub.ru",
		},
		{
			name:       "leading and trailing whitespace trimmed",
			raw:        "   partners@relax-hub.ru\t",
			wantHeader: "<partners@relax-hub.ru>",
			wantEnv:    "partners@relax-hub.ru",
		},
		{
			name:        "empty",
			raw:         "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			raw:         "   ",
			expectError: true,
		},
		{
			name:        "garbage",
			raw:         "not-an-address",
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			header, env, err := parseFrom(tc.raw)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error, got header=%q env=%q", header, env)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if header != tc.wantHeader {
				t.Errorf("header: got %q, want %q", header, tc.wantHeader)
			}
			if env != tc.wantEnv {
				t.Errorf("envelope: got %q, want %q", env, tc.wantEnv)
			}
		})
	}
}

func newQuietLogger() *logger.Logger { return logger.New(logger.LevelError) }

func TestNewSMTPEmailSender_Constructs(t *testing.T) {
	cfg := SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "RelaxHUB <noreply@relax-hub.ru>",
	}
	sender := NewSMTPEmailSender(cfg, newQuietLogger())
	if sender == nil {
		t.Fatal("NewSMTPEmailSender returned nil")
	}
}

func TestSMTPEmailSender_Send_InvalidFromReturnsError(t *testing.T) {
	sender := NewSMTPEmailSender(SMTPConfig{From: "  "}, newQuietLogger())
	err := sender.Send(context.Background(), "to@example.com", "subject", "body")
	if err == nil {
		t.Fatal("expected error for empty From")
	}
	if !strings.Contains(err.Error(), "invalid From") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSMTPEmailSender_Send_InvalidHostReturnsError(t *testing.T) {
	// Port 1 on localhost is reserved; deliver() is forced through the dial
	// error branch, exercising the failure-logging path.
	sender := NewSMTPEmailSender(SMTPConfig{
		Host: "127.0.0.1",
		Port: 1,
		From: "from@example.com",
	}, newQuietLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := sender.Send(ctx, "to@example.com", "subject", "body")
	if err == nil {
		t.Fatal("expected dial error")
	}
	if !strings.Contains(err.Error(), "send email") {
		t.Errorf("unexpected error message: %v", err)
	}
}
