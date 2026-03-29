package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/seo"
)

type CreateBathhouseInput struct {
	Name         string
	Description  string
	Address      string
	CityID       int64
	Latitude     float64
	Longitude    float64
	PricePerHour int64
	MinDuration  int
	MaxGuests    int
	HasPool      bool
	HasSauna     bool
	HasSteamRoom bool
	HasHotTub    bool
	HasBBQ       bool
	HasKaraoke   bool
	Images       []string
	WorkingHours []domain.WorkingHours
}

type UpdateBathhouseInput struct {
	Name                       *string
	Description                *string
	Address                    *string
	CityID                     *int64
	Latitude                   *float64
	Longitude                  *float64
	PricePerHour               *int64
	MinDuration                *int
	MaxGuests                  *int
	HasPool                    *bool
	HasSauna                   *bool
	HasSteamRoom               *bool
	HasHotTub                  *bool
	HasBBQ                     *bool
	HasKaraoke                 *bool
	LongSessionThresholdHours  *int
	LongSessionDiscountPercent *int
	BaseCapacity               *int
	ExtraGuestSurcharge        *int64
	LastMinuteEnabled          *bool
	LastMinuteDiscountPercent  *int
	LastMinuteHoursThreshold   *int
	BufferMinutes              *int
	LeadTimeHours              *int
	MaxAdvanceDays             *int
	BookingMode                *string
	RequestTimeout             *int
	CancellationPolicy         *string
	SecurityDepositPercent     *int
	Images                     []string
	WorkingHours               []domain.WorkingHours
}

// CompletenessItem represents a single field check in the completeness result.
type CompletenessItem struct {
	Field    string `json:"field"`
	Label    string `json:"label"`
	Complete bool   `json:"complete"`
	Required bool   `json:"required"`
}

// CompletenessResult is the result of a listing completeness check.
type CompletenessResult struct {
	Score         int                `json:"score"`
	TotalRequired int                `json:"total_required"`
	DoneRequired  int                `json:"done_required"`
	TotalOptional int                `json:"total_optional"`
	DoneOptional  int                `json:"done_optional"`
	Ready         bool               `json:"ready"`
	Items         []CompletenessItem `json:"items"`
}

