package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSlotBlock_Validate(t *testing.T) {
	validBlock := func() *SlotBlock {
		return &SlotBlock{
			BathhouseID: uuid.New(),
			StartTime:   time.Now(),
			EndTime:     time.Now().Add(2 * time.Hour),
			Source:      SlotBlockSourceManual,
		}
	}

	tests := []struct {
		name    string
		modify  func(sb *SlotBlock)
		wantErr bool
	}{
		{"valid block", func(sb *SlotBlock) {}, false},
		{"valid google calendar", func(sb *SlotBlock) { sb.Source = SlotBlockSourceGoogleCalendar }, false},
		{"valid yandex calendar", func(sb *SlotBlock) { sb.Source = SlotBlockSourceYandexCalendar }, false},
		{"missing bathhouse ID", func(sb *SlotBlock) { sb.BathhouseID = uuid.Nil }, true},
		{"end before start", func(sb *SlotBlock) { sb.EndTime = sb.StartTime.Add(-time.Hour) }, true},
		{"end equals start", func(sb *SlotBlock) { sb.EndTime = sb.StartTime }, true},
		{"empty source", func(sb *SlotBlock) { sb.Source = "" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := validBlock()
			tt.modify(sb)
			err := sb.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
