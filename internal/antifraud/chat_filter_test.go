package antifraud_test

import (
	"context"
	"strings"
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

	safe := []string{
		"Привет, хочу забронировать баню на субботу",
		"Можно приехать в 14:00?",
		"Сколько стоит на 3 часа?",
		"Отлично, бронирую",
		"Баня на 8 человек",
		"У вас 2 бассейна?",
		"Температура 80 градусов — норм?",
	}
	for _, text := range safe {
		result := f.Filter(ctx, text)
		if result.WasFiltered {
			t.Errorf("expected no filtering for %q, detections: %v", text, result.Detections)
		}
		if result.Filtered != text {
			t.Errorf("text changed: %q -> %q", text, result.Filtered)
		}
	}
}

// --- Phone number detection ---

func TestChatFilter_PhoneRussian_Plus7(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Позвоните мне +7 999 123 45 67",
		"Мой номер +79991234567",
		"Звоните +7-999-123-45-67",
		"Телефон +7(999)1234567",
		"Номер: +7 (999) 123-45-67 для связи",
	}

	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for: %q", input)
		}
		assertDetectionType(t, result, "phone", input)
	}
}

func TestChatFilter_PhoneRussian_8(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Звоните 8 999 123 45 67",
		"Телефон 89991234567",
		"Номер 8-999-123-45-67",
		"8(999)123-45-67 звоните",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for: %q", input)
		}
		assertDetectionType(t, result, "phone", input)
	}
}

func TestChatFilter_PhoneBelarus(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Звоните +375 29 123 45 67",
		"+375291234567",
		"+375-33-123-45-67",
		"+375 (44) 123-45-67",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for Belarus phone: %q", input)
		}
		assertDetectionType(t, result, "phone", input)
	}
}

func TestChatFilter_PhoneObfuscated(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"9.9.9.1.2.3.4.5.6.7",
		"9 9 9 1 2 3 4 5 6 7",
		"9-9-9-1-2-3-4-5-6-7",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for obfuscated phone: %q", input)
		}
		assertDetectionType(t, result, "phone", input)
	}
}

func TestChatFilter_PhoneWordsRussian(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"восемь девятьсот пятьдесят три двадцать четыре",
		"плюс семь девятьсот сто двадцать три",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for word-phone: %q", input)
		}
		assertDetectionType(t, result, "phone", input)
	}
}

// --- Email detection ---

func TestChatFilter_Email(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Напишите на user@example.com",
		"Мой email: ivan.petrov@mail.ru",
		"Пишите user+tag@gmail.com",
		"contact@yandex.ru для связи",
		"info@company.by",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for: %q", input)
		}
		assertDetectionType(t, result, "email", input)
	}
}

func TestChatFilter_EmailObfuscated(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"user собака mail точка ru",
		"ivan собачка gmail точка com",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for obfuscated email: %q", input)
		}
		assertDetectionType(t, result, "email", input)
	}
}

// --- URL detection ---

func TestChatFilter_URLs(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Посмотрите https://example.com/page",
		"Ссылка http://example.com",
		"Заходите на www.example.com",
		"Вот сайт https://my-site.ru/path?q=1",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for: %q", input)
		}
		assertDetectionType(t, result, "url", input)
	}
}

func TestChatFilter_BareDomains(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Заходите на example.com",
		"Мой сайт mybath.ru",
		"Смотрите site.org/page",
		"my-page.net",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for bare domain: %q", input)
		}
		assertDetectionType(t, result, "url", input)
	}
}

// --- Messenger links ---

func TestChatFilter_Telegram(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Пишите в t.me/myusername",
		"Мой телеграм @username123",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for: %q", input)
		}
		assertDetectionType(t, result, "telegram", input)
	}
}

func TestChatFilter_VK(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Добавляйтесь vk.com/id12345")
	if !result.WasFiltered {
		t.Error("expected filtering for VK link")
	}
	assertDetectionType(t, result, "vk", "vk.com/id12345")
}

func TestChatFilter_WhatsApp(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Пишите wa.me/79991234567")
	if !result.WasFiltered {
		t.Error("expected filtering for WhatsApp link")
	}
	assertDetectionType(t, result, "whatsapp", "wa.me/79991234567")
}

func TestChatFilter_Viber(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Пишите viber://chat?number=79991234567",
		"Добавляйтесь viber://add?number=375291234567",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for Viber link: %q", input)
		}
		assertDetectionType(t, result, "viber", input)
	}
}

func TestChatFilter_Instagram(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Мой инстаграм instagram.com/myprofile",
		"Подписывайтесь instagr.am/myprofile",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for Instagram link: %q", input)
		}
		assertDetectionType(t, result, "instagram", input)
	}
}

func TestChatFilter_OK(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Добавляйтесь ok.ru/profile/123456")
	if !result.WasFiltered {
		t.Error("expected filtering for OK link")
	}
	assertDetectionType(t, result, "ok", "ok.ru/profile/123456")
}

