package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newBroadcastTestService() (service.BroadcastService, *mock.BroadcastRepo, *mock.GuestCardRepo) {
	broadcastRepo := mock.NewBroadcastRepo().(*mock.BroadcastRepo)
	gcRepo := mock.NewGuestCardRepo().(*mock.GuestCardRepo)
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	notifSvc := &noopNotifService{}
	log := logger.New(logger.LevelWarn)
	return service.NewBroadcastService(broadcastRepo, gcRepo, notifSvc, access, log), broadcastRepo, gcRepo
}

func TestBroadcastService_Create(t *testing.T) {
	svc, broadcastRepo, _ := newBroadcastTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentRegular,
		Title:    "Приглашаем вас!",
		Body:     "Скидка 20% на все услуги",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
	}

	err := svc.Create(ctx, ownerID, domain.RoleOwner, broadcast)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if broadcast.Status != domain.BroadcastStatusDraft {
		t.Errorf("expected status=draft, got %s", broadcast.Status)
	}
	if broadcast.OwnerID != ownerID {
		t.Errorf("expected owner_id=%s, got %s", ownerID, broadcast.OwnerID)
	}

	// Verify it was stored
	stored, err := broadcastRepo.GetByID(ctx, broadcast.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Title != "Приглашаем вас!" {
		t.Errorf("expected title='Приглашаем вас!', got %s", stored.Title)
	}
}

func TestBroadcastService_Create_ForbiddenForClient(t *testing.T) {
	svc, _, _ := newBroadcastTestService()
	ctx := context.Background()

	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentRegular,
		Title:    "Test",
		Body:     "Test body",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
	}

	err := svc.Create(ctx, uuid.New(), domain.RoleClient, broadcast)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestBroadcastService_Create_InvalidInput(t *testing.T) {
	svc, _, _ := newBroadcastTestService()
	ctx := context.Background()

	// Empty title
	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentRegular,
		Title:    "",
		Body:     "Test body",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
	}

	err := svc.Create(ctx, uuid.New(), domain.RoleOwner, broadcast)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}

	// Empty channels
	broadcast2 := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentRegular,
		Title:    "Test",
		Body:     "Test body",
		Channels: []domain.BroadcastChannel{},
	}

	err = svc.Create(ctx, uuid.New(), domain.RoleOwner, broadcast2)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for empty channels, got %v", err)
	}

	// Invalid segment
	broadcast3 := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.GuestSegmentSlug("nonexistent"),
		Title:    "Test",
		Body:     "Test body",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
	}

	err = svc.Create(ctx, uuid.New(), domain.RoleOwner, broadcast3)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for invalid segment, got %v", err)
	}
}

func TestBroadcastService_Send(t *testing.T) {
	svc, broadcastRepo, gcRepo := newBroadcastTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	// Create a guest card so there's someone to send to
	err := gcRepo.Upsert(ctx, &domain.GuestCard{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		ClientID:    clientID,
		BathhouseID: bathhouseID,
		VisitCount:  3,
		TotalSpent:  300000,
		Tags:        []string{},
	})
	if err != nil {
		t.Fatalf("Upsert guest card: %v", err)
	}

	// Create a broadcast
	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentRegular,
		Title:    "Скидка для постоянных!",
		Body:     "Только для вас — 20%",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
	}
	err = svc.Create(ctx, ownerID, domain.RoleOwner, broadcast)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Send it
	err = svc.Send(ctx, ownerID, domain.RoleOwner, broadcast.ID)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	// Check status changed to sent
	sent, err := broadcastRepo.GetByID(ctx, broadcast.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if sent.Status != domain.BroadcastStatusSent {
		t.Errorf("expected status=sent, got %s", sent.Status)
	}
	if sent.Delivered != 1 {
		t.Errorf("expected delivered=1, got %d", sent.Delivered)
	}
	if sent.SentAt == nil {
		t.Error("expected sent_at to be set")
	}
}

func TestBroadcastService_Send_NotDraft(t *testing.T) {
	svc, _, _ := newBroadcastTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentNew,
		Title:    "Test",
		Body:     "Test body",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelEmail},
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, broadcast)

	// Send once
	_ = svc.Send(ctx, ownerID, domain.RoleOwner, broadcast.ID)

	// Try sending again — should fail because it's no longer draft
	err := svc.Send(ctx, ownerID, domain.RoleOwner, broadcast.ID)
	if err != domain.ErrBroadcastNotDraft {
		t.Errorf("expected ErrBroadcastNotDraft, got %v", err)
	}
}