type BathhouseService interface {
	Create(ctx context.Context, ownerID uuid.UUID, input CreateBathhouseInput) (*domain.Bathhouse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Bathhouse, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*domain.Bathhouse, error)
	Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID, input UpdateBathhouseInput) (*domain.Bathhouse, error)
	Delete(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) error
	Search(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error)
	GetWidgetKey(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error)
	RegenerateWidgetKey(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error)
	CheckCompleteness(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (*CompletenessResult, error)
	SubmitForModeration(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error
	DuplicateBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (*domain.Bathhouse, error)
	DeactivateBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error
	ActivateBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error
	ArchiveBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	ComputeBadges(ctx context.Context, bh *domain.Bathhouse) []string
	IsLastMinuteActive(bh *domain.Bathhouse) bool
	GetAreaAvgPrice(ctx context.Context, cityID int64, lat, lng float64) (int64, error)
	// Admin moderation:
	Approve(ctx context.Context, id uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
}

type bathhouseService struct {
	bhRepo         repository.BathhouseRepository
	bookingRepo    repository.BookingRepository
	photoRepo      repository.BathhousePhotoRepository
	subRepo        repository.SubscriptionRepository
	access         *AccessChecker
	kycSvc         KYCService
	offerSvc       OfferService
	paymentDetails PaymentDetailsService
	auditSvc       AuditLogService
	logger         *logger.Logger
}

func NewBathhouseService(bhRepo repository.BathhouseRepository, bookingRepo repository.BookingRepository, photoRepo repository.BathhousePhotoRepository, subRepo repository.SubscriptionRepository, access *AccessChecker, kycSvc KYCService, offerSvc OfferService, paymentDetails PaymentDetailsService, auditSvc AuditLogService, log *logger.Logger) BathhouseService {
	return &bathhouseService{bhRepo: bhRepo, bookingRepo: bookingRepo, photoRepo: photoRepo, subRepo: subRepo, access: access, kycSvc: kycSvc, offerSvc: offerSvc, paymentDetails: paymentDetails, auditSvc: auditSvc, logger: log}
}

func (s *bathhouseService) Create(ctx context.Context, ownerID uuid.UUID, input CreateBathhouseInput) (*domain.Bathhouse, error) {
	// Onboarding gate: KYC must be approved
	approved, err := s.kycSvc.IsApproved(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !approved {
		return nil, domain.ErrKYCNotApproved
	}

	// Onboarding gate: offer must be accepted
	accepted, err := s.offerSvc.IsAccepted(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return nil, domain.ErrOfferNotAccepted
	}

	// Onboarding gate: payment details must be set and valid
	if err := s.paymentDetails.Validate(ctx, ownerID); err != nil {
		return nil, err
	}

	now := time.Now()

	slug, err := seo.GenerateUniqueSlug(input.Name, func(slug string) (bool, error) {
		return s.bhRepo.SlugExists(ctx, slug)
	})
	if err != nil {
		return nil, err
	}

	bh := &domain.Bathhouse{
		ID:                         uuid.New(),
		OwnerID:                    ownerID,
		Name:                       input.Name,
		Slug:                       slug,
		Description:                input.Description,
		Address:                    input.Address,
		CityID:                     input.CityID,
		Latitude:                   input.Latitude,
		Longitude:                  input.Longitude,
		PricePerHour:               input.PricePerHour,
		MinDuration:                input.MinDuration,
		MaxGuests:                  input.MaxGuests,
		HasPool:                    input.HasPool,
		HasSauna:                   input.HasSauna,
		HasSteamRoom:               input.HasSteamRoom,
		HasHotTub:                  input.HasHotTub,
		HasBBQ:                     input.HasBBQ,
		HasKaraoke:                 input.HasKaraoke,
		Images:                     input.Images,
		WorkingHours:               input.WorkingHours,
		LongSessionThresholdHours:  4,
		LongSessionDiscountPercent: 0,
		BaseCapacity:               input.MaxGuests,
		ExtraGuestSurcharge:        0,
		Status:                     domain.BathhouseStatusPending,
		ApiKey:                     uuid.New().String(),
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}

	if err := bh.Validate(); err != nil {
		return nil, err
	}

	if err := s.bhRepo.Create(ctx, bh); err != nil {
		return nil, err
	}

	return bh, nil
}

func (s *bathhouseService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	bh, err := s.bhRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Public endpoint: only return active bathhouses
	if bh.Status != domain.BathhouseStatusActive {
		return nil, domain.ErrNotFound
	}

	return bh, nil
}

func (s *bathhouseService) GetBySlug(ctx context.Context, slug string) (*domain.Bathhouse, error) {
	bh, err := s.bhRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if bh.Status != domain.BathhouseStatusActive {
		return nil, domain.ErrNotFound
	}

	return bh, nil
}

func (s *bathhouseService) Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID, input UpdateBathhouseInput) (*domain.Bathhouse, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, role, id); err != nil {
		return nil, err
	}

	bh, err := s.bhRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Capture old state for audit logging
	oldBh := *bh

	if input.Name != nil {
		bh.Name = *input.Name
		newSlug, err := seo.GenerateUniqueSlug(*input.Name, func(slug string) (bool, error) {
			if slug == bh.Slug {
				return false, nil
			}
			return s.bhRepo.SlugExists(ctx, slug)
		})
		if err != nil {
			return nil, err
		}
		bh.Slug = newSlug
	}
	if input.Description != nil {
		bh.Description = *input.Description
	}
	if input.Address != nil {
		bh.Address = *input.Address
	}
	if input.CityID != nil {
		bh.CityID = *input.CityID
	}
	if input.Latitude != nil {
		bh.Latitude = *input.Latitude
	}
	if input.Longitude != nil {
		bh.Longitude = *input.Longitude
	}
	if input.PricePerHour != nil {
		bh.PricePerHour = *input.PricePerHour
	}
	if input.MinDuration != nil {
		bh.MinDuration = *input.MinDuration
	}
	if input.MaxGuests != nil {
		bh.MaxGuests = *input.MaxGuests
	}
	if input.HasPool != nil {
		bh.HasPool = *input.HasPool
	}
	if input.HasSauna != nil {
		bh.HasSauna = *input.HasSauna
	}
	if input.HasSteamRoom != nil {
		bh.HasSteamRoom = *input.HasSteamRoom
	}
	if input.HasHotTub != nil {
		bh.HasHotTub = *input.HasHotTub
	}
	if input.HasBBQ != nil {
		bh.HasBBQ = *input.HasBBQ
	}
	if input.HasKaraoke != nil {
		bh.HasKaraoke = *input.HasKaraoke
	}
	if input.Images != nil {
		bh.Images = input.Images
	}
	if input.WorkingHours != nil {
		bh.WorkingHours = input.WorkingHours
	}
	if input.LongSessionThresholdHours != nil {
		bh.LongSessionThresholdHours = *input.LongSessionThresholdHours
	}
	if input.LongSessionDiscountPercent != nil {
		bh.LongSessionDiscountPercent = *input.LongSessionDiscountPercent
	}
	if input.BaseCapacity != nil {
		bh.BaseCapacity = *input.BaseCapacity
	}
	if input.ExtraGuestSurcharge != nil {
		bh.ExtraGuestSurcharge = *input.ExtraGuestSurcharge
	}
	if input.LastMinuteEnabled != nil {
		bh.LastMinuteEnabled = *input.LastMinuteEnabled
	}
	if input.LastMinuteDiscountPercent != nil {
		bh.LastMinuteDiscountPercent = *input.LastMinuteDiscountPercent
	}
	if input.LastMinuteHoursThreshold != nil {
		bh.LastMinuteHoursThreshold = *input.LastMinuteHoursThreshold
	}
	if input.BufferMinutes != nil {
		bh.BufferMinutes = *input.BufferMinutes
	}
	if input.LeadTimeHours != nil {
		bh.LeadTimeHours = *input.LeadTimeHours
	}
	if input.MaxAdvanceDays != nil {
		bh.MaxAdvanceDays = *input.MaxAdvanceDays
	}
	if input.BookingMode != nil {
		bh.BookingMode = *input.BookingMode
	}
	if input.RequestTimeout != nil {
		bh.RequestTimeout = *input.RequestTimeout
	}
	if input.CancellationPolicy != nil {
		bh.CancellationPolicy = domain.CancellationPolicy(*input.CancellationPolicy)
	}
	if input.SecurityDepositPercent != nil {
		bh.SecurityDepositPercent = *input.SecurityDepositPercent
	}
	// Detect substantial changes before persisting, so status update is atomic with data update
	changedFields := buildChangedFields(&oldBh, bh)
	if len(changedFields) > 0 && bh.Status != domain.BathhouseStatusPending && s.auditSvc.IsSubstantialChange(&oldBh, bh) {
		bh.Status = domain.BathhouseStatusPending
	}

	if err := bh.Validate(); err != nil {
		return nil, err
	}

	if err := s.bhRepo.Update(ctx, bh); err != nil {
		return nil, err
	}

	// Log the change in audit log
	if len(changedFields) > 0 {
		if err := s.auditSvc.LogChange(ctx, "bathhouse", id, userID, domain.AuditActionUpdate, changedFields); err != nil {
			s.logger.Error("failed to log audit change", "bathhouse_id", id, "error", err)
		}
	}

	return bh, nil
}

func (s *bathhouseService) Delete(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) error {
	bh, err := s.bhRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if bh.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	// Check for active bookings (pending or confirmed)
	activeCount, err := s.bookingRepo.CountActiveByBathhouse(ctx, id)
	if err != nil {
		return err
	}
	if activeCount > 0 {
		return domain.ErrBathhouseHasBookings
	}

	return s.bhRepo.Delete(ctx, id)
}

const maxPromotedPerPage = 3

func (s *bathhouseService) Search(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	result, err := s.bhRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Enforce max promoted listings per page: cap at maxPromotedPerPage
	promotedCount := 0
	for i := range result.Items {
		if result.Items[i].IsPromoted {
			promotedCount++
			if promotedCount > maxPromotedPerPage {
				result.Items[i].IsPromoted = false
			}
		}
	}

	return result, nil
}

func (s *bathhouseService) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return s.bhRepo.ListByOwner(ctx, ownerID, page, pageSize)
}

