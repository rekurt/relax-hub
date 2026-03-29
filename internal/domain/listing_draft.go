package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ListingDraftStatus - статус черновика объявления
type ListingDraftStatus string

const (
	ListingDraftStatusDraft     ListingDraftStatus = "draft"
	ListingDraftStatusSubmitted ListingDraftStatus = "submitted"
)

// ListingDraft - черновик объявления бани (7-шаговый визард)
type ListingDraft struct {
	ID          uuid.UUID               `json:"id"`
	UserID      uuid.UUID               `json:"user_id"`
	Status      ListingDraftStatus      `json:"status"`
	CurrentStep int                     `json:"current_step"`
	StepData    map[int]json.RawMessage `json:"step_data"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

const (
	ListingStepBasicInfo       = 1 // name, type, description
	ListingStepLocation        = 2 // address, city, coordinates
	ListingStepAmenities       = 3 // amenities, capacity
	ListingStepPhotos          = 4 // photos (min 3)
	ListingStepPricing         = 5 // pricing & schedule
	ListingStepBookingSettings = 6 // booking mode, buffer, lead time
	ListingStepReview          = 7 // review & submit
	ListingTotalSteps          = 7
)

// IsStepValid проверяет валидность номера шага
func IsStepValid(step int) bool {
	return step >= ListingStepBasicInfo && step <= ListingTotalSteps
}

// AllStepsComplete проверяет что все шаги (1-6) заполнены
// Шаг 7 — это просто обзор, данные не требуются
func (d *ListingDraft) AllStepsComplete() bool {
	for step := ListingStepBasicInfo; step <= ListingStepBookingSettings; step++ {
		if _, ok := d.StepData[step]; !ok {
			return false
		}
	}
	return true
}
