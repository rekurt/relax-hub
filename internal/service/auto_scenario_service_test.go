package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

func newAutoScenarioTestService() (service.AutoScenarioService, *mock.AutoScenarioRepo, *mock.GuestCardRepo) {
	scenarioRepo := mock.NewAutoScenarioRepo().(*mock.AutoScenarioRepo)
	gcRepo := mock.NewGuestCardRepo().(*mock.GuestCardRepo)
	notifSvc := &noopNotifService{}
	log := logger.New(logger.LevelWarn)
	return service.NewAutoScenarioService(scenarioRepo, gcRepo, notifSvc, log), scenarioRepo, gcRepo
}

func TestAutoScenarioService_ListScenarios_Defaults(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	scenarios, err := svc.ListScenarios(ctx, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("ListScenarios: %v", err)
	}

	// Should return all 5 predefined scenarios
	if len(scenarios) != 5 {
		t.Errorf("expected 5 scenarios, got %d", len(scenarios))
	}

	// All should be disabled by default
	for _, s := range scenarios {
		if s.Enabled {
			t.Errorf("scenario %s should be disabled by default", s.Type)
		}
	}
}

func TestAutoScenarioService_ListScenarios_ForbiddenForClient(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()

	_, err := svc.ListScenarios(ctx, uuid.New(), domain.RoleClient)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestAutoScenarioService_UpdateScenario(t *testing.T) {
	svc, scenarioRepo, _ := newAutoScenarioTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	scenario := &domain.AutoScenario{
		ID:         uuid.New(),
		Type:       domain.ScenarioThankAfterVisit,
		Enabled:    true,
		CustomText: "Спасибо, что пришли!",
		Channel:    domain.BroadcastChannelEmail,
		DelayHours: 3,
	}

	err := svc.UpdateScenario(ctx, ownerID, domain.RoleOwner, scenario)
	if err != nil {
		t.Fatalf("UpdateScenario: %v", err)
	}

	// Verify stored
	stored, err := scenarioRepo.GetByOwnerAndType(ctx, ownerID, domain.ScenarioThankAfterVisit)
	if err != nil {
		t.Fatalf("GetByOwnerAndType: %v", err)
	}
	if !stored.Enabled {
		t.Error("expected enabled=true")
	}
	if stored.CustomText != "Спасибо, что пришли!" {
		t.Errorf("expected custom text, got %s", stored.CustomText)
	}
	if stored.Channel != domain.BroadcastChannelEmail {
		t.Errorf("expected channel=email, got %s", stored.Channel)
	}
	if stored.DelayHours != 3 {
		t.Errorf("expected delay_hours=3, got %d", stored.DelayHours)
	}
}

func TestAutoScenarioService_UpdateScenario_ForbiddenForClient(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()

	scenario := &domain.AutoScenario{
		ID:      uuid.New(),
		Type:    domain.ScenarioThankAfterVisit,
		Channel: domain.BroadcastChannelPush,
	}

	err := svc.UpdateScenario(ctx, uuid.New(), domain.RoleClient, scenario)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestAutoScenarioService_UpdateScenario_InvalidType(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()

	scenario := &domain.AutoScenario{
		ID:      uuid.New(),
		Type:    domain.AutoScenarioType("nonexistent"),
		Channel: domain.BroadcastChannelPush,
	}

	err := svc.UpdateScenario(ctx, uuid.New(), domain.RoleOwner, scenario)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAutoScenarioService_UpdateScenario_InvalidChannel(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()

	scenario := &domain.AutoScenario{
		ID:      uuid.New(),
		Type:    domain.ScenarioThankAfterVisit,
		Channel: domain.BroadcastChannel("sms"),
	}

	err := svc.UpdateScenario(ctx, uuid.New(), domain.RoleOwner, scenario)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAutoScenarioService_UpdateScenario_Upsert(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Create
	scenario := &domain.AutoScenario{
		ID:         uuid.New(),
		Type:       domain.ScenarioRequestReview,
		Enabled:    true,
		CustomText: "Оставьте отзыв!",
		Channel:    domain.BroadcastChannelPush,
		DelayHours: 2,
	}
	err := svc.UpdateScenario(ctx, ownerID, domain.RoleOwner, scenario)
	if err != nil {
		t.Fatalf("UpdateScenario (create): %v", err)
	}

	// Update same type
	scenario2 := &domain.AutoScenario{
		ID:         uuid.New(),
		Type:       domain.ScenarioRequestReview,
		Enabled:    false,
		CustomText: "Расскажите о визите",
		Channel:    domain.BroadcastChannelTelegram,
		DelayHours: 4,
	}
	err = svc.UpdateScenario(ctx, ownerID, domain.RoleOwner, scenario2)
	if err != nil {
		t.Fatalf("UpdateScenario (upsert): %v", err)
	}

	// Verify list returns only 5 scenarios (one updated)
	scenarios, err := svc.ListScenarios(ctx, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("ListScenarios: %v", err)
	}
	if len(scenarios) != 5 {
		t.Errorf("expected 5 scenarios, got %d", len(scenarios))
	}

	// Find the updated one
	for _, s := range scenarios {
		if s.Type == domain.ScenarioRequestReview {
			if s.Enabled {
				t.Error("expected enabled=false after upsert")
			}
			if s.CustomText != "Расскажите о визите" {
				t.Errorf("expected updated custom text, got %s", s.CustomText)
			}
			if s.Channel != domain.BroadcastChannelTelegram {
				t.Errorf("expected channel=telegram, got %s", s.Channel)
			}
		}
	}
}

func TestAutoScenarioService_ListScenarios_MergesExistingWithDefaults(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Configure one scenario
	scenario := &domain.AutoScenario{
		ID:         uuid.New(),
		Type:       domain.ScenarioReactivateLost,
		Enabled:    true,
		CustomText: "Возвращайтесь!",
		Channel:    domain.BroadcastChannelEmail,
		DelayHours: 2160,
	}
	err := svc.UpdateScenario(ctx, ownerID, domain.RoleOwner, scenario)
	if err != nil {
		t.Fatalf("UpdateScenario: %v", err)
	}

	scenarios, err := svc.ListScenarios(ctx, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("ListScenarios: %v", err)
	}

	if len(scenarios) != 5 {
		t.Errorf("expected 5 scenarios, got %d", len(scenarios))
	}

	configuredCount := 0
	for _, s := range scenarios {
		if s.Type == domain.ScenarioReactivateLost {
			configuredCount++
			if !s.Enabled {
				t.Error("reactivate_lost should be enabled")
			}
			if s.CustomText != "Возвращайтесь!" {
				t.Errorf("expected custom text")
			}
		}
	}
	if configuredCount != 1 {
		t.Errorf("expected exactly 1 configured scenario, got %d", configuredCount)
	}
}

func TestAutoScenarioService_ExecuteScenarios(t *testing.T) {
	svc, scenarioRepo, gcRepo := newAutoScenarioTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	// Create a guest card with recent visit (90 minutes ago — within the 2h filter window)
	guestCardID := uuid.New()
	lastVisit := time.Now().Add(-90 * time.Minute)
	err := gcRepo.Upsert(ctx, &domain.GuestCard{
		ID:           guestCardID,
		OwnerID:      ownerID,
		ClientID:     clientID,
		BathhouseID:  bathhouseID,
		FirstVisitAt: lastVisit,
		LastVisitAt:  lastVisit,
		VisitCount:   1,
		TotalSpent:   500000,
		Tags:         []string{},
	})
	if err != nil {
		t.Fatalf("Upsert guest card: %v", err)
	}

	// Enable thank_after_visit scenario with 1 hour delay
	scenario := &domain.AutoScenario{
		ID:         uuid.New(),
		Type:       domain.ScenarioThankAfterVisit,
		Enabled:    true,
		Channel:    domain.BroadcastChannelPush,
		DelayHours: 1,
	}
	scenario.OwnerID = ownerID
	err = scenarioRepo.Upsert(ctx, scenario)
	if err != nil {
		t.Fatalf("Upsert scenario: %v", err)
	}

	// Execute
	sent, err := svc.ExecuteScenarios(ctx)
	if err != nil {
		t.Fatalf("ExecuteScenarios: %v", err)
	}

	if sent != 1 {
		t.Errorf("expected 1 sent, got %d", sent)
	}

	// Execute again - should not re-send (dedup)
	sent2, err := svc.ExecuteScenarios(ctx)
	if err != nil {
		t.Fatalf("ExecuteScenarios (2nd): %v", err)
	}
	if sent2 != 0 {
		t.Errorf("expected 0 sent on re-execution, got %d", sent2)
	}
}

func TestAutoScenarioService_ExecuteScenarios_DisabledNotExecuted(t *testing.T) {
	svc, scenarioRepo, gcRepo := newAutoScenarioTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Create a guest card
	err := gcRepo.Upsert(ctx, &domain.GuestCard{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		ClientID:     uuid.New(),
		BathhouseID:  uuid.New(),
		FirstVisitAt: time.Now().Add(-2 * time.Hour),
		LastVisitAt:  time.Now().Add(-2 * time.Hour),
		VisitCount:   1,
		TotalSpent:   100000,
		Tags:         []string{},
	})
	if err != nil {
		t.Fatalf("Upsert guest card: %v", err)
	}

	// Create disabled scenario
	scenario := &domain.AutoScenario{
		ID:         uuid.New(),
		Type:       domain.ScenarioThankAfterVisit,
		Enabled:    false,
		Channel:    domain.BroadcastChannelPush,
		DelayHours: 1,
	}
	scenario.OwnerID = ownerID
	err = scenarioRepo.Upsert(ctx, scenario)
	if err != nil {
		t.Fatalf("Upsert scenario: %v", err)
	}

	sent, err := svc.ExecuteScenarios(ctx)
	if err != nil {
		t.Fatalf("ExecuteScenarios: %v", err)
	}
	if sent != 0 {
		t.Errorf("expected 0 sent for disabled scenario, got %d", sent)
	}
}

func TestAutoScenarioService_ExecuteScenarios_NoGuests(t *testing.T) {
	svc, scenarioRepo, _ := newAutoScenarioTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Enable scenario but no guests
	scenario := &domain.AutoScenario{
		ID:         uuid.New(),
		Type:       domain.ScenarioThankAfterVisit,
		Enabled:    true,
		Channel:    domain.BroadcastChannelPush,
		DelayHours: 1,
	}
	scenario.OwnerID = ownerID
	err := scenarioRepo.Upsert(ctx, scenario)
	if err != nil {
		t.Fatalf("Upsert scenario: %v", err)
	}

	sent, err := svc.ExecuteScenarios(ctx)
	if err != nil {
		t.Fatalf("ExecuteScenarios: %v", err)
	}
	if sent != 0 {
		t.Errorf("expected 0 sent with no guests, got %d", sent)
	}
}

func TestAutoScenarioService_RepresentativeCanList(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()

	scenarios, err := svc.ListScenarios(ctx, uuid.New(), domain.RoleRepresentative)
	if err != nil {
		t.Fatalf("ListScenarios for representative: %v", err)
	}
	if len(scenarios) != 5 {
		t.Errorf("expected 5, got %d", len(scenarios))
	}
}

func TestAutoScenarioService_RepresentativeCanUpdate(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()

	scenario := &domain.AutoScenario{
		ID:      uuid.New(),
		Type:    domain.ScenarioBirthdayGreeting,
		Enabled: true,
		Channel: domain.BroadcastChannelPush,
	}
	err := svc.UpdateScenario(ctx, uuid.New(), domain.RoleRepresentative, scenario)
	if err != nil {
		t.Fatalf("UpdateScenario for representative: %v", err)
	}
}

func TestAutoScenarioService_NegativeDelayHours(t *testing.T) {
	svc, _, _ := newAutoScenarioTestService()
	ctx := context.Background()

	scenario := &domain.AutoScenario{
		ID:         uuid.New(),
		Type:       domain.ScenarioThankAfterVisit,
		Channel:    domain.BroadcastChannelPush,
		DelayHours: -1,
	}
	err := svc.UpdateScenario(ctx, uuid.New(), domain.RoleOwner, scenario)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for negative delay, got %v", err)
	}
}