func (s *bathhouseService) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return s.bhRepo.IncrementViewCount(ctx, id)
}

func (s *bathhouseService) Approve(ctx context.Context, id uuid.UUID) error {
	bh, err := s.bhRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if bh.Status != domain.BathhouseStatusPending {
		return domain.ErrInvalidInput
	}

	return s.bhRepo.UpdateStatus(ctx, id, domain.BathhouseStatusActive)
}

func (s *bathhouseService) Reject(ctx context.Context, id uuid.UUID) error {
	bh, err := s.bhRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if bh.Status != domain.BathhouseStatusPending {
		return domain.ErrInvalidInput
	}

	return s.bhRepo.UpdateStatus(ctx, id, domain.BathhouseStatusRejected)
}

func (s *bathhouseService) GetWidgetKey(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error) {
	// Use AccessChecker for proper authorization (supports owner and representative)
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return "", err
	}

	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return "", err
	}

	return bh.ApiKey, nil
}

func (s *bathhouseService) RegenerateWidgetKey(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error) {
	// Use AccessChecker for proper authorization (supports owner and representative)
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return "", err
	}

	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return "", err
	}

	bh.ApiKey = uuid.New().String()
	if err := s.bhRepo.Update(ctx, bh); err != nil {
		return "", err
	}

	return bh.ApiKey, nil
}

