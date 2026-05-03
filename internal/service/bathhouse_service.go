package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/rekurt/relax-hub/internal/antifraud"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/rekurt/relax-hub/internal/seo"
)

const (
	areaAvgPriceCachePrefix = "area_avg_price:"
	areaAvgPriceCacheTTL    = 1 * time.Hour
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
	fraudEngine    antifraud.FraudEngine
	stoplistRepo   repository.StoplistRepository
	userRepo       repository.UserRepository
	pdRepo         repository.PaymentDetailsRepository
	redisClient    *redis.Client
	logger         *logger.Logger
}

func NewBathhouseService(bhRepo repository.BathhouseRepository, bookingRepo repository.BookingRepository, photoRepo repository.BathhousePhotoRepository, subRepo repository.SubscriptionRepository, access *AccessChecker, kycSvc KYCService, offerSvc OfferService, paymentDetails PaymentDetailsService, auditSvc AuditLogService, fraudEngine antifraud.FraudEngine, stoplistRepo repository.StoplistRepository, userRepo repository.UserRepository, pdRepo repository.PaymentDetailsRepository, redisClient *redis.Client, log *logger.Logger) BathhouseService {
	return &bathhouseService{bhRepo: bhRepo, bookingRepo: bookingRepo, photoRepo: photoRepo, subRepo: subRepo, access: access, kycSvc: kycSvc, offerSvc: offerSvc, paymentDetails: paymentDetails, auditSvc: auditSvc, fraudEngine: fraudEngine, stoplistRepo: stoplistRepo, userRepo: userRepo, pdRepo: pdRepo, redisClient: redisClient, logger: log}
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

	// Antifraud: check stoplist and duplicate owner accounts
	if err := s.checkListingAntifraud(ctx, ownerID); err != nil {
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

func (s *bathhouseService) checkListingAntifraud(ctx context.Context, ownerID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return err
	}

	pd, _ := s.pdRepo.GetByUserID(ctx, ownerID)

	var inn, bankCard string
	if pd != nil {
		inn = pd.INN
		bankCard = pd.BankCardNumber
	}

	isBlocked, err := s.stoplistRepo.IsBlocked(ctx, user.Phone, user.Email, inn, bankCard)
	if err != nil {
		s.logger.Error("antifraud stoplist check failed", "error", err, "owner_id", ownerID)
		return err
	}

	dupCount, err := s.stoplistRepo.CountDuplicateOwners(ctx, user.Phone, user.Email, inn, ownerID)
	if err != nil {
		s.logger.Error("antifraud duplicate check failed", "error", err, "owner_id", ownerID)
		return err
	}

	return s.fraudEngine.CheckListingCreate(ctx, ownerID, isBlocked, dupCount)
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
