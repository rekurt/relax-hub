package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestConversation_Validate(t *testing.T) {
	valid := &Conversation{
		BathhouseID: uuid.New(),
		ClientID:    uuid.New(),
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid conversation returned error: %v", err)
	}

	tests := []struct {
		name string
		c    Conversation
	}{
		{"nil bathhouse", Conversation{BathhouseID: uuid.Nil, ClientID: uuid.New()}},
		{"nil client", Conversation{BathhouseID: uuid.New(), ClientID: uuid.Nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Validate(); err == nil {
				t.Error("expected error for invalid conversation")
			}
		})
	}
}

func TestMessage_Validate(t *testing.T) {
	valid := &Message{
		ConversationID: uuid.New(),
		SenderID:       uuid.New(),
		Text:           "Hello",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid message returned error: %v", err)
	}

	tests := []struct {
		name string
		m    Message
	}{
		{"nil conversation", Message{ConversationID: uuid.Nil, SenderID: uuid.New(), Text: "hi"}},
		{"nil sender", Message{ConversationID: uuid.New(), SenderID: uuid.Nil, Text: "hi"}},
		{"empty text", Message{ConversationID: uuid.New(), SenderID: uuid.New(), Text: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.Validate(); err == nil {
				t.Error("expected error for invalid message")
			}
		})
	}
}
