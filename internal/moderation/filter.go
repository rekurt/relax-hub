package moderation

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	urlPattern   = regexp.MustCompile(`(?i)(https?://|www\.|\.ru|\.com|\.org)`)
	phonePattern = regexp.MustCompile(`(?i)(\+7|8-?)\s*\(?[0-9]{3}\)?[\s-]?[0-9]{3}[\s-]?[0-9]{2}[\s-]?[0-9]{2}`)
)

// FilterResult contains the result of content filtering
type FilterResult struct {
	IsClean bool
	Reasons []string
}

// ContentFilter performs moderation checks on review text
type ContentFilter struct {
	enabled     bool
	autoApprove bool
}

// NewContentFilter creates a new content filter
func NewContentFilter(enabled, autoApprove bool) *ContentFilter {
	return &ContentFilter{
		enabled:     enabled,
		autoApprove: autoApprove,
	}
}

// CheckText validates the review text and returns whether it's clean and any violation reasons
func (cf *ContentFilter) CheckText(text string) FilterResult {
	result := FilterResult{
		IsClean: true,
		Reasons: []string{},
	}

	// Check length
	if err := cf.checkLength(text); err != "" {
		result.IsClean = false
		result.Reasons = append(result.Reasons, err)
	}

	// Check for profanity
	if violations := cf.checkProfanity(text); len(violations) > 0 {
		result.IsClean = false
		result.Reasons = append(result.Reasons, violations...)
	}

	// Check for spam patterns
	if violations := cf.checkSpamPatterns(text); len(violations) > 0 {
		result.IsClean = false
		result.Reasons = append(result.Reasons, violations...)
	}

	return result
}

// checkLength validates text length (10-5000 characters)
func (cf *ContentFilter) checkLength(text string) string {
	text = strings.TrimSpace(text)
	length := utf8.RuneCountInString(text)

	if length < 10 {
		return "text_too_short"
	}
	if length > 5000 {
		return "text_too_long"
	}
	return ""
}

// checkProfanity checks for profanity and variations
func (cf *ContentFilter) checkProfanity(text string) []string {
	var violations []string
	textLower := strings.ToLower(text)
	foundProfanity := false

	for _, word := range Stopwords {
		for _, variation := range word.Variations() {
			if strings.Contains(textLower, strings.ToLower(variation)) {
				foundProfanity = true
				break
			}
		}
		if foundProfanity {
			break
		}
	}

	if foundProfanity {
		violations = append(violations, "contains_profanity")
	}

	return violations
}

// checkSpamPatterns checks for common spam patterns
func (cf *ContentFilter) checkSpamPatterns(text string) []string {
	var violations []string

	// Check for excessive repetition (e.g., "aaaaaaa")
	if cf.hasExcessiveRepetition(text) {
		violations = append(violations, "excessive_repetition")
	}

	// Check for excessive caps lock
	if cf.hasExcessiveCapsLock(text) {
		violations = append(violations, "excessive_caps_lock")
	}

	// Check for links
	if cf.containsLinks(text) {
		violations = append(violations, "contains_links")
	}

	// Check for phone numbers (competitor phone numbers)
	if cf.containsPhoneNumbers(text) {
		violations = append(violations, "contains_phone_numbers")
	}

	return violations
}

// hasExcessiveRepetition checks for patterns like "aaaa" or "1111"
func (cf *ContentFilter) hasExcessiveRepetition(text string) bool {
	runes := []rune(text)
	for i := 0; i < len(runes)-3; i++ {
		// Check if 4+ consecutive identical characters
		if runes[i] == runes[i+1] && runes[i] == runes[i+2] && runes[i] == runes[i+3] {
			// Allow repeated punctuation like "!!!" or "..."
			if !isPunctuation(runes[i]) {
				return true
			}
		}
	}
	return false
}

// isPunctuation checks if a rune is punctuation
func isPunctuation(r rune) bool {
	return unicode.IsPunct(r) || r == '!' || r == '?' || r == '.'
}

// hasExcessiveCapsLock checks if more than 50% of letters are uppercase
func (cf *ContentFilter) hasExcessiveCapsLock(text string) bool {
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

	// If more than 70% of letters are uppercase, it's considered excessive
	return (float64(upperCount) / float64(letterCount)) > 0.7
}

// containsLinks checks for URLs
func (cf *ContentFilter) containsLinks(text string) bool {
	return urlPattern.MatchString(text)
}

// containsPhoneNumbers checks for phone number patterns
func (cf *ContentFilter) containsPhoneNumbers(text string) bool {
	return phonePattern.MatchString(text)
}

// IsEnabled returns whether moderation is enabled
func (cf *ContentFilter) IsEnabled() bool {
	return cf.enabled
}

// ShouldAutoApprove returns whether auto-approve is enabled
func (cf *ContentFilter) ShouldAutoApprove() bool {
	return cf.autoApprove
}
