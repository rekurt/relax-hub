package moderation

import (
	"testing"
)

func TestContentFilter_CheckLength(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectClean bool
		expectError string
	}{
		{
			name:        "valid text",
			text:        "Отличный сервис, все понравилось!",
			expectClean: true,
			expectError: "",
		},
		{
			name:        "text too short",
			text:        "Хорошо",
			expectClean: false,
			expectError: "text_too_short",
		},
		{
			name:        "text too long",
			text:        "a" + string(make([]byte, 5001)),
			expectClean: false,
			expectError: "text_too_long",
		},
		{
			name:        "exactly min length",
			text:        "Это очень хороший",
			expectClean: true,
			expectError: "",
		},
		{
			name:        "with spaces trimmed",
			text:        "   Хороший сервис   ",
			expectClean: true,
			expectError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(true, false)
			result := cf.CheckText(tt.text)
			if result.IsClean != tt.expectClean {
				t.Errorf("expected IsClean=%v, got %v", tt.expectClean, result.IsClean)
			}
			if tt.expectError != "" && (len(result.Reasons) == 0 || result.Reasons[0] != tt.expectError) {
				t.Errorf("expected error %s, got %v", tt.expectError, result.Reasons)
			}
		})
	}
}

func TestContentFilter_Profanity(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectClean bool
	}{
		{
			name:        "clean text",
			text:        "Отличное место! Спасибо за обслуживание.",
			expectClean: true,
		},
		{
			name:        "contains profanity",
			text:        "Это просто пиздец, не ходите туда!",
			expectClean: false,
		},
		{
			name:        "contains transliterated profanity",
			text:        "Что за pizda этот сервис!",
			expectClean: false,
		},
		{
			name:        "contains replaced symbols",
			text:        "Это п1здец просто",
			expectClean: false,
		},
		{
			name:        "contains another profanity",
			text:        "Гавно полное, не советую никому",
			expectClean: false,
		},
		{
			name:        "case insensitive profanity",
			text:        "ПИЗДЕЦ как плохо тут",
			expectClean: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(true, false)
			result := cf.CheckText(tt.text)
			if result.IsClean != tt.expectClean {
				t.Errorf("expected IsClean=%v, got %v", tt.expectClean, result.IsClean)
			}
		})
	}
}

func TestContentFilter_ExcessiveRepetition(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectClean bool
	}{
		{
			name:        "clean text",
			text:        "Спасибо за хороший сервис",
			expectClean: true,
		},
		{
			name:        "excessive letter repetition",
			text:        "Оооооочень плохо было",
			expectClean: false,
		},
		{
			name:        "excessive digit repetition",
			text:        "Цена 99999 рублей это дорого",
			expectClean: false,
		},
		{
			name:        "exclamation marks ok",
			text:        "Отлично!!! Спасибо!!!",
			expectClean: true,
		},
		{
			name:        "exactly 3 repetitions ok",
			text:        "Ааа как хорошо",
			expectClean: true,
		},
		{
			name:        "4 repetitions bad",
			text:        "Aaaaa это спам",
			expectClean: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(true, false)
			result := cf.CheckText(tt.text)
			if result.IsClean != tt.expectClean {
				t.Errorf("expected IsClean=%v, got %v", tt.expectClean, result.IsClean)
			}
		})
	}
}

func TestContentFilter_ExcessiveCapsLock(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectClean bool
	}{
		{
			name:        "normal case",
			text:        "Отличное место для отдыха",
			expectClean: true,
		},
		{
			name:        "some caps",
			text:        "ОЧЕНЬ хорошее место",
			expectClean: true,
		},
		{
			name:        "excessive caps",
			text:        "ЭТО ПРОСТО УЖАСНОЕ МЕСТО И БОЛЬШЕ НЕ ХОДИТЕ",
			expectClean: false,
		},
		{
			name:        "all caps short",
			text:        "ОК и хорошо",
			expectClean: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(true, false)
			result := cf.CheckText(tt.text)
			if result.IsClean != tt.expectClean {
				t.Errorf("expected IsClean=%v, got %v", tt.expectClean, result.IsClean)
			}
		})
	}
}

