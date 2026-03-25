package service_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newGuestCardTestService() (service.GuestCardService, *mock.GuestCardRepo) {
	gcRepo := mock.NewGuestCardRepo().(*mock.GuestCardRepo)
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	return service.NewGuestCardService(gcRepo, access, log), gcRepo
}

func TestGuestCardService_RecordVisit(t *testing.T) {
	svc, gcRepo := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	// First visit
	err := svc.RecordVisit(ctx, ownerID, clientID, bathhouseID, 500000) // 5000 RUB
	if err != nil {
		t.Fatalf("RecordVisit: %v", err)
	}

	card, err := gcRepo.GetByOwnerAndClient(ctx, ownerID, clientID, bathhouseID)
	if err != nil {
		t.Fatalf("GetByOwnerAndClient: %v", err)
	}

	if card.VisitCount != 1 {
		t.Errorf("expected visit_count=1, got %d", card.VisitCount)
	}
	if card.TotalSpent != 500000 {
		t.Errorf("expected total_spent=500000, got %d", card.TotalSpent)
	}
	if card.AvgCheck != 500000 {
		t.Errorf("expected avg_check=500000, got %d", card.AvgCheck)
	}

	// Second visit
	err = svc.RecordVisit(ctx, ownerID, clientID, bathhouseID, 300000) // 3000 RUB
	if err != nil {
		t.Fatalf("RecordVisit second: %v", err)
	}

	card, err = gcRepo.GetByOwnerAndClient(ctx, ownerID, clientID, bathhouseID)
	if err != nil {
		t.Fatalf("GetByOwnerAndClient second: %v", err)
	}

	if card.VisitCount != 2 {
		t.Errorf("expected visit_count=2, got %d", card.VisitCount)
	}
	if card.TotalSpent != 800000 {
		t.Errorf("expected total_spent=800000, got %d", card.TotalSpent)
	}
	if card.AvgCheck != 400000 {
		t.Errorf("expected avg_check=400000, got %d", card.AvgCheck)
	}
}

