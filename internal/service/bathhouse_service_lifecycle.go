package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/seo"
)

func (s *bathhouseService) Approve(ctx context.Context, id uuid.UUID) error {
	bh, err := s.bhRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if bh.Status != domain.BathhouseStatusPending {
		return fmt.Errorf("only pending bathhouses can be approved: %w", domain.ErrInvalidInput)
	}

	return s.bhRepo.UpdateStatus(ctx, id, domain.BathhouseStatusActive)
}

func (s *bathhouseService) Reject(ctx context.Context, id uuid.UUID) error {
	bh, err := s.bhRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if bh.Status != domain.BathhouseStatusPending {
		return fmt.Errorf("only pending bathhouses can be rejected: %w", domain.ErrInvalidInput)
	}

	return s.bhRepo.UpdateStatus(ctx, id, domain.BathhouseStatusRejected)
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

func (s *bathhouseService) GetWidgetKey(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error) {
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
