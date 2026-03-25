package moderation

import (
	"context"
	"math"
	"testing"
)

func TestRegexTextModerator_CleanText(t *testing.T) {
	m := NewRegexTextModerator()
	result := m.Analyze(context.Background(), "Отличное место для отдыха! Спасибо за обслуживание.")

	if result.Flagged {
		t.Error("expected clean text not to be flagged")
	}
	if result.Score != 0 {
		t.Errorf("expected score 0, got %f", result.Score)
	}
	if len(result.Flags) != 0 {
		t.Errorf("expected no flags, got %v", result.Flags)
	}
}

func TestRegexTextModerator_Profanity(t *testing.T) {
	m := NewRegexTextModerator()
	result := m.Analyze(context.Background(), "Это просто пиздец, не ходите туда никогда!")

	if !result.Flagged {
		t.Error("expected profanity to be flagged (score >= 0.7)")
	}
	if result.Score < scoreProfanity {
		t.Errorf("expected score >= %f, got %f", scoreProfanity, result.Score)
	}
	assertContainsFlag(t, result.Flags, "contains_profanity")
}

func TestRegexTextModerator_LinkDetection(t *testing.T) {
	m := NewRegexTextModerator()
	result := m.Analyze(context.Background(), "Лучше посмотрите https://example.com для информации")

	assertContainsFlag(t, result.Flags, "contains_links")
	if result.Score < scoreContainsLink {
		t.Errorf("expected score >= %f, got %f", scoreContainsLink, result.Score)
	}
}

func TestRegexTextModerator_PhoneDetection(t *testing.T) {
	m := NewRegexTextModerator()
	result := m.Analyze(context.Background(), "Позвоните по номеру +7 999 123-45-67 для записи")

	assertContainsFlag(t, result.Flags, "contains_phone_numbers")
}

func TestRegexTextModerator_EmailDetection(t *testing.T) {
	m := NewRegexTextModerator()
	result := m.Analyze(context.Background(), "Пишите на email test@example.com для связи")

	assertContainsFlag(t, result.Flags, "contains_email")
	if result.Score < scoreContainsEmail {
		t.Errorf("expected score >= %f, got %f", scoreContainsEmail, result.Score)
	}
}

func TestRegexTextModerator_ExcessiveRepetition(t *testing.T) {
	m := NewRegexTextModerator()
	result := m.Analyze(context.Background(), "Оооооочень плохо было тут")

	assertContainsFlag(t, result.Flags, "excessive_repetition")
}

func TestRegexTextModerator_ExcessiveCaps(t *testing.T) {
	m := NewRegexTextModerator()
	result := m.Analyze(context.Background(), "ЭТО ПРОСТО УЖАСНОЕ МЕСТО И БОЛЬШЕ НЕ ХОДИТЕ")

	assertContainsFlag(t, result.Flags, "excessive_caps_lock")
}

func TestRegexTextModerator_MultipleViolations(t *testing.T) {
	m := NewRegexTextModerator()
	// profanity (0.5) + link (0.3) + caps (0.15) = 0.95 -> capped at 0.95
	result := m.Analyze(context.Background(), "ПИЗДЕЦ!!! СМОТРИТЕ https://example.com")

	if !result.Flagged {
		t.Error("expected multiple violations to be flagged")
	}
	if result.Score <= 0.7 {
		t.Errorf("expected score > 0.7, got %f", result.Score)
	}
	if result.Score > 1.0 {
		t.Errorf("expected score capped at 1.0, got %f", result.Score)
	}
}

func TestRegexTextModerator_ScoreCappedAtOne(t *testing.T) {
	m := NewRegexTextModerator()
	// profanity (0.5) + link (0.3) + phone (0.3) + email (0.3) = 1.4 -> capped at 1.0
	result := m.Analyze(context.Background(), "ПИЗДЕЦ https://example.com +7 999 123-45-67 test@mail.ru")

	if result.Score > 1.0 {
		t.Errorf("expected score capped at 1.0, got %f", result.Score)
	}
	if math.Abs(result.Score-1.0) > 0.001 {
		t.Errorf("expected score 1.0, got %f", result.Score)
	}
}

func TestRegexTextModerator_BelowThreshold(t *testing.T) {
	m := NewRegexTextModerator()
	// Only caps lock (0.15) - below 0.7 threshold
	result := m.Analyze(context.Background(), "ЭТО ОЧЕНЬ ПЛОХО НА МОЙ ВЗГЛЯД")

	if result.Flagged {
		t.Error("expected below-threshold score not to be flagged")
	}
	assertContainsFlag(t, result.Flags, "excessive_caps_lock")
}

func TestRegexTextModerator_ShortText(t *testing.T) {
	m := NewRegexTextModerator()
	result := m.Analyze(context.Background(), "Плохо")

	assertContainsFlag(t, result.Flags, "text_too_short")
	if result.Score < scoreTextTooShort {
		t.Errorf("expected score >= %f, got %f", scoreTextTooShort, result.Score)
	}
}

func TestRegexTextModerator_ExclamationsOK(t *testing.T) {
	m := NewRegexTextModerator()
	result := m.Analyze(context.Background(), "Отлично!!! Спасибо!!! Замечательно!!!")

	// Exclamation marks are punctuation and should not trigger repetition
	for _, f := range result.Flags {
		if f == "excessive_repetition" {
			t.Error("exclamation marks should not trigger excessive_repetition")
		}
	}
}

func TestRegexTextModerator_CapsThreshold50Percent(t *testing.T) {
	m := NewRegexTextModerator()
	// Exactly at 50% - "ОЧЕНЬ хорошее место" has mixed case
	result := m.Analyze(context.Background(), "ОЧЕНЬ хорошее место для отдыха")

	for _, f := range result.Flags {
		if f == "excessive_caps_lock" {
			t.Error("50% or lower caps should not be flagged")
		}
	}
}

func assertContainsFlag(t *testing.T, flags []string, expected string) {
	t.Helper()
	for _, f := range flags {
		if f == expected {
			return
		}
	}
	t.Errorf("expected flag %q in %v", expected, flags)
}
