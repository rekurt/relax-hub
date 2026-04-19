package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

type AutoScenarioService interface {
	ListScenarios(ctx context.Context, userID uuid.UUID, role domain.UserRole) ([]domain.AutoScenario, error)
	UpdateScenario(ctx context.Context, userID uuid.UUID, role domain.UserRole, scenario *domain.AutoScenario) error
	ExecuteScenarios(ctx context.Context) (int, error)
}

type autoScenarioService struct {
	scenarioRepo  repository.AutoScenarioRepository
	guestCardRepo repository.GuestCardRepository
	notifSvc      NotificationService
	logger        *logger.Logger
}

func NewAutoScenarioService(
	scenarioRepo repository.AutoScenarioRepository,
	guestCardRepo repository.GuestCardRepository,
	notifSvc NotificationService,
	log *logger.Logger,
) AutoScenarioService {
	return &autoScenarioService{
		scenarioRepo:  scenarioRepo,
		guestCardRepo: guestCardRepo,
		notifSvc:      notifSvc,
		logger:        log,
	}
}

func (s *autoScenarioService) ListScenarios(ctx context.Context, userID uuid.UUID, role domain.UserRole) ([]domain.AutoScenario, error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	existing, err := s.scenarioRepo.ListByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list auto scenarios: %w", err)
	}

	// Build map of existing scenarios
	existingMap := make(map[domain.AutoScenarioType]*domain.AutoScenario)
	for i := range existing {
		existingMap[existing[i].Type] = &existing[i]
	}

	// Return all predefined scenarios, filling defaults for unconfigured ones
	var result []domain.AutoScenario
	for _, st := range domain.AllScenarioTypes() {
		if s, ok := existingMap[st]; ok {
			result = append(result, *s)
		} else {
			_, _, defaultDelay := domain.ScenarioMeta(st)
			result = append(result, domain.AutoScenario{
				OwnerID:    userID,
				Type:       st,
				Enabled:    false,
				Channel:    domain.BroadcastChannelPush,
				DelayHours: defaultDelay,
			})
		}
	}

	return result, nil
}

func (s *autoScenarioService) UpdateScenario(ctx context.Context, userID uuid.UUID, role domain.UserRole, scenario *domain.AutoScenario) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	scenario.OwnerID = userID
	if err := scenario.Validate(); err != nil {
		return err
	}

	return s.scenarioRepo.Upsert(ctx, scenario)
}

func (s *autoScenarioService) ExecuteScenarios(ctx context.Context) (int, error) {
	scenarios, err := s.scenarioRepo.ListEnabled(ctx)
	if err != nil {
		return 0, fmt.Errorf("list enabled scenarios: %w", err)
	}

	totalSent := 0
	now := time.Now()

	for _, scenario := range scenarios {
		sent, err := s.executeScenario(ctx, scenario, now)
		if err != nil {
			s.logger.Error("execute auto scenario", "type", scenario.Type, "owner_id", scenario.OwnerID, "error", err)
			continue
		}
		totalSent += sent
	}

	return totalSent, nil
}