func TestGuestCardService_ListGuests(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()

	// Record 3 different guests
	for i := 0; i < 3; i++ {
		err := svc.RecordVisit(ctx, ownerID, uuid.New(), bathhouseID, int64(100000*(i+1)))
		if err != nil {
			t.Fatalf("RecordVisit: %v", err)
		}
	}

	result, err := svc.ListGuests(ctx, ownerID, domain.RoleOwner, domain.GuestCardFilter{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListGuests: %v", err)
	}

	if result.TotalCount != 3 {
		t.Errorf("expected total_count=3, got %d", result.TotalCount)
	}
	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
}

func TestGuestCardService_ListGuests_ForbiddenForClient(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	_, err := svc.ListGuests(ctx, uuid.New(), domain.RoleClient, domain.GuestCardFilter{})
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGuestCardService_UpdateNotes(t *testing.T) {
	svc, gcRepo := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	err := svc.RecordVisit(ctx, ownerID, clientID, bathhouseID, 500000)
	if err != nil {
		t.Fatalf("RecordVisit: %v", err)
	}

	card, _ := gcRepo.GetByOwnerAndClient(ctx, ownerID, clientID, bathhouseID)

	err = svc.UpdateGuestNotes(ctx, ownerID, domain.RoleOwner, card.ID, "VIP guest", []string{"vip", "regular"})
	if err != nil {
		t.Fatalf("UpdateGuestNotes: %v", err)
	}

	updated, _ := gcRepo.GetByOwnerAndClient(ctx, ownerID, clientID, bathhouseID)
	if updated.Notes != "VIP guest" {
		t.Errorf("expected notes='VIP guest', got '%s'", updated.Notes)
	}
	if len(updated.Tags) != 2 || updated.Tags[0] != "vip" {
		t.Errorf("expected tags=[vip, regular], got %v", updated.Tags)
	}
}

func TestGuestCardService_UpdateNotes_ForbiddenForClient(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	err := svc.UpdateGuestNotes(ctx, uuid.New(), domain.RoleClient, uuid.New(), "test", nil)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGuestCardService_ExportCSV(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()

	// Create some data
	err := svc.RecordVisit(ctx, ownerID, uuid.New(), bathhouseID, 500000)
	if err != nil {
		t.Fatalf("RecordVisit: %v", err)
	}

	var buf bytes.Buffer
	err = svc.ExportCSV(ctx, ownerID, domain.RoleOwner, domain.GuestCardFilter{}, &buf)
	if err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty CSV output")
	}

	// Check header
	if !bytes.Contains(buf.Bytes(), []byte("client_id")) {
		t.Error("CSV missing header")
	}
}

func TestGuestCardService_ExportCSV_ForbiddenForClient(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	var buf bytes.Buffer
	err := svc.ExportCSV(ctx, uuid.New(), domain.RoleClient, domain.GuestCardFilter{}, &buf)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGuestCardService_GetStats(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()

	// Create 2 guest cards
	err := svc.RecordVisit(ctx, ownerID, uuid.New(), bathhouseID, 400000)
	if err != nil {
		t.Fatalf("RecordVisit: %v", err)
	}
	err = svc.RecordVisit(ctx, ownerID, uuid.New(), bathhouseID, 600000)
	if err != nil {
		t.Fatalf("RecordVisit: %v", err)
	}

	stats, err := svc.GetStats(ctx, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	if stats.TotalGuests != 2 {
		t.Errorf("expected total_guests=2, got %d", stats.TotalGuests)
	}
}

func TestGuestCardService_GetStats_ForbiddenForClient(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	_, err := svc.GetStats(ctx, uuid.New(), domain.RoleClient)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGuestCardService_ListGuests_WithTagFilter(t *testing.T) {
	svc, gcRepo := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()

	client1 := uuid.New()
	client2 := uuid.New()

	// Create 2 guest cards
	_ = svc.RecordVisit(ctx, ownerID, client1, bathhouseID, 500000)
	_ = svc.RecordVisit(ctx, ownerID, client2, bathhouseID, 300000)

	// Tag only client1
	card1, _ := gcRepo.GetByOwnerAndClient(ctx, ownerID, client1, bathhouseID)
	_ = gcRepo.UpdateNotes(ctx, card1.ID, "VIP", []string{"vip"})

	tag := "vip"
	result, err := svc.ListGuests(ctx, ownerID, domain.RoleOwner, domain.GuestCardFilter{
		Tag:      &tag,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListGuests with tag: %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("expected total_count=1 for tag filter, got %d", result.TotalCount)
	}
}

func TestGuestCardService_MultipleVisitsSameGuest(t *testing.T) {
	svc, gcRepo := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	// 5 visits
	amounts := []int64{100000, 200000, 300000, 400000, 500000}
	for _, a := range amounts {
		err := svc.RecordVisit(ctx, ownerID, clientID, bathhouseID, a)
		if err != nil {
			t.Fatalf("RecordVisit: %v", err)
		}
	}

	card, err := gcRepo.GetByOwnerAndClient(ctx, ownerID, clientID, bathhouseID)
	if err != nil {
		t.Fatalf("GetByOwnerAndClient: %v", err)
	}

	if card.VisitCount != 5 {
		t.Errorf("expected visit_count=5, got %d", card.VisitCount)
	}

	expectedTotal := int64(1500000)
	if card.TotalSpent != expectedTotal {
		t.Errorf("expected total_spent=%d, got %d", expectedTotal, card.TotalSpent)
	}

	expectedAvg := expectedTotal / 5
	if card.AvgCheck != expectedAvg {
		t.Errorf("expected avg_check=%d, got %d", expectedAvg, card.AvgCheck)
	}
}

func TestGuestCardService_ListSegments(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()

	// Create guests with different profiles:
	// client1: 1 visit (new segment)
	client1 := uuid.New()
	_ = svc.RecordVisit(ctx, ownerID, client1, bathhouseID, 100000)

	// client2: 3 visits (regular segment)
	client2 := uuid.New()
	_ = svc.RecordVisit(ctx, ownerID, client2, bathhouseID, 200000)
	_ = svc.RecordVisit(ctx, ownerID, client2, bathhouseID, 200000)
	_ = svc.RecordVisit(ctx, ownerID, client2, bathhouseID, 200000)

	// client3: VIP (> 50,000 RUB = 5,000,000 kopecks)
	client3 := uuid.New()
	_ = svc.RecordVisit(ctx, ownerID, client3, bathhouseID, 5100000)

	segments, err := svc.ListSegments(ctx, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("ListSegments: %v", err)
	}

	if len(segments) != 5 {
		t.Fatalf("expected 5 segments, got %d", len(segments))
	}

	// Build a map for easier assertions
	segMap := make(map[domain.GuestSegmentSlug]int64)
	for _, s := range segments {
		segMap[s.Slug] = s.Count
	}

	if segMap[domain.SegmentNew] != 2 { // client1 (1 visit) and client3 (1 visit)
		t.Errorf("expected new=2, got %d", segMap[domain.SegmentNew])
	}
	if segMap[domain.SegmentRegular] != 1 { // client2
		t.Errorf("expected regular=1, got %d", segMap[domain.SegmentRegular])
	}
	if segMap[domain.SegmentVIP] != 1 { // client3
		t.Errorf("expected vip=1, got %d", segMap[domain.SegmentVIP])
	}
	if segMap[domain.SegmentLost] != 0 { // no one is lost (all recent)
		t.Errorf("expected lost=0, got %d", segMap[domain.SegmentLost])
	}
	if segMap[domain.SegmentBirthdaySoon] != 0 { // no birthday data
		t.Errorf("expected birthday_soon=0, got %d", segMap[domain.SegmentBirthdaySoon])
	}
}

func TestGuestCardService_ListSegments_ForbiddenForClient(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	_, err := svc.ListSegments(ctx, uuid.New(), domain.RoleClient)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGuestCardService_GetGuestsInSegment(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()

	// client1: 1 visit (new)
	_ = svc.RecordVisit(ctx, ownerID, uuid.New(), bathhouseID, 100000)

	// client2: 3 visits (regular)
	client2 := uuid.New()
	_ = svc.RecordVisit(ctx, ownerID, client2, bathhouseID, 200000)
	_ = svc.RecordVisit(ctx, ownerID, client2, bathhouseID, 200000)
	_ = svc.RecordVisit(ctx, ownerID, client2, bathhouseID, 200000)

	// Get new segment guests
	result, err := svc.GetGuestsInSegment(ctx, ownerID, domain.RoleOwner, domain.SegmentNew, 1, 10)
	if err != nil {
		t.Fatalf("GetGuestsInSegment new: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 new guest, got %d", result.TotalCount)
	}

	// Get regular segment guests
	result, err = svc.GetGuestsInSegment(ctx, ownerID, domain.RoleOwner, domain.SegmentRegular, 1, 10)
	if err != nil {
		t.Fatalf("GetGuestsInSegment regular: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 regular guest, got %d", result.TotalCount)
	}
	if len(result.Items) > 0 && result.Items[0].VisitCount < 3 {
		t.Errorf("expected visit_count >= 3, got %d", result.Items[0].VisitCount)
	}
}

func TestGuestCardService_GetGuestsInSegment_ForbiddenForClient(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	_, err := svc.GetGuestsInSegment(ctx, uuid.New(), domain.RoleClient, domain.SegmentNew, 1, 10)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGuestCardService_GetGuestsInSegment_EmptySegment(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	ownerID := uuid.New()

	result, err := svc.GetGuestsInSegment(ctx, ownerID, domain.RoleOwner, domain.SegmentVIP, 1, 10)
	if err != nil {
		t.Fatalf("GetGuestsInSegment: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected 0 VIP guests, got %d", result.TotalCount)
	}
}

func TestGuestCardService_ListSegments_SegmentMetadata(t *testing.T) {
	svc, _ := newGuestCardTestService()
	ctx := context.Background()

	segments, err := svc.ListSegments(ctx, uuid.New(), domain.RoleOwner)
	if err != nil {
		t.Fatalf("ListSegments: %v", err)
	}

	// Verify all segments have name and description
	for _, seg := range segments {
		if seg.Name == "" {
			t.Errorf("segment %s has empty name", seg.Slug)
		}
		if seg.Description == "" {
			t.Errorf("segment %s has empty description", seg.Slug)
		}
	}
}