func TestBroadcastService_Send_RateLimit(t *testing.T) {
	svc, _, _ := newBroadcastTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Create and send 3 broadcasts (the weekly limit)
	for i := 0; i < 3; i++ {
		b := &domain.Broadcast{
			ID:       uuid.New(),
			Segment:  domain.SegmentNew,
			Title:    "Broadcast",
			Body:     "Body",
			Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
		}
		_ = svc.Create(ctx, ownerID, domain.RoleOwner, b)
		err := svc.Send(ctx, ownerID, domain.RoleOwner, b.ID)
		if err != nil {
			t.Fatalf("Send #%d: %v", i+1, err)
		}
	}

	// 4th should be rate limited
	b4 := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentNew,
		Title:    "Broadcast 4",
		Body:     "Body 4",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, b4)
	err := svc.Send(ctx, ownerID, domain.RoleOwner, b4.ID)
	if err != domain.ErrBroadcastRateLimit {
		t.Errorf("expected ErrBroadcastRateLimit, got %v", err)
	}
}

func TestBroadcastService_Send_ForbiddenOtherOwner(t *testing.T) {
	svc, _, _ := newBroadcastTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()

	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentNew,
		Title:    "Test",
		Body:     "Body",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, broadcast)

	// Another owner tries to send
	err := svc.Send(ctx, otherOwnerID, domain.RoleOwner, broadcast.ID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestBroadcastService_ListBroadcasts(t *testing.T) {
	svc, _, _ := newBroadcastTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Create 2 broadcasts
	for i := 0; i < 2; i++ {
		b := &domain.Broadcast{
			ID:       uuid.New(),
			Segment:  domain.SegmentNew,
			Title:    "Broadcast",
			Body:     "Body",
			Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
		}
		_ = svc.Create(ctx, ownerID, domain.RoleOwner, b)
	}

	result, err := svc.ListBroadcasts(ctx, ownerID, domain.RoleOwner, domain.BroadcastFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListBroadcasts: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected 2 broadcasts, got %d", result.TotalCount)
	}
}

func TestBroadcastService_GetBroadcast(t *testing.T) {
	svc, _, _ := newBroadcastTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentVIP,
		Title:    "VIP рассылка",
		Body:     "Особое предложение для VIP",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelEmail, domain.BroadcastChannelTelegram},
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, broadcast)

	got, err := svc.GetBroadcast(ctx, ownerID, domain.RoleOwner, broadcast.ID)
	if err != nil {
		t.Fatalf("GetBroadcast: %v", err)
	}
	if got.Title != "VIP рассылка" {
		t.Errorf("expected title='VIP рассылка', got %s", got.Title)
	}
	if len(got.Channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(got.Channels))
	}
}

func TestBroadcastService_GetBroadcast_NotFound(t *testing.T) {
	svc, _, _ := newBroadcastTestService()
	ctx := context.Background()

	_, err := svc.GetBroadcast(ctx, uuid.New(), domain.RoleOwner, uuid.New())
	if err != domain.ErrBroadcastNotFound {
		t.Errorf("expected ErrBroadcastNotFound, got %v", err)
	}
}

func TestBroadcastService_GetBroadcast_ForbiddenOtherOwner(t *testing.T) {
	svc, _, _ := newBroadcastTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentNew,
		Title:    "Test",
		Body:     "Body",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, broadcast)

	_, err := svc.GetBroadcast(ctx, uuid.New(), domain.RoleOwner, broadcast.ID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestBroadcastService_Send_NoGuests(t *testing.T) {
	svc, broadcastRepo, _ := newBroadcastTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.SegmentVIP,
		Title:    "Test",
		Body:     "Body",
		Channels: []domain.BroadcastChannel{domain.BroadcastChannelPush},
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, broadcast)

	// Send with no guests — should succeed but delivered=0
	err := svc.Send(ctx, ownerID, domain.RoleOwner, broadcast.ID)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	sent, _ := broadcastRepo.GetByID(ctx, broadcast.ID)
	if sent.Delivered != 0 {
		t.Errorf("expected delivered=0, got %d", sent.Delivered)
	}
	if sent.Status != domain.BroadcastStatusSent {
		t.Errorf("expected status=sent, got %s", sent.Status)
	}
}