func TestContentFilter_Links(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectClean bool
	}{
		{
			name:        "clean text",
			text:        "Отличный сервис, спасибо",
			expectClean: true,
		},
		{
			name:        "contains http link",
			text:        "Посетите http://example.com для информации",
			expectClean: false,
		},
		{
			name:        "contains https link",
			text:        "Смотрите https://example.com",
			expectClean: false,
		},
		{
			name:        "contains www",
			text:        "Посетите www.example.com",
			expectClean: false,
		},
		{
			name:        "contains .ru domain",
			text:        "Моя страница на example.ru",
			expectClean: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(true, false)
			result := cf.CheckText(tt.text)
			if result.IsClean != tt.expectClean {
				t.Errorf("expected IsClean=%v, got %v", tt.expectClean, result.IsClean)
			}
		})
	}
}

func TestContentFilter_PhoneNumbers(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectClean bool
	}{
		{
			name:        "clean text",
			text:        "Отличный сервис",
			expectClean: true,
		},
		{
			name:        "contains phone with +7",
			text:        "Позвоните на +7 999 123-45-67",
			expectClean: false,
		},
		{
			name:        "contains phone with 8",
			text:        "Номер 8-999-123-45-67",
			expectClean: false,
		},
		{
			name:        "fake phone format",
			text:        "Мой номер +7(999)123-45-67",
			expectClean: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(true, false)
			result := cf.CheckText(tt.text)
			if result.IsClean != tt.expectClean {
				t.Errorf("expected IsClean=%v, got %v", tt.expectClean, result.IsClean)
			}
		})
	}
}

func TestContentFilter_MultipleViolations(t *testing.T) {
	cf := NewContentFilter(true, false)
	text := "ПИЗДЕЦ!!!! Позвоните https://example.com"
	result := cf.CheckText(text)

	if result.IsClean {
		t.Error("expected multiple violations to be detected")
	}

	// Should have multiple reasons
	if len(result.Reasons) == 0 {
		t.Error("expected at least one reason")
	}
}

func TestContentFilter_IsEnabled(t *testing.T) {
	cf1 := NewContentFilter(true, false)
	if !cf1.IsEnabled() {
		t.Error("expected filter to be enabled")
	}

	cf2 := NewContentFilter(false, true)
	if cf2.IsEnabled() {
		t.Error("expected filter to be disabled")
	}
}

func TestContentFilter_ShouldAutoApprove(t *testing.T) {
	cf1 := NewContentFilter(true, true)
	if !cf1.ShouldAutoApprove() {
		t.Error("expected auto-approve to be enabled")
	}

	cf2 := NewContentFilter(true, false)
	if cf2.ShouldAutoApprove() {
		t.Error("expected auto-approve to be disabled")
	}
}

func TestContentFilter_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectClean bool
	}{
		{
			name:        "only numbers",
			text:        "1234567890123",
			expectClean: true,
		},
		{
			name:        "emoji in text",
			text:        "Отлично! 😊 Спасибо за обслуживание",
			expectClean: true,
		},
		{
			name:        "unicode characters",
			text:        "Привет мир! Это хороший сервис.",
			expectClean: true,
		},
		{
			name:        "mixed case profanity bypass attempt",
			text:        "пИзДеЦ этот сервис",
			expectClean: false,
		},
		{
			name:        "profanity with spaces",
			text:        "п и з д е ц плохо",
			expectClean: true, // Currently won't match with spaces
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(true, false)
			result := cf.CheckText(tt.text)
			if result.IsClean != tt.expectClean {
				t.Errorf("expected IsClean=%v, got %v", tt.expectClean, result.IsClean)
			}
		})
	}
}

func TestContentFilter_BypassAttempts(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		shouldCatch bool
	}{
		{
			name:        "obvious profanity",
			text:        "Гавно полное",
			shouldCatch: true,
		},
		{
			name:        "transliterated profanity",
			text:        "Gavno polnoe",
			shouldCatch: true,
		},
		{
			name:        "replaced chars like @ for a",
			text:        "Г@вно это ужас",
			shouldCatch: true,
		},
		{
			name:        "numbers replace",
			text:        "Г4вн0 это плохо",
			shouldCatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(true, false)
			result := cf.CheckText(tt.text)
			if tt.shouldCatch && result.IsClean {
				t.Errorf("expected to catch violation in: %s", tt.text)
			}
		})
	}
}
