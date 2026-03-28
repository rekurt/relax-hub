package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestPersonalizeMessage(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     BroadcastPersonalizationData
		want     string
	}{
		{
			name:     "replace guest name",
			template: "Привет, {{guest_name}}!",
			data:     BroadcastPersonalizationData{GuestName: "Иван"},
			want:     "Привет, Иван!",
		},
		{
			name:     "replace all tokens",
			template: "{{guest_name}}, вы были у нас {{visit_count}} раз. Последний визит: {{last_visit_date}}. Промокод: {{promo_code}}",
			data: BroadcastPersonalizationData{
				GuestName:     "Мария",
				VisitCount:    5,
				LastVisitDate: "15.03.2026",
				PromoCode:     "SPRING20",
			},
			want: "Мария, вы были у нас 5 раз. Последний визит: 15.03.2026. Промокод: SPRING20",
		},
		{
			name:     "no tokens in template",
			template: "Простое сообщение без токенов",
			data:     BroadcastPersonalizationData{GuestName: "Иван"},
			want:     "Простое сообщение без токенов",
		},
		{
			name:     "empty data uses defaults",
			template: "Привет, {{guest_name}}! Визитов: {{visit_count}}",
			data:     BroadcastPersonalizationData{},
			want:     "Привет, ! Визитов: 0",
		},
		{
			name:     "multiple same tokens",
			template: "{{guest_name}} и ещё раз {{guest_name}}",
			data:     BroadcastPersonalizationData{GuestName: "Пётр"},
			want:     "Пётр и ещё раз Пётр",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PersonalizeMessage(tt.template, tt.data)
			if got != tt.want {
				t.Errorf("PersonalizeMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBroadcast_Validate_SMSChannel(t *testing.T) {
	b := &Broadcast{
		ID:       uuid.New(),
		OwnerID:  uuid.New(),
		Segment:  SegmentRegular,
		Title:    "Test",
		Body:     "Body",
		Channels: []BroadcastChannel{BroadcastChannelSMS},
	}
	if err := b.Validate(); err != nil {
		t.Errorf("Validate with SMS channel should pass, got: %v", err)
	}
}

func TestBroadcast_HasChannel(t *testing.T) {
	b := &Broadcast{
		Channels: []BroadcastChannel{BroadcastChannelPush, BroadcastChannelSMS},
	}

	if !b.HasChannel(BroadcastChannelPush) {
		t.Error("expected HasChannel(push) = true")
	}
	if !b.HasChannel(BroadcastChannelSMS) {
		t.Error("expected HasChannel(sms) = true")
	}
	if b.HasChannel(BroadcastChannelEmail) {
		t.Error("expected HasChannel(email) = false")
	}
}

func TestAvailablePersonalizationTokens(t *testing.T) {
	tokens := AvailablePersonalizationTokens()
	expected := []string{"{{guest_name}}", "{{last_visit_date}}", "{{visit_count}}", "{{promo_code}}"}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tok := range tokens {
		if tok != expected[i] {
			t.Errorf("token[%d] = %q, want %q", i, tok, expected[i])
		}
	}
}