func (s *bathhouseService) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Bathhouse, error) {
	return s.bhRepo.GetByAPIKey(ctx, apiKey)
}

func (s *bathhouseService) CheckCompleteness(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (*CompletenessResult, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}

	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return nil, err
	}

	var verifiedPhotoCount int
	if s.photoRepo != nil {
		photos, err := s.photoRepo.ListVerifiedByBathhouse(ctx, bathhouseID)
		if err == nil {
			verifiedPhotoCount = len(photos)
		}
	}

	amenityCount := countAmenities(bh)

	items := []CompletenessItem{
		{Field: "name", Label: "Название", Complete: bh.Name != "", Required: true},
		{Field: "description", Label: "Описание (50+ символов)", Complete: len([]rune(bh.Description)) >= 50, Required: true},
		{Field: "address", Label: "Адрес", Complete: bh.Address != "", Required: true},
		{Field: "city", Label: "Город", Complete: bh.CityID > 0, Required: true},
		{Field: "coordinates", Label: "Координаты", Complete: bh.Latitude != 0 && bh.Longitude != 0, Required: true},
		{Field: "photos", Label: "Фотографии (3+)", Complete: verifiedPhotoCount >= 3, Required: true},
		{Field: "price", Label: "Цена за час", Complete: bh.PricePerHour > 0, Required: true},
		{Field: "schedule", Label: "Расписание работы", Complete: len(bh.WorkingHours) > 0, Required: true},
		{Field: "capacity", Label: "Вместимость", Complete: bh.MaxGuests > 0, Required: true},
		{Field: "amenities", Label: "Удобства (3+)", Complete: amenityCount >= 3, Required: false},
	}

	var totalReq, doneReq, totalOpt, doneOpt int
	for _, item := range items {
		if item.Required {
			totalReq++
			if item.Complete {
				doneReq++
			}
		} else {
			totalOpt++
			if item.Complete {
				doneOpt++
			}
		}
	}

	total := totalReq + totalOpt
	done := doneReq + doneOpt
	score := 0
	if total > 0 {
		score = done * 100 / total
	}

	return &CompletenessResult{
		Score:         score,
		TotalRequired: totalReq,
		DoneRequired:  doneReq,
		TotalOptional: totalOpt,
		DoneOptional:  doneOpt,
		Ready:         doneReq == totalReq,
		Items:         items,
	}, nil
}

