package seo

import (
	"fmt"
	"testing"
)

func TestTransliterate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Баня на Липовой", "banya na lipovoy"},
		{"Русская баня", "russkaya banya"},
		{"Ёлки-Палки", "yolki-palki"},
		{"Щука и Лещ", "shchuka i leshch"},
		{"Объединение", "obedinenie"},
		{"Hello World", "Hello World"},
		{"Микс Mix 123", "miks Mix 123"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Transliterate(tt.input)
			if result != tt.expected {
				t.Errorf("Transliterate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Баня на Липовой", "banya-na-lipovoy"},
		{"Русская баня №1", "russkaya-banya-1"},
		{"  Пробелы  вокруг  ", "probely-vokrug"},
		{"Hello World", "hello-world"},
		{"Баня---test", "banya-test"},
		{"Спец!символы@здесь", "spets-simvoly-zdes"},
		{"123 числа first", "123-chisla-first"},
		{"", ""},
		{"!!!???", ""},
		{"Ёжик в тумане", "yozhik-v-tumane"},
		{"Баня.на" + ".Липовой", "banya-na-lipovoy"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateUniqueSlug(t *testing.T) {
	t.Run("no collision", func(t *testing.T) {
		slug, err := GenerateUniqueSlug("Баня на Липовой", func(s string) (bool, error) {
			return false, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slug != "banya-na-lipovoy" {
			t.Errorf("got %q, want %q", slug, "banya-na-lipovoy")
		}
	})

	t.Run("first collision adds -2", func(t *testing.T) {
		taken := map[string]bool{"banya-na-lipovoy": true}
		slug, err := GenerateUniqueSlug("Баня на Липовой", func(s string) (bool, error) {
			return taken[s], nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slug != "banya-na-lipovoy-2" {
			t.Errorf("got %q, want %q", slug, "banya-na-lipovoy-2")
		}
	})

	t.Run("multiple collisions", func(t *testing.T) {
		taken := map[string]bool{
			"banya-na-lipovoy":   true,
			"banya-na-lipovoy-2": true,
			"banya-na-lipovoy-3": true,
		}
		slug, err := GenerateUniqueSlug("Баня на Липовой", func(s string) (bool, error) {
			return taken[s], nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slug != "banya-na-lipovoy-4" {
			t.Errorf("got %q, want %q", slug, "banya-na-lipovoy-4")
		}
	})

	t.Run("empty name uses fallback", func(t *testing.T) {
		slug, err := GenerateUniqueSlug("!!!", func(s string) (bool, error) {
			return false, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slug != "bathhouse" {
			t.Errorf("got %q, want %q", slug, "bathhouse")
		}
	})

	t.Run("exists error propagated", func(t *testing.T) {
		_, err := GenerateUniqueSlug("Test", func(s string) (bool, error) {
			return false, fmt.Errorf("db error")
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"banya-na-lipovoy", true},
		{"hello-world-123", true},
		{"simple", true},
		{"", false},
		{"has space", false},
		{"has@special", false},
		{"has.dot", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsValidSlug(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidSlug(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
