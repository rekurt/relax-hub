package seo

import (
	"fmt"
	"regexp"
	"strings"
)

var translitMap = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
	'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i",
	'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
	'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
	'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch",
	'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
	'э': "e", 'ю': "yu", 'я': "ya",

	'А': "a", 'Б': "b", 'В': "v", 'Г': "g", 'Д': "d",
	'Е': "e", 'Ё': "yo", 'Ж': "zh", 'З': "z", 'И': "i",
	'Й': "y", 'К': "k", 'Л': "l", 'М': "m", 'Н': "n",
	'О': "o", 'П': "p", 'Р': "r", 'С': "s", 'Т': "t",
	'У': "u", 'Ф': "f", 'Х': "kh", 'Ц': "ts", 'Ч': "ch",
	'Ш': "sh", 'Щ': "shch", 'Ъ': "", 'Ы': "y", 'Ь': "",
	'Э': "e", 'Ю': "yu", 'Я': "ya",
}

var (
	nonAlphanumRegex  = regexp.MustCompile(`[^a-z0-9-]+`)
	multiDashRegex    = regexp.MustCompile(`-{2,}`)
	leadTrailDashTrim = regexp.MustCompile(`^-+|-+$`)
)

// Transliterate converts a Russian string to Latin transliteration.
func Transliterate(s string) string {
	var b strings.Builder
	for _, r := range s {
		if mapped, ok := translitMap[r]; ok {
			b.WriteString(mapped)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// GenerateSlug creates a URL-friendly slug from a name.
// It transliterates Russian characters, lowercases, replaces spaces/special chars with dashes.
func GenerateSlug(name string) string {
	slug := strings.ToLower(Transliterate(name))
	slug = nonAlphanumRegex.ReplaceAllString(slug, "-")
	slug = multiDashRegex.ReplaceAllString(slug, "-")
	slug = leadTrailDashTrim.ReplaceAllString(slug, "")
	// Truncate to leave room for numeric suffixes (e.g., "-999")
	if len(slug) > 200 {
		slug = slug[:200]
		slug = leadTrailDashTrim.ReplaceAllString(slug, "")
	}
	return slug
}

// GenerateUniqueSlug creates a unique slug by appending a numeric suffix if needed.
// The exists function checks whether a given slug is already taken.
func GenerateUniqueSlug(name string, exists func(slug string) (bool, error)) (string, error) {
	base := GenerateSlug(name)
	if base == "" {
		base = "bathhouse"
	}

	taken, err := exists(base)
	if err != nil {
		return "", fmt.Errorf("check slug existence: %w", err)
	}
	if !taken {
		return base, nil
	}

	for i := 2; i <= 1000; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		taken, err := exists(candidate)
		if err != nil {
			return "", fmt.Errorf("check slug existence: %w", err)
		}
		if !taken {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not generate unique slug for %q after 1000 attempts", name)
}

// IsValidSlug checks if a string is a valid slug format (lowercase ASCII alphanumeric and dashes).
func IsValidSlug(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return true
}
