package domain

import (
	"time"

	"github.com/google/uuid"
)

type AutoScenarioType string

const (
	ScenarioThankAfterVisit  AutoScenarioType = "thank_after_visit"
	ScenarioRequestReview    AutoScenarioType = "request_review"
	ScenarioRemindRevisit30d AutoScenarioType = "remind_revisit_30d"
	ScenarioReactivateLost   AutoScenarioType = "reactivate_lost_90d"
	ScenarioBirthdayGreeting AutoScenarioType = "birthday_greeting"
)

func AllScenarioTypes() []AutoScenarioType {
	return []AutoScenarioType{
		ScenarioThankAfterVisit,
		ScenarioRequestReview,
		ScenarioRemindRevisit30d,
		ScenarioReactivateLost,
		ScenarioBirthdayGreeting,
	}
}

func (t AutoScenarioType) IsValid() bool {
	for _, s := range AllScenarioTypes() {
		if s == t {
			return true
		}
	}
	return false
}

// ScenarioMeta returns display name, description, and default delay hours for a scenario type.
func ScenarioMeta(t AutoScenarioType) (name, description string, defaultDelayHours int) {
	switch t {
	case ScenarioThankAfterVisit:
		return "Благодарность после визита", "Автоматическое сообщение с благодарностью после завершения визита", 1
	case ScenarioRequestReview:
		return "Запрос отзыва", "Запрос оставить отзыв после визита", 2
	case ScenarioRemindRevisit30d:
		return "Напоминание о повторном визите", "Напоминание гостю вернуться через 30 дней", 720 // 30 * 24
	case ScenarioReactivateLost:
		return "Реактивация потерянных", "Сообщение гостям, не посещавшим более 90 дней", 2160 // 90 * 24
	case ScenarioBirthdayGreeting:
		return "Поздравление с днём рождения", "Автоматическое поздравление с днём рождения", 0
	default:
		return string(t), "", 0
	}
}

type AutoScenario struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Type        AutoScenarioType
	Enabled     bool
	CustomText  string
	Channel     BroadcastChannel
	DelayHours  int
	PromoCodeID *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *AutoScenario) Validate() error {
	if s.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	if !s.Type.IsValid() {
		return ErrInvalidInput
	}
	switch s.Channel {
	case BroadcastChannelPush, BroadcastChannelEmail, BroadcastChannelTelegram:
	default:
		return ErrInvalidInput
	}
	if s.DelayHours < 0 {
		return ErrInvalidInput
	}
	return nil
}