func (s *bathhouseService) SubmitForModeration(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error {
	result, err := s.CheckCompleteness(ctx, userID, userRole, bathhouseID)
	if err != nil {
		return err
	}
	if !result.Ready {
		return domain.ErrListingIncomplete
	}

	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return err
	}

	if bh.Status == domain.BathhouseStatusPending {
		return nil // already pending
	}

	if bh.Status == domain.BathhouseStatusArchived {
		return domain.ErrInvalidInput
	}

	return s.bhRepo.UpdateStatus(ctx, bathhouseID, domain.BathhouseStatusPending)
}

func (s *bathhouseService) DuplicateBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (*domain.Bathhouse, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}

	original, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return nil, err
	}

	copyName := original.Name + " (копия)"
	slug, err := seo.GenerateUniqueSlug(copyName, func(slug string) (bool, error) {
		return s.bhRepo.SlugExists(ctx, slug)
	})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	dup := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      original.OwnerID,
		Name:         copyName,
		Slug:         slug,
		Description:  original.Description,
		Address:      original.Address,
		CityID:       original.CityID,
		Latitude:     original.Latitude,
		Longitude:    original.Longitude,
		PricePerHour: original.PricePerHour,
		MinDuration:  original.MinDuration,
		MaxGuests:    original.MaxGuests,
		HasPool:      original.HasPool,
		HasSauna:     original.HasSauna,
		HasSteamRoom: original.HasSteamRoom,
		HasHotTub:    original.HasHotTub,
		HasBBQ:       original.HasBBQ,
		HasKaraoke:   original.HasKaraoke,
		Images:       original.Images,
		WorkingHours: original.WorkingHours,
		Status:       domain.BathhouseStatusInactive,
		ApiKey:       uuid.New().String(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.bhRepo.Create(ctx, dup); err != nil {
		return nil, err
	}

	return dup, nil
}

func (s *bathhouseService) DeactivateBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return err
	}

	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return err
	}

	if bh.Status != domain.BathhouseStatusActive {
		return domain.ErrInvalidInput
	}

	if err := s.bhRepo.UpdateStatus(ctx, bathhouseID, domain.BathhouseStatusInactive); err != nil {
		return err
	}

	if err := s.auditSvc.LogChange(ctx, "bathhouse", bathhouseID, userID, domain.AuditActionStatusChange,
		map[string]interface{}{"status": map[string]string{"old": string(bh.Status), "new": string(domain.BathhouseStatusInactive)}}); err != nil {
		s.logger.Error("failed to log audit change", "bathhouse_id", bathhouseID, "error", err)
	}

	return nil
}

func (s *bathhouseService) ActivateBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return err
	}

	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return err
	}

	if bh.Status != domain.BathhouseStatusInactive {
		return domain.ErrInvalidInput
	}

	if err := s.bhRepo.UpdateStatus(ctx, bathhouseID, domain.BathhouseStatusActive); err != nil {
		return err
	}

	if err := s.auditSvc.LogChange(ctx, "bathhouse", bathhouseID, userID, domain.AuditActionStatusChange,
		map[string]interface{}{"status": map[string]string{"old": string(bh.Status), "new": string(domain.BathhouseStatusActive)}}); err != nil {
		s.logger.Error("failed to log audit change", "bathhouse_id", bathhouseID, "error", err)
	}

	return nil
}

