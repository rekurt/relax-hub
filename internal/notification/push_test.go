package notification_test

import (
	"context"
	"testing"

	"github.com/nikitaaldaev/bani/internal/notification"
)

func TestNoopPushSender_Send(t *testing.T) {
	sender := notification.NewNoopPushSender()
	err := sender.Send(context.Background(), "{}", "Title", "Body", map[string]string{"key": "val"})
	if err != nil {
		t.Errorf("NoopPushSender should not return error, got: %v", err)
	}
}

func TestNoopPushSender_SendNilData(t *testing.T) {
	sender := notification.NewNoopPushSender()
	err := sender.Send(context.Background(), "{}", "Title", "Body", nil)
	if err != nil {
		t.Errorf("NoopPushSender should not return error, got: %v", err)
	}
}
