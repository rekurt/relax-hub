package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
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
	Name         *string
	Description  *string
	Address      *string
	CityID       *int64
	Latitude     *float64
	Longitude    *float64
	PricePerHour *int64
	MinDuration  *int
	MaxGuests    *int
	HasPool      *bool
	HasSauna     *bool
	HasSteamRoom *bool
	HasHotTub    *bool
	HasBBQ       *bool
	HasKaraoke   *bool
	Images       []string
	WorkingHours []domain.WorkingHours
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
	// Admin moderation:
	Approve(ctx context.Context, id uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
}

type bathhouseService struct {
	bhRepo         repository.BathhouseRepository
	bookingRepo    repository.BookingRepository
	access         *AccessChecker
	kycSvc         KYCService
	offerSvc       OfferService
	paymentDetails PaymentDetailsService
}

func NewBathhouseService(bhRepo repository.BathhouseRepository, bookingRepo repository.BookingRepository, access *AccessChecker, kycSvc KYCService, offerSvc OfferService, paymentDetails PaymentDetailsService) BathhouseService {
	return &bathhouseService{bhRepo: bhRepo, bookingRepo: bookingRepo, access: access, kycSvc: kycSvc, offerSvc: offerSvc, paymentDetails: paymentDetails}
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
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         input.Name,
		Slug:         slug,
		Description:  input.Description,
		Address:      input.Address,
		CityID:       input.CityID,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		PricePerHour: input.PricePerHour,
		MinDuration:  input.MinDuration,
		MaxGuests:    input.MaxGuests,
		HasPool:      input.HasPool,
		HasSauna:     input.HasSauna,
		HasSteamRoom: input.HasSteamRoom,
		HasHotTub:    input.HasHotTub,
		HasBBQ:       input.HasBBQ,
		HasKaraoke:   input.HasKaraoke,
		Images:       input.Images,
		WorkingHours: input.WorkingHours,
		Status:       domain.BathhouseStatusPending,
		ApiKey:       uuid.New().String(),
		CreatedAt:    now,
		UpdatedAt:    now,
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
	bh.UpdatedAt = time.Now()

	if err := bh.Validate(); err != nil {
		return nil, err
	}

	if err := s.bhRepo.Update(ctx, bh); err != nil {
		return nil, err
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

func (s *bathhouseService) Search(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return s.bhRepo.List(ctx, filter)
}

func (s *bathhouseService) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return s.bhRepo.ListByOwner(ctx, ownerID, page, pageSize)
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