func (s *bathhouseService) ArchiveBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return err
	}

	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return err
	}

	if bh.Status == domain.BathhouseStatusArchived {
		return domain.ErrInvalidInput
	}

	// Cannot archive if there are active bookings
	activeCount, err := s.bookingRepo.CountActiveByBathhouse(ctx, bathhouseID)
	if err != nil {
		return err
	}
	if activeCount > 0 {
		return domain.ErrBathhouseHasBookings
	}

	if err := s.bhRepo.UpdateStatus(ctx, bathhouseID, domain.BathhouseStatusArchived); err != nil {
		return err
	}

	if err := s.auditSvc.LogChange(ctx, "bathhouse", bathhouseID, userID, domain.AuditActionStatusChange,
		map[string]interface{}{"status": map[string]string{"old": string(bh.Status), "new": string(domain.BathhouseStatusArchived)}}); err != nil {
		s.logger.Error("failed to log audit change", "bathhouse_id", bathhouseID, "error", err)
	}

	return nil
}

func buildChangedFields(old, new *domain.Bathhouse) map[string]interface{} {
	changes := make(map[string]interface{})

	if old.Name != new.Name {
		changes["name"] = map[string]string{"old": old.Name, "new": new.Name}
	}
	if old.Description != new.Description {
		changes["description"] = map[string]string{"old": old.Description, "new": new.Description}
	}
	if old.Address != new.Address {
		changes["address"] = map[string]string{"old": old.Address, "new": new.Address}
	}
	if old.CityID != new.CityID {
		changes["city_id"] = map[string]int64{"old": old.CityID, "new": new.CityID}
	}
	if old.Latitude != new.Latitude {
		changes["latitude"] = map[string]float64{"old": old.Latitude, "new": new.Latitude}
	}
	if old.Longitude != new.Longitude {
		changes["longitude"] = map[string]float64{"old": old.Longitude, "new": new.Longitude}
	}
	if old.PricePerHour != new.PricePerHour {
		changes["price_per_hour"] = map[string]int64{"old": old.PricePerHour, "new": new.PricePerHour}
	}
	if old.MinDuration != new.MinDuration {
		changes["min_duration"] = map[string]int{"old": old.MinDuration, "new": new.MinDuration}
	}
	if old.MaxGuests != new.MaxGuests {
		changes["max_guests"] = map[string]int{"old": old.MaxGuests, "new": new.MaxGuests}
	}
	if old.HasPool != new.HasPool {
		changes["has_pool"] = map[string]bool{"old": old.HasPool, "new": new.HasPool}
	}
	if old.HasSauna != new.HasSauna {
		changes["has_sauna"] = map[string]bool{"old": old.HasSauna, "new": new.HasSauna}
	}
	if old.HasSteamRoom != new.HasSteamRoom {
		changes["has_steam_room"] = map[string]bool{"old": old.HasSteamRoom, "new": new.HasSteamRoom}
	}
	if old.HasHotTub != new.HasHotTub {
		changes["has_hot_tub"] = map[string]bool{"old": old.HasHotTub, "new": new.HasHotTub}
	}
	if old.HasBBQ != new.HasBBQ {
		changes["has_bbq"] = map[string]bool{"old": old.HasBBQ, "new": new.HasBBQ}
	}
	if old.HasKaraoke != new.HasKaraoke {
		changes["has_karaoke"] = map[string]bool{"old": old.HasKaraoke, "new": new.HasKaraoke}
	}
	if !stringSlicesEqual(old.Images, new.Images) {
		changes["images"] = map[string]interface{}{"old": old.Images, "new": new.Images}
	}
	if !workingHoursEqual(old.WorkingHours, new.WorkingHours) {
		oldWH, _ := json.Marshal(old.WorkingHours)
		newWH, _ := json.Marshal(new.WorkingHours)
		changes["working_hours"] = map[string]string{"old": string(oldWH), "new": string(newWH)}
	}

	return changes
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func workingHoursEqual(a, b []domain.WorkingHours) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

const (
	badgeVerified   = "verified"
	badgeTop        = "top"
	badgePremium    = "premium"
	badgeNew        = "new"
	badgeLastMinute = "last_minute"

	badgeTopMinRating  = 4.5
	badgeTopMinReviews = 10
	badgeNewMaxDays    = 30
	badgeNewMaxReviews = 3
)

func (s *bathhouseService) ComputeBadges(ctx context.Context, bh *domain.Bathhouse) []string {
	var badges []string

	// "Verified": photos moderated + KYC approved
	if bh.IsPhotoVerified {
		if approved, err := s.kycSvc.IsApproved(ctx, bh.OwnerID); err == nil && approved {
			badges = append(badges, badgeVerified)
		}
	}

	// "Top": Bayesian >= 4.5 AND review count >= 10
	if bh.BayesianRating >= badgeTopMinRating && bh.ReviewCount >= badgeTopMinReviews {
		badges = append(badges, badgeTop)
	}

	// "Premium": active subscription (premium or promoted plan)
	if s.subRepo != nil {
		sub, err := s.subRepo.GetActiveBybathhouse(ctx, bh.ID)
		if err == nil && sub != nil && (sub.Plan == domain.PlanPremium || sub.Plan == domain.PlanPromoted) {
			badges = append(badges, badgePremium)
		}
	}

	// "New": < 30 days old AND < 3 reviews
	if time.Since(bh.CreatedAt) < badgeNewMaxDays*24*time.Hour && bh.ReviewCount < badgeNewMaxReviews {
		badges = append(badges, badgeNew)
	}

	// "LastMinute": last-minute discount is currently active
	if s.IsLastMinuteActive(bh) {
		badges = append(badges, badgeLastMinute)
	}

	return badges
}

// IsLastMinuteActive checks whether a bathhouse currently qualifies for its last-minute discount.
// Returns true when the bathhouse has last-minute enabled and there are open working hours
// starting within the threshold window from now.
func (s *bathhouseService) IsLastMinuteActive(bh *domain.Bathhouse) bool {
	if !bh.LastMinuteEnabled || bh.LastMinuteHoursThreshold <= 0 {
		return false
	}

	now := time.Now()
	threshold := now.Add(time.Duration(bh.LastMinuteHoursThreshold) * time.Hour)

	// Check if any working hours slot today or tomorrow starts within the threshold window
	for dayOffset := 0; dayOffset <= 1; dayOffset++ {
		checkDate := now.AddDate(0, 0, dayOffset)
		wd := checkDate.Weekday()
		dayOfWeek := int(wd) - 1
		if wd == time.Sunday {
			dayOfWeek = 6
		}

		for _, wh := range bh.WorkingHours {
			if wh.DayOfWeek != dayOfWeek {
				continue
			}
			// Parse open time on that day
			if len(wh.OpenTime) != 5 {
				continue
			}
			h := (int(wh.OpenTime[0]-'0') * 10) + int(wh.OpenTime[1]-'0')
			m := (int(wh.OpenTime[3]-'0') * 10) + int(wh.OpenTime[4]-'0')
			slotStart := time.Date(checkDate.Year(), checkDate.Month(), checkDate.Day(), h, m, 0, 0, now.Location())

			// Slot must be in the future and within threshold
			if slotStart.After(now) && !slotStart.After(threshold) {
				return true
			}
		}
	}
	return false
}

// GetAreaAvgPrice returns the average price per hour for active bathhouses in the same city
// within a 5km radius of the given coordinates. Returns 0 if no data available.
func (s *bathhouseService) GetAreaAvgPrice(ctx context.Context, cityID int64, lat, lng float64) (int64, error) {
	return s.bhRepo.GetAreaAvgPrice(ctx, cityID, lat, lng)
}

func countAmenities(bh *domain.Bathhouse) int {
	count := 0
	if bh.HasPool {
		count++
	}
	if bh.HasSauna {
		count++
	}
	if bh.HasSteamRoom {
		count++
	}
	if bh.HasHotTub {
		count++
	}
	if bh.HasBBQ {
		count++
	}
	if bh.HasKaraoke {
		count++
	}
	return count
}