func TestChatFilter_WhatsAppGroup(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Вступайте chat.whatsapp.com/invite123abc")
	if !result.WasFiltered {
		t.Error("expected filtering for WhatsApp group link")
	}
	assertDetectionType(t, result, "whatsapp", "chat.whatsapp.com/invite123abc")
}

func TestChatFilter_MessengerKeywords(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	tests := []string{
		"Лучше напиши в вотсап",
		"Давай напиши в whatsapp",
		"Вот мой телеграм",
		"Напиши в телеграм пожалуйста",
		"Вот мой вайбер",
		"Пиши в лс",
		"Пиши в личку",
		"Свяжемся в телеграм",
		"Давай в вотсап",
		"Мой одноклассники",
	}
	for _, input := range tests {
		result := f.Filter(ctx, input)
		if !result.WasFiltered {
			t.Errorf("expected filtering for: %q", input)
		}
		assertDetectionType(t, result, "messenger_keyword", input)
	}
}

// --- Replacement behavior ---

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

	expected := "Звоните [контактные данные скрыты] или пишите [контактные данные скрыты]"
	if result.Filtered != expected {
		t.Errorf("filtered text = %q, want %q", result.Filtered, expected)
	}
}

func TestChatFilter_ReplacementPreservesText(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	result := f.Filter(ctx, "Баня на 3 часа, звоните +79991234567, буду в 15:00")
	if !result.WasFiltered {
		t.Fatal("expected filtering")
	}

	// The surrounding text must be preserved
	if !strings.Contains(result.Filtered, "Баня на 3 часа") {
		t.Error("leading text lost")
	}
	if !strings.Contains(result.Filtered, "буду в 15:00") {
		t.Error("trailing text lost")
	}
	if !strings.Contains(result.Filtered, "[контактные данные скрыты]") {
		t.Error("replacement text missing")
	}
	// Original phone must NOT be present
	if strings.Contains(result.Filtered, "999") {
		t.Error("phone number was not replaced")
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

// --- Logging ---

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
	if len(items[0].Detections) == 0 {
		t.Error("expected detections to be stored in log record")
	}
}

func TestChatFilter_LogFiltered_RecordsMultipleTypes(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	original := "Звоните +79991234567 или user@mail.ru"
	result := f.Filter(ctx, original)
	f.LogFiltered(ctx, uuid.New(), uuid.New(), original, result)

	items, _, _ := f.ListFiltered(ctx, 1, 10)
	if len(items) != 1 {
		t.Fatalf("expected 1 logged record, got %d", len(items))
	}
	types := map[string]bool{}
	for _, d := range items[0].Detections {
		types[d.Type] = true
	}
	if !types["phone"] || !types["email"] {
		t.Errorf("expected both phone and email in log detections, got: %v", items[0].Detections)
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

// --- Edge cases ---

func TestChatFilter_NotFalsePositive_ShortNumbers(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	// Short number sequences that are NOT phone numbers
	safe := []string{
		"Баня стоит 3500 рублей",
		"Код бронирования: 123456",
		"Заезд в 14:30",
		"3 гостя на 2 часа",
	}
	for _, text := range safe {
		result := f.Filter(ctx, text)
		// These might not trigger at all, or only in harmless ways.
		// The key test: the original text should largely survive.
		if result.WasFiltered {
			// If something was filtered, it should be a type we expect
			for _, d := range result.Detections {
				if d.Type == "phone" {
					t.Errorf("false positive phone detection in %q: matched %q", text, d.Match)
				}
			}
		}
	}
}

func TestChatFilter_MixedContent(t *testing.T) {
	f := newTestFilter()
	ctx := context.Background()

	input := "Привет! Мой номер +79991234567, почта user@test.com, пишите в t.me/mybot"
	result := f.Filter(ctx, input)
	if !result.WasFiltered {
		t.Fatal("expected filtering")
	}

	// Should have at least 3 detections (phone + email + telegram)
	types := map[string]bool{}
	for _, d := range result.Detections {
		types[d.Type] = true
	}
	for _, expected := range []string{"phone", "email", "telegram"} {
		if !types[expected] {
			t.Errorf("missing detection type %q in mixed content", expected)
		}
	}

	// Original contact info must be removed
	if strings.Contains(result.Filtered, "999") {
		t.Error("phone not replaced")
	}
	if strings.Contains(result.Filtered, "user@") {
		t.Error("email not replaced")
	}
}

// --- Helper ---

func assertDetectionType(t *testing.T, result antifraud.FilterResult, expectedType, input string) {
	t.Helper()
	for _, d := range result.Detections {
		if d.Type == expectedType {
			return
		}
	}
	t.Errorf("expected %q detection for: %q, got detections: %v", expectedType, input, result.Detections)
}
