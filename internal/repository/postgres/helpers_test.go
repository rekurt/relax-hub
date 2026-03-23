package postgres

import (
	"errors"
	"testing"
)

func TestBuildPrefixTsQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single word", "баня", "баня:*"},
		{"multiple words", "русская баня", "русская:* & баня:*"},
		{"with special chars", "баня!@#", "баня:*"},
		{"empty string", "", ""},
		{"only special chars", "!@#$%", ""},
		{"mixed", "hello world123", "hello:* & world123:*"},
		{"extra spaces", "  баня   сауна  ", "баня:* & сауна:*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPrefixTsQuery(tt.input)
			if result != tt.expected {
				t.Errorf("buildPrefixTsQuery(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "duplicate key error with text",
			err:      errors.New("duplicate key value violates unique constraint"),
			expected: true,
		},
		{
			name:     "postgres error code 23505",
			err:      errors.New("ERROR: 23505 unique_violation"),
			expected: true,
		},
		{
			name:     "regular error",
			err:      errors.New("connection refused"),
			expected: false,
		},
		{
			name:     "empty error",
			err:      errors.New(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDuplicateKeyError(tt.err)
			if result != tt.expected {
				t.Errorf("isDuplicateKeyError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}
