package moderation

import (
	"context"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ModerationResult holds the result of text analysis.
type ModerationResult struct {
	Flagged bool     `json:"flagged"`
	Score   float64  `json:"score"`
	Flags   []string `json:"flags"`
}

// TextModerationService analyzes text for policy violations.
type TextModerationService interface {
	Analyze(ctx context.Context, text string) ModerationResult
}

// Score weights for different violation types.
const (
	scoreProfanity          = 0.8
	scoreExcessiveRepeat    = 0.2
	scoreExcessiveCaps      = 0.15
	scoreContainsLink       = 0.3
	scoreContainsPhone      = 0.3
	scoreContainsEmail      = 0.3
	scoreTextTooShort       = 0.1
	scoreTextTooLong        = 0.1
	moderationFlagThreshold = 0.7
)

var emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

// RegexTextModerator is a regex-based implementation of TextModerationService.
type RegexTextModerator struct{}

// NewRegexTextModerator creates a new regex-based text moderator.
func NewRegexTextModerator() *RegexTextModerator {
	return &RegexTextModerator{}
}

// Analyze checks text for policy violations and returns a scored result.
func (m *RegexTextModerator) Analyze(_ context.Context, text string) ModerationResult {
	result := ModerationResult{
		Score: 0,
		Flags: []string{},
	}

	trimmed := strings.TrimSpace(text)
	length := utf8.RuneCountInString(trimmed)

	if length < 10 {
		result.Score += scoreTextTooShort
		result.Flags = append(result.Flags, "text_too_short")
	}
	if length > 5000 {
		result.Score += scoreTextTooLong
		result.Flags = append(result.Flags, "text_too_long")
	}

	if hasProfanity(text) {
		result.Score += scoreProfanity
		result.Flags = append(result.Flags, "contains_profanity")
	}

	if hasRepetition(text) {
		result.Score += scoreExcessiveRepeat
		result.Flags = append(result.Flags, "excessive_repetition")
	}

	if hasCaps(text) {
		result.Score += scoreExcessiveCaps
		result.Flags = append(result.Flags, "excessive_caps_lock")
	}

	if urlPattern.MatchString(text) {
		result.Score += scoreContainsLink
		result.Flags = append(result.Flags, "contains_links")
	}

	if phonePattern.MatchString(text) {
		result.Score += scoreContainsPhone
		result.Flags = append(result.Flags, "contains_phone_numbers")
	}

	if emailPattern.MatchString(text) {
		result.Score += scoreContainsEmail
		result.Flags = append(result.Flags, "contains_email")
	}

	// Cap score at 1.0
	if result.Score > 1.0 {
		result.Score = 1.0
	}

	result.Flagged = result.Score >= moderationFlagThreshold
	return result
}

// hasProfanity checks for profanity in the text.
func hasProfanity(text string) bool {
	textLower := strings.ToLower(text)
	for _, word := range Stopwords {
		for _, variation := range word.Variations() {
			if strings.Contains(textLower, strings.ToLower(variation)) {
				return true
			}
		}
	}
	return false
}

// hasRepetition checks for 4+ consecutive identical non-punctuation characters.
func hasRepetition(text string) bool {
	runes := []rune(text)
	for i := 0; i < len(runes)-3; i++ {
		if runes[i] == runes[i+1] && runes[i] == runes[i+2] && runes[i] == runes[i+3] {
			if !isPunctuation(runes[i]) {
				return true
			}
		}
	}
	return false
}

// hasCaps checks if more than 50% of letters are uppercase.
func hasCaps(text string) bool {
	letterCount := 0
	upperCount := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			letterCount++
			if unicode.IsUpper(r) {
				upperCount++
			}
		}
	}
	if letterCount == 0 {
		return false
	}
	return (float64(upperCount) / float64(letterCount)) > 0.5
}
