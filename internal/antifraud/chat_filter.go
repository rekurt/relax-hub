package antifraud

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/logger"
)

const (
	replacementText    = "[контактные данные скрыты]"
	maxFilteredRecords = 10000
)

// Detection describes a single detected contact info pattern in a message.
type Detection struct {
	Type    string `json:"type"`
	Match   string `json:"match"`
	Pattern string `json:"pattern"`
}

// FilterResult holds the outcome of filtering a chat message.
type FilterResult struct {
	Filtered    string
	WasFiltered bool
	Detections  []Detection
}

// FilteredChatMessage records a filtered chat message for admin review.
type FilteredChatMessage struct {
	ID             uuid.UUID   `json:"id"`
	ConversationID uuid.UUID   `json:"conversation_id"`
	SenderID       uuid.UUID   `json:"sender_id"`
	OriginalText   string      `json:"original_text"`
	FilteredText   string      `json:"filtered_text"`
	Detections     []Detection `json:"detections"`
	CreatedAt      time.Time   `json:"created_at"`
}

// ChatFilter detects and masks contact information in chat messages.
type ChatFilter interface {
	Filter(ctx context.Context, text string) FilterResult
	LogFiltered(ctx context.Context, conversationID, senderID uuid.UUID, original string, result FilterResult)
	ListFiltered(ctx context.Context, page, pageSize int) ([]FilteredChatMessage, int64, error)
}

type patternDef struct {
	name    string
	pattern string
	re      *regexp.Regexp
}

type chatFilter struct {
	patterns []patternDef
	logger   *logger.Logger

	mu      sync.RWMutex
	records []FilteredChatMessage
}