func (s *autoScenarioService) executeScenario(ctx context.Context, scenario domain.AutoScenario, now time.Time) (int, error) {
	// Get all guests for this owner
	filter := domain.GuestCardFilter{
		OwnerID:  scenario.OwnerID,
		Page:     1,
		PageSize: 10000,
	}

	// Apply segment filter based on scenario type
	switch scenario.Type {
	case domain.ScenarioThankAfterVisit:
		// Guests with recent visits: fetch window must cover the full shouldExecute range
		// shouldExecute checks [LastVisitAt + delayHours, LastVisitAt + delayHours + 2h)
		// so LastVisitAt must be >= now - delayHours - 2h
		cutoff := now.Add(-time.Duration(scenario.DelayHours+2) * time.Hour)
		filter.DateFrom = &cutoff
	case domain.ScenarioRequestReview:
		cutoff := now.Add(-time.Duration(scenario.DelayHours+2) * time.Hour)
		filter.DateFrom = &cutoff
	case domain.ScenarioRemindRevisit30d:
		seg := domain.SegmentRegular
		filter.Segment = &seg
	case domain.ScenarioReactivateLost:
		seg := domain.SegmentLost
		filter.Segment = &seg
	case domain.ScenarioBirthdayGreeting:
		seg := domain.SegmentBirthdaySoon
		filter.Segment = &seg
	}

	guests, err := s.guestCardRepo.ListByOwner(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("list guests: %w", err)
	}

	sent := 0
	for _, guest := range guests.Items {
		// Check if already executed for this guest
		executed, err := s.scenarioRepo.HasBeenExecuted(ctx, scenario.ID, guest.ID)
		if err != nil {
			s.logger.Error("check scenario execution", "scenario_id", scenario.ID, "guest_id", guest.ID, "error", err)
			continue
		}
		if executed {
			continue
		}

		// For time-based scenarios, check the delay
		if !s.shouldExecute(scenario, guest, now) {
			continue
		}

		// Build notification
		title, body := s.buildMessage(scenario)

		err = s.notifSvc.Send(ctx, guest.ClientID, domain.NotifAutoScenario, title, body, map[string]string{
			"scenario_type": string(scenario.Type),
			"owner_id":      scenario.OwnerID.String(),
			"bathhouse_id":  guest.BathhouseID.String(),
		})
		if err != nil {
			s.logger.Error("send auto scenario notification", "client_id", guest.ClientID, "scenario_type", scenario.Type, "error", err)
			continue
		}

		// Record execution
		if err := s.scenarioRepo.RecordExecution(ctx, scenario.ID, guest.ID); err != nil {
			s.logger.Error("record scenario execution", "scenario_id", scenario.ID, "guest_id", guest.ID, "error", err)
		}

		sent++
	}

	return sent, nil
}

func (s *autoScenarioService) shouldExecute(scenario domain.AutoScenario, guest domain.GuestCard, now time.Time) bool {
	switch scenario.Type {
	case domain.ScenarioThankAfterVisit, domain.ScenarioRequestReview:
		// Execute if last visit was delay_hours ago (within a 1-hour window)
		targetTime := guest.LastVisitAt.Add(time.Duration(scenario.DelayHours) * time.Hour)
		return now.After(targetTime) && now.Before(targetTime.Add(2*time.Hour))
	case domain.ScenarioRemindRevisit30d:
		// Execute if last visit was ~30 days ago
		daysSinceVisit := now.Sub(guest.LastVisitAt).Hours() / 24
		return daysSinceVisit >= 29 && daysSinceVisit <= 31
	case domain.ScenarioReactivateLost:
		// Execute for guests with >90 days since last visit
		daysSinceVisit := now.Sub(guest.LastVisitAt).Hours() / 24
		return daysSinceVisit >= 90
	case domain.ScenarioBirthdayGreeting:
		// Birthday greeting — always execute for guests in the birthday_soon segment
		return true
	}
	return false
}

func (s *autoScenarioService) buildMessage(scenario domain.AutoScenario) (title, body string) {
	if scenario.CustomText != "" {
		return s.defaultTitle(scenario.Type), scenario.CustomText
	}

	switch scenario.Type {
	case domain.ScenarioThankAfterVisit:
		return "Спасибо за визит!", "Благодарим вас за посещение! Надеемся, вам понравилось. Ждём вас снова!"
	case domain.ScenarioRequestReview:
		return "Оставьте отзыв", "Расскажите о вашем визите — ваш отзыв поможет другим гостям!"
	case domain.ScenarioRemindRevisit30d:
		return "Мы скучаем!", "Прошло уже 30 дней с вашего последнего визита. Самое время вернуться!"
	case domain.ScenarioReactivateLost:
		return "Давно не виделись!", "Мы заметили, что вы давно не заходили. Приходите — у нас много нового!"
	case domain.ScenarioBirthdayGreeting:
		return "С днём рождения!", "Поздравляем с днём рождения! Желаем здоровья и хорошего настроения!"
	default:
		return "Уведомление", "Сообщение от бани"
	}
}

func (s *autoScenarioService) defaultTitle(t domain.AutoScenarioType) string {
	name, _, _ := domain.ScenarioMeta(t)
	return name
}
