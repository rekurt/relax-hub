package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func TestSlotBlockRepo_CreateAndGetByID(t *testing.T) {
	repo := NewSlotBlockRepo()
	bathhouseID := uuid.New()

	block := &domain.SlotBlock{
		BathhouseID: bathhouseID,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(2 * time.Hour),
		Source:      domain.SlotBlockSourceManual,
		Description: "Test block",
	}

	if err := repo.Create(context.Background(), block); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if block.ID == uuid.Nil {
		t.Fatal("expected ID to be assigned")
	}

	got, err := repo.GetByID(context.Background(), block.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.BathhouseID != bathhouseID {
		t.Errorf("BathhouseID = %v, want %v", got.BathhouseID, bathhouseID)
	}
	if got.Description != "Test block" {
		t.Errorf("Description = %q, want %q", got.Description, "Test block")
	}
}

func TestSlotBlockRepo_Delete(t *testing.T) {
	repo := NewSlotBlockRepo()

	block := &domain.SlotBlock{
		BathhouseID: uuid.New(),
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		Source:      domain.SlotBlockSourceManual,
	}
	_ = repo.Create(context.Background(), block)

	if err := repo.Delete(context.Background(), block.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.GetByID(context.Background(), block.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got: %v", err)
	}
}

func TestSlotBlockRepo_ListByBathhouse(t *testing.T) {
	repo := NewSlotBlockRepo()
	bhID := uuid.New()
	otherBhID := uuid.New()

	for i := 0; i < 3; i++ {
		_ = repo.Create(context.Background(), &domain.SlotBlock{
			BathhouseID: bhID,
			StartTime:   time.Now().Add(time.Duration(i) * time.Hour),
			EndTime:     time.Now().Add(time.Duration(i+1) * time.Hour),
			Source:      domain.SlotBlockSourceManual,
		})
	}
	_ = repo.Create(context.Background(), &domain.SlotBlock{
		BathhouseID: otherBhID,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		Source:      domain.SlotBlockSourceManual,
	})

	blocks, err := repo.ListByBathhouse(context.Background(), bhID)
	if err != nil {
		t.Fatalf("ListByBathhouse failed: %v", err)
	}
	if len(blocks) != 3 {
		t.Errorf("expected 3 blocks, got %d", len(blocks))
	}
}

func TestSlotBlockRepo_GetByExternalID(t *testing.T) {
	repo := NewSlotBlockRepo()
	bhID := uuid.New()

	block := &domain.SlotBlock{
		BathhouseID: bhID,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		Source:      domain.SlotBlockSourceGoogleCalendar,
		ExternalID:  "google-event-123",
	}
	_ = repo.Create(context.Background(), block)

	got, err := repo.GetByExternalID(context.Background(), bhID, domain.SlotBlockSourceGoogleCalendar, "google-event-123")
	if err != nil {
		t.Fatalf("GetByExternalID failed: %v", err)
	}
	if got.ID != block.ID {
		t.Errorf("expected ID %v, got %v", block.ID, got.ID)
	}

	// Not found for different source
	_, err = repo.GetByExternalID(context.Background(), bhID, domain.SlotBlockSourceYandexCalendar, "google-event-123")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound for different source, got: %v", err)
	}
}

func TestSlotBlockRepo_DeleteBySource(t *testing.T) {
	repo := NewSlotBlockRepo()
	bhID := uuid.New()

	_ = repo.Create(context.Background(), &domain.SlotBlock{
		BathhouseID: bhID, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		Source: domain.SlotBlockSourceGoogleCalendar,
	})
	_ = repo.Create(context.Background(), &domain.SlotBlock{
		BathhouseID: bhID, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		Source: domain.SlotBlockSourceGoogleCalendar,
	})
	_ = repo.Create(context.Background(), &domain.SlotBlock{
		BathhouseID: bhID, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		Source: domain.SlotBlockSourceManual,
	})

	if err := repo.DeleteBySource(context.Background(), bhID, domain.SlotBlockSourceGoogleCalendar); err != nil {
		t.Fatalf("DeleteBySource failed: %v", err)
	}

	blocks, _ := repo.ListByBathhouse(context.Background(), bhID)
	if len(blocks) != 1 {
		t.Errorf("expected 1 block remaining, got %d", len(blocks))
	}
	if blocks[0].Source != domain.SlotBlockSourceManual {
		t.Errorf("remaining block source = %q, want %q", blocks[0].Source, domain.SlotBlockSourceManual)
	}
}

func TestSlotBlockRepo_HasOverlapping(t *testing.T) {
	repo := NewSlotBlockRepo()
	bhID := uuid.New()
	base := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	_ = repo.Create(context.Background(), &domain.SlotBlock{
		BathhouseID: bhID,
		StartTime:   base,
		EndTime:     base.Add(2 * time.Hour),
		Source:      domain.SlotBlockSourceManual,
	})

	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  bool
	}{
		{"exact overlap", base, base.Add(2 * time.Hour), true},
		{"partial overlap start", base.Add(-time.Hour), base.Add(time.Hour), true},
		{"partial overlap end", base.Add(time.Hour), base.Add(3 * time.Hour), true},
		{"no overlap before", base.Add(-2 * time.Hour), base, false},
		{"no overlap after", base.Add(2 * time.Hour), base.Add(4 * time.Hour), false},
		{"contained within", base.Add(30 * time.Minute), base.Add(90 * time.Minute), true},
		{"contains block", base.Add(-time.Hour), base.Add(3 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.HasOverlapping(context.Background(), bhID, tt.start, tt.end)
			if err != nil {
				t.Fatalf("HasOverlapping failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("HasOverlapping = %v, want %v", got, tt.want)
			}
		})
	}
}