// NewChatFilter creates a new ChatFilter with detection patterns for phone numbers,
// emails, URLs, and messenger references. Patterns are ordered so that more specific
// matchers (e.g. t.me, vk.com) run before the generic URL pattern to avoid collisions.
func NewChatFilter(log *logger.Logger) ChatFilter {
	patterns := []patternDef{
		// --- Phones ---
		// Russian +7 / 8 prefix (10-digit body)
		{
			name:    "phone",
			pattern: "phone_ru",
			re:      regexp.MustCompile(`(?i)(?:\+7|8)[\s\-\(]*\d{3}[\s\-\)]*\d{3}[\s\-]*\d{2}[\s\-]*\d{2}`),
		},
		// Belarus +375 prefix (9-digit body: 2-digit operator + 7 digits)
		{
			name:    "phone",
			pattern: "phone_by",
			re:      regexp.MustCompile(`\+375[\s\-\(]*\d{2}[\s\-\)]*\d{3}[\s\-]*\d{2}[\s\-]*\d{2}`),
		},
		// Obfuscated digits: 10+ digits separated by dots, spaces, or dashes (e.g. "9.9.9.1.2.3.4.5.6.7")
		{
			name:    "phone",
			pattern: "phone_obfuscated",
			re:      regexp.MustCompile(`(?:\d[\s.\-]){9,}\d`),
		},
		// Phone numbers spelled in Russian words (e.g. "восемь девятьсот...")
		{
			name:    "phone",
			pattern: "phone_words_ru",
			re: regexp.MustCompile(`(?i)(?:восемь|плюс\s*семь)[\s,]*` +
				`(?:девятьсот|восемьсот|девятьсот)[\s,]*` +
				`(?:[\wа-яё]+[\s,]*){2,5}`),
		},
		// --- Email ---
		{
			name:    "email",
			pattern: "email",
			re:      regexp.MustCompile(`(?i)[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
		},
		// Email obfuscated with "собака"/"собачка" instead of @
		{
			name:    "email",
			pattern: "email_obfuscated",
			re:      regexp.MustCompile(`(?i)[a-zA-Z0-9._%+\-]+\s*(?:собак[аи]|собачк[аи])\s*[a-zA-Z0-9.\-]+\s*(?:точка|\.)\s*[a-zA-Z]{2,}`),
		},
		// --- Messenger-specific links (before generic URL) ---
		{
			name:    "telegram",
			pattern: "telegram_link",
			re:      regexp.MustCompile(`(?i)(?:t\.me/|@)[a-zA-Z0-9_]{3,}`),
		},
		{
			name:    "vk",
			pattern: "vk_link",
			re:      regexp.MustCompile(`(?i)vk\.com/[^\s]+`),
		},
		{
			name:    "whatsapp",
			pattern: "whatsapp_link",
			re:      regexp.MustCompile(`(?i)wa\.me/[^\s]+`),
		},
		// --- Generic URLs ---
		// With protocol or www prefix
		{
			name:    "url",
			pattern: "url",
			re:      regexp.MustCompile(`(?i)(?:https?://|www\.)[^\s]+`),
		},
		// Bare domain names (e.g. "example.com", "сайт.рф")
		{
			name:    "url",
			pattern: "url_bare_domain",
			re:      regexp.MustCompile(`(?i)\b[a-zA-Z0-9\-]+\.(?:com|ru|net|org|рф|by|info|biz|me|io|co)\b(?:/[^\s]*)?`),
		},
		// --- Messenger keywords ---
		{
			name:    "messenger_keyword",
			pattern: "messenger_keyword",
			re:      regexp.MustCompile(`(?i)(?:напиши\s+(?:в\s+)?(?:вотсап|whatsapp|ватсап|вацап|телегр(?:ам|амм)|вайбер|viber)|мой\s+(?:телегр(?:ам|амм)|вотсап|whatsapp|ватсап|вацап|вайбер|viber|инст(?:а|аграм)|instagram)|пиши\s+в\s+(?:лс|личк[уи]|директ|direct))`),
		},
	}

	return &chatFilter{
		patterns: patterns,
		logger:   log,
	}
}

func (f *chatFilter) Filter(_ context.Context, text string) FilterResult {
	var detections []Detection
	filtered := text

	for _, p := range f.patterns {
		matches := p.re.FindAllString(filtered, -1)
		for _, m := range matches {
			detections = append(detections, Detection{
				Type:    p.name,
				Match:   m,
				Pattern: p.pattern,
			})
		}
		filtered = p.re.ReplaceAllString(filtered, replacementText)
	}

	return FilterResult{
		Filtered:    filtered,
		WasFiltered: len(detections) > 0,
		Detections:  detections,
	}
}

func (f *chatFilter) LogFiltered(_ context.Context, conversationID, senderID uuid.UUID, original string, result FilterResult) {
	record := FilteredChatMessage{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		OriginalText:   original,
		FilteredText:   result.Filtered,
		Detections:     result.Detections,
		CreatedAt:      time.Now(),
	}

	f.mu.Lock()
	f.records = append(f.records, record)
	// Cap in-memory records to prevent unbounded growth.
	if len(f.records) > maxFilteredRecords {
		f.records = f.records[len(f.records)-maxFilteredRecords:]
	}
	f.mu.Unlock()

	types := make([]string, len(result.Detections))
	for i, d := range result.Detections {
		types[i] = d.Type
	}

	f.logger.Warn("chat message filtered",
		"conversation_id", conversationID,
		"sender_id", senderID,
		"detection_types", strings.Join(types, ","),
		"detection_count", len(result.Detections),
	)
}

func (f *chatFilter) ListFiltered(_ context.Context, page, pageSize int) ([]FilteredChatMessage, int64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	total := int64(len(f.records))
	start := (page - 1) * pageSize
	if start >= len(f.records) {
		return nil, total, nil
	}

	// Return in reverse chronological order.
	reversed := make([]FilteredChatMessage, len(f.records))
	for i, r := range f.records {
		reversed[len(f.records)-1-i] = r
	}

	end := start + pageSize
	if end > len(reversed) {
		end = len(reversed)
	}

	return reversed[start:end], total, nil
}
