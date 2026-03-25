package antifraud_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/antifraud"
	"github.com/nikitaaldaev/bani/internal/logger"
)

func newTestFilter() antifraud.ChatFilter {
	log := logger.New(logger.LevelError)
	return antifraud.NewChatFilter(log)
}

func TestChatFilter_NoDetection(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Привет, хочу забронировать баню на субботу")
	if result.WasFiltered {
		t.Error("expected no filtering for normal text")
	}
	if result.Filtered != "Привет, хочу забронировать баню на субботу" {
		t.Errorf("text should not change, got: %q", result.Filtered)
	}
}

func TestChatFilter_PhoneRussian_Plus7(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
	}{
		{"standard", "Позвоните мне +7 999 123 45 67"},
		{"no spaces", "Мой номер +79991234567"},
		{"with dashes", "Звоните +7-999-123-45-67"},
		{"with parens", "Телефон +7(999)1234567"},
		{"mixed", "Номер: +7 (999) 123-45-67 для связи"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Filter(ctx, tt.input)
			if !result.WasFiltered {
				t.Errorf("expected filtering for: %q", tt.input)
			}
			found := false
			for _, d := range result.Detections {
				if d.Type == "phone" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected phone detection for: %q, detections: %v", tt.input, result.Detections)
			}
		})
	}
}

func TestChatFilter_PhoneRussian_8(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Звоните 8 999 123 45 67")
	if !result.WasFiltered {
		t.Error("expected filtering for 8-prefix phone")
	}
	found := false
	for _, d := range result.Detections {
		if d.Type == "phone" {
			found = true
		}
	}
	if !found {
		t.Error("expected phone detection")
	}
}

func TestChatFilter_Email(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
	}{
		{"simple", "Напишите на user@example.com"},
		{"with dots", "Мой email: ivan.petrov@mail.ru"},
		{"with plus", "Пишите user+tag@gmail.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Filter(ctx, tt.input)
			if !result.WasFiltered {
				t.Errorf("expected filtering for: %q", tt.input)
			}
			found := false
			for _, d := range result.Detections {
				if d.Type == "email" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected email detection for: %q", tt.input)
			}
		})
	}
}

func TestChatFilter_URLs(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
	}{
		{"https", "Посмотрите https://example.com/page"},
		{"http", "Ссылка http://example.com"},
		{"www", "Заходите на www.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Filter(ctx, tt.input)
			if !result.WasFiltered {
				t.Errorf("expected filtering for: %q", tt.input)
			}
			found := false
			for _, d := range result.Detections {
				if d.Type == "url" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected url detection for: %q", tt.input)
			}
		})
	}
}

func TestChatFilter_Telegram(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
	}{
		{"t.me link", "Пишите в t.me/myusername"},
		{"@ mention", "Мой телеграм @username123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Filter(ctx, tt.input)
			if !result.WasFiltered {
				t.Errorf("expected filtering for: %q", tt.input)
			}
			found := false
			for _, d := range result.Detections {
				if d.Type == "telegram" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected telegram detection for: %q, detections: %v", tt.input, result.Detections)
			}
		})
	}
}

func TestChatFilter_VK(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Добавляйтесь vk.com/id12345")
	if !result.WasFiltered {
		t.Error("expected filtering for VK link")
	}
	found := false
	for _, d := range result.Detections {
		if d.Type == "vk" {
			found = true
		}
	}
	if !found {
		t.Error("expected vk detection")
	}
}

func TestChatFilter_WhatsApp(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Пишите wa.me/79991234567")
	if !result.WasFiltered {
		t.Error("expected filtering for WhatsApp link")
	}
	found := false
	for _, d := range result.Detections {
		if d.Type == "whatsapp" {
			found = true
		}
	}
	if !found {
		t.Error("expected whatsapp detection")
	}
}

func TestChatFilter_MessengerKeywords(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
	}{
		{"напиши в вотсап", "Лучше напиши в вотсап"},
		{"напиши в whatsapp", "Давай напиши в whatsapp"},
		{"мой телеграм", "Вот мой телеграм"},
		{"напиши в телеграм", "Напиши в телеграм пожалуйста"},
		{"мой вайбер", "Вот мой вайбер"},
		{"пиши в лс", "Пиши в лс"},
		{"пиши в личку", "Пиши в личку"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Filter(ctx, tt.input)
			if !result.WasFiltered {
				t.Errorf("expected filtering for: %q", tt.input)
			}
			found := false
			for _, d := range result.Detections {
				if d.Type == "messenger_keyword" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected messenger_keyword detection for: %q, detections: %v", tt.input, result.Detections)
			}
		})
	}
}

func TestChatFilter_Replacement(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Звоните +7 999 123 45 67 или пишите user@mail.ru")
	if !result.WasFiltered {
		t.Fatal("expected filtering")
	}

	if len(result.Detections) < 2 {
		t.Fatalf("expected at least 2 detections, got %d", len(result.Detections))
	}

	// Both phone and email should be replaced
	expected := "Звоните [контактные данные скрыты] или пишите [контактные данные скрыты]"
	if result.Filtered != expected {
		t.Errorf("filtered text = %q, want %q", result.Filtered, expected)
	}
}

func TestChatFilter_MultipleDetections(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Телефон +79991234567, email test@test.com, телеграм @myuser")
	if !result.WasFiltered {
		t.Fatal("expected filtering")
	}
	if len(result.Detections) < 3 {
		t.Errorf("expected at least 3 detections, got %d", len(result.Detections))
	}
}

func TestChatFilter_LogFiltered(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	convID := uuid.New()
	senderID := uuid.New()

	result := f.Filter(ctx, "Мой номер +79991234567")
	f.LogFiltered(ctx, convID, senderID, "Мой номер +79991234567", result)

	items, total, err := f.ListFiltered(ctx, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 record, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ConversationID != convID {
		t.Errorf("conversation_id = %v, want %v", items[0].ConversationID, convID)
	}
	if items[0].SenderID != senderID {
		t.Errorf("sender_id = %v, want %v", items[0].SenderID, senderID)
	}
	if items[0].OriginalText != "Мой номер +79991234567" {
		t.Errorf("original_text = %q, want %q", items[0].OriginalText, "Мой номер +79991234567")
	}
}

func TestChatFilter_ListFiltered_Pagination(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		result := f.Filter(ctx, "Звоните +79991234567")
		f.LogFiltered(ctx, uuid.New(), uuid.New(), "Звоните +79991234567", result)
	}

	items, total, err := f.ListFiltered(ctx, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	// Page 3 should have 1 item
	items, _, err = f.ListFiltered(ctx, 3, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item on page 3, got %d", len(items))
	}
}

func TestChatFilter_ListFiltered_Empty(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	items, total, err := f.ListFiltered(ctx, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if items != nil {
		t.Errorf("expected nil items, got %v", items)
	}
}
