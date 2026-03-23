package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type bathhouseTestEnv struct {
	svc         service.BathhouseService
	bhRepo      *mock.BathhouseRepo
	repRepo     *mock.RepresentativeRepo
	bookingRepo *mock.BookingRepo
	photoRepo   *mock.BathhousePhotoRepo
	kycRepo     *mock.KYCRepo
	offerRepo   *mock.OfferRepo
	pdRepo      *mock.PaymentDetailsRepo
}

func newBathhouseTestEnv() *bathhouseTestEnv {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	bookingRepo := mock.NewBookingRepo()
	photoRepo := mock.NewBathhousePhotoRepo()
	kycRepo := mock.NewKYCRepo()
	offerRepo := mock.NewOfferRepo()
	pdRepo := mock.NewPaymentDetailsRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	kycSvc := service.NewKYCService(kycRepo, &noopNotifService{}, log)
	offerSvc := service.NewOfferService(offerRepo, log)
	pdSvc := service.NewPaymentDetailsService(pdRepo, log)
	auditLogRepo := mock.NewAuditLogRepo()
	auditSvc := service.NewAuditLogService(auditLogRepo, log)
	svc := service.NewBathhouseService(bhRepo, bookingRepo, photoRepo, access, kycSvc, offerSvc, pdSvc, auditSvc)
	return &bathhouseTestEnv{
		svc: svc, bhRepo: bhRepo, repRepo: repRepo, bookingRepo: bookingRepo,
		photoRepo: photoRepo, kycRepo: kycRepo, offerRepo: offerRepo, pdRepo: pdRepo,
	}
}

// setupOnboardingGate prepares KYC, offer, and payment details for a user so they pass the gate.
func (e *bathhouseTestEnv) setupOnboardingGate(t *testing.T, userID uuid.UUID) {
	t.Helper()

	// Approved KYC
	expiresAt := time.Now().Add(365 * 24 * time.Hour)
	kyc := &domain.KYCApplication{
		ID:       uuid.New(),
		UserID:   userID,
		Status:   domain.KYCStatusApproved,
		FullName: "Test User",
		INN:      "123456789012",
		ExpiresAt: &expiresAt,
	}
	if err := e.kycRepo.Create(context.Background(), kyc); err != nil {
		t.Fatalf("setup KYC: %v", err)
	}

	// Accepted offer (version 1.0)
	offer := &domain.OfferAcceptance{
		ID:           uuid.New(),
		UserID:       userID,
		OfferVersion: "1.0",
	}
	if err := e.offerRepo.Create(context.Background(), offer); err != nil {
		t.Fatalf("setup offer: %v", err)
	}

	// Payment details
	pd := &domain.PaymentDetails{
		ID:             uuid.New(),
		UserID:         userID,
		EntityType:     domain.KYCEntityIndividual,
		BankCardNumber: "4111111111111111",
		CardHolderName: "TEST USER",
	}
	if err := e.pdRepo.Upsert(context.Background(), pd); err != nil {
		t.Fatalf("setup payment details: %v", err)
	}
}

func newBathhouseService() (service.BathhouseService, *mock.BathhouseRepo, *mock.RepresentativeRepo, *mock.BookingRepo) {
	env := newBathhouseTestEnv()
	return env.svc, env.bhRepo, env.repRepo, env.bookingRepo
}

func TestBathhouseService_Create_PendingStatus(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	env.setupOnboardingGate(t, ownerID)

	bh, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name:         "My Bathhouse",
		Address:      "123 Street",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bh.Status != domain.BathhouseStatusPending {
		t.Errorf("status = %q, want %q", bh.Status, domain.BathhouseStatusPending)
	}
	if bh.OwnerID != ownerID {
		t.Errorf("ownerID = %v, want %v", bh.OwnerID, ownerID)
	}
}

func TestBathhouseService_Create_InvalidInput(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	env.setupOnboardingGate(t, ownerID)

	_, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name: "", // empty name
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should return ErrInvalidInput, got: %v", err)
	}
}

func TestBathhouseService_Update_OwnerAllowed(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	newName := "Updated Name"
	updated, err := svc.Update(context.Background(), ownerID, domain.RoleOwner, bh.ID, service.UpdateBathhouseInput{
		Name: &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("name = %q, want %q", updated.Name, newName)
	}
}

func TestBathhouseService_Update_RepresentativeAllowed(t *testing.T) {
	svc, bhRepo, repRepo, _ := newBathhouseService()
	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	rep := &domain.Representative{
		ID: uuid.New(), UserID: repUserID, BathhouseID: bh.ID, OwnerID: ownerID,
	}
	_ = repRepo.Create(context.Background(), rep)

	newName := "Updated by Rep"
	_, err := svc.Update(context.Background(), repUserID, domain.RoleRepresentative, bh.ID, service.UpdateBathhouseInput{
		Name: &newName,
	})

	if err != nil {
		t.Errorf("representative should be allowed to update assigned bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Update_ClientForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	newName := "Hacked"
	_, err := svc.Update(context.Background(), clientID, domain.RoleClient, bh.ID, service.UpdateBathhouseInput{
		Name: &newName,
	})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("client should be forbidden from updating bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Delete_OwnerOnly(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := svc.Delete(context.Background(), ownerID, bh.ID)
	if err != nil {
		t.Fatalf("owner should be able to delete own bathhouse, got: %v", err)
	}

	_, err = svc.GetByID(context.Background(), bh.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("bathhouse should be deleted, got: %v", err)
	}
}

func TestBathhouseService_Delete_OtherOwnerForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := svc.Delete(context.Background(), otherOwnerID, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("other owner should be forbidden from deleting bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Approve_AdminOnly(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Pending Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusPending,
	}
	_ = bhRepo.Create(context.Background(), bh)

	err := svc.Approve(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("approve should work: %v", err)
	}

	updated, _ := svc.GetByID(context.Background(), bh.ID)
	if updated.Status != domain.BathhouseStatusActive {
		t.Errorf("status = %q, want %q", updated.Status, domain.BathhouseStatusActive)
	}
}

func TestBathhouseService_Approve_NotPendingFails(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	bh := createBathhouse(t, bhRepo, uuid.New()) // active status

	err := svc.Approve(context.Background(), bh.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for non-pending bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Reject(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Pending Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusPending,
	}
	_ = bhRepo.Create(context.Background(), bh)

	err := svc.Reject(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("reject should work: %v", err)
	}

	// After rejection, public GetByID should not find the bathhouse
	_, err = svc.GetByID(context.Background(), bh.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("rejected bathhouse should not be visible via GetByID, got: %v", err)
	}

	// Verify the status was actually changed in the repo
	updated, _ := bhRepo.GetByID(context.Background(), bh.ID)
	if updated.Status != domain.BathhouseStatusRejected {
		t.Errorf("status = %q, want %q", updated.Status, domain.BathhouseStatusRejected)
	}
}

func TestBathhouseService_GetWidgetKey_Success(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	env.setupOnboardingGate(t, ownerID)

	bh, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name:         "Test Bath",
		Address:      "123 St",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	apiKey, err := env.svc.GetWidgetKey(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey == "" {
		t.Errorf("expected non-empty API key")
	}
	if apiKey != bh.ApiKey {
		t.Errorf("apiKey = %q, want %q", apiKey, bh.ApiKey)
	}
}

func TestBathhouseService_GetWidgetKey_OtherOwnerForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.GetWidgetKey(context.Background(), otherOwnerID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("other owner should be forbidden, got: %v", err)
	}
}

func TestBathhouseService_RegenerateWidgetKey_Success(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)
	originalKey := bh.ApiKey

	newKey, err := svc.RegenerateWidgetKey(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newKey == "" {
		t.Errorf("expected non-empty API key")
	}
	if newKey == originalKey {
		t.Errorf("new key should be different from old key")
	}

	// Verify the key was updated in the repository
	updated, _ := bhRepo.GetByID(context.Background(), bh.ID)
	if updated.ApiKey != newKey {
		t.Errorf("stored apiKey = %q, want %q", updated.ApiKey, newKey)
	}
}

func TestBathhouseService_RegenerateWidgetKey_OtherOwnerForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.RegenerateWidgetKey(context.Background(), otherOwnerID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("other owner should be forbidden, got: %v", err)
	}
}

func TestBathhouseService_Create_GeneratesSlug(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	env.setupOnboardingGate(t, ownerID)

	bh, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name:         "Баня на Липовой",
		Address:      "ул. Липовая 5",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bh.Slug == "" {
		t.Error("slug should not be empty")
	}
	if bh.Slug != "banya-na-lipovoy" {
		t.Errorf("slug = %q, want %q", bh.Slug, "banya-na-lipovoy")
	}
}

func TestBathhouseService_Create_UniqueSlug(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	env.setupOnboardingGate(t, ownerID)

	bh1, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name:         "Баня",
		Address:      "ул. А",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bh2, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name:         "Баня",
		Address:      "ул. Б",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bh1.Slug == bh2.Slug {
		t.Errorf("slugs should be different: %q vs %q", bh1.Slug, bh2.Slug)
	}
	if bh2.Slug != "banya-2" {
		t.Errorf("second slug = %q, want %q", bh2.Slug, "banya-2")
	}
}

func TestBathhouseService_Update_RegeneratesSlugOnNameChange(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)
	oldSlug := bh.Slug

	newName := "Новое Название"
	updated, err := svc.Update(context.Background(), ownerID, domain.RoleOwner, bh.ID, service.UpdateBathhouseInput{
		Name: &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Slug == oldSlug {
		t.Error("slug should change when name changes")
	}
	if updated.Slug != "novoe-nazvanie" {
		t.Errorf("slug = %q, want %q", updated.Slug, "novoe-nazvanie")
	}
}

func TestBathhouseService_GetBySlug_ActiveOnly(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Active bathhouse should be found
	found, err := svc.GetBySlug(context.Background(), bh.Slug)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != bh.ID {
		t.Errorf("found wrong bathhouse: %v", found.ID)
	}

	// Inactive bathhouse should not be found
	_ = bhRepo.UpdateStatus(context.Background(), bh.ID, domain.BathhouseStatusPending)
	_, err = svc.GetBySlug(context.Background(), bh.Slug)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("pending bathhouse should not be found via GetBySlug, got: %v", err)
	}
}

func TestBathhouseService_GetBySlug_NotFound(t *testing.T) {
	svc, _, _, _ := newBathhouseService()

	_, err := svc.GetBySlug(context.Background(), "nonexistent-slug")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("should return ErrNotFound for nonexistent slug, got: %v", err)
	}
}

func TestBathhouseService_Create_Gate_NoKYC(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	_, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name: "Test", Address: "123 St", CityID: 1, PricePerHour: 5000, MinDuration: 1, MaxGuests: 10,
	})
	if !errors.Is(err, domain.ErrKYCNotApproved) {
		t.Errorf("expected KYC not approved error without KYC, got: %v", err)
	}
}

func TestBathhouseService_Create_Gate_KYCNotApproved(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	kyc := &domain.KYCApplication{
		ID: uuid.New(), UserID: ownerID, Status: domain.KYCStatusPending,
		FullName: "Test", INN: "123456789012",
	}
	_ = env.kycRepo.Create(context.Background(), kyc)

	_, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name: "Test", Address: "123 St", CityID: 1, PricePerHour: 5000, MinDuration: 1, MaxGuests: 10,
	})
	if !errors.Is(err, domain.ErrKYCNotApproved) {
		t.Errorf("expected ErrKYCNotApproved, got: %v", err)
	}
}

func TestBathhouseService_Create_Gate_NoOffer(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	// Set up approved KYC but no offer
	expiresAt := time.Now().Add(365 * 24 * time.Hour)
	kyc := &domain.KYCApplication{
		ID: uuid.New(), UserID: ownerID, Status: domain.KYCStatusApproved,
		FullName: "Test", INN: "123456789012", ExpiresAt: &expiresAt,
	}
	_ = env.kycRepo.Create(context.Background(), kyc)

	_, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name: "Test", Address: "123 St", CityID: 1, PricePerHour: 5000, MinDuration: 1, MaxGuests: 10,
	})
	if !errors.Is(err, domain.ErrOfferNotAccepted) {
		t.Errorf("expected ErrOfferNotAccepted, got: %v", err)
	}
}

func TestBathhouseService_Create_Gate_NoPaymentDetails(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	// Set up KYC + offer but no payment details
	expiresAt := time.Now().Add(365 * 24 * time.Hour)
	kyc := &domain.KYCApplication{
		ID: uuid.New(), UserID: ownerID, Status: domain.KYCStatusApproved,
		FullName: "Test", INN: "123456789012", ExpiresAt: &expiresAt,
	}
	_ = env.kycRepo.Create(context.Background(), kyc)
	offer := &domain.OfferAcceptance{
		ID: uuid.New(), UserID: ownerID, OfferVersion: "1.0",
	}
	_ = env.offerRepo.Create(context.Background(), offer)

	_, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name: "Test", Address: "123 St", CityID: 1, PricePerHour: 5000, MinDuration: 1, MaxGuests: 10,
	})
	if !errors.Is(err, domain.ErrPaymentDetailsNotSet) {
		t.Errorf("expected ErrPaymentDetailsNotSet, got: %v", err)
	}
}

func TestBathhouseService_Create_Gate_AllPassed(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	env.setupOnboardingGate(t, ownerID)

	bh, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name: "Gate Test Bath", Address: "123 St", CityID: 1, PricePerHour: 5000, MinDuration: 1, MaxGuests: 10,
	})
	if err != nil {
		t.Fatalf("expected success with all gates passed, got: %v", err)
	}
	if bh.Status != domain.BathhouseStatusPending {
		t.Errorf("status = %q, want %q", bh.Status, domain.BathhouseStatusPending)
	}
}

func TestBathhouseService_CheckCompleteness_AllRequired(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         "Полная баня",
		Description:  "Это описание бани длиной более пятидесяти символов для прохождения проверки",
		Address:      "ул. Тестовая 123",
		CityID:       1,
		Latitude:     55.75,
		Longitude:    37.62,
		PricePerHour: 5000,
		MaxGuests:    10,
		MinDuration:  1,
		HasPool:      true,
		HasSauna:     true,
		HasSteamRoom: true,
		WorkingHours: []domain.WorkingHours{{DayOfWeek: 0, OpenTime: "09:00", CloseTime: "22:00"}},
		Status:       domain.BathhouseStatusPending,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	// Add 3 verified photos
	for i := 0; i < 3; i++ {
		_ = env.photoRepo.Create(context.Background(), &domain.BathhousePhoto{
			ID:          uuid.New(),
			BathhouseID: bh.ID,
			URL:         "https://example.com/photo.jpg",
			Status:      domain.PhotoStatusVerified,
			Position:    i,
		})
	}

	result, err := env.svc.CheckCompleteness(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Ready {
		t.Errorf("expected Ready=true, got false")
		for _, item := range result.Items {
			if !item.Complete {
				t.Logf("incomplete: %s (%s)", item.Field, item.Label)
			}
		}
	}
	if result.DoneRequired != result.TotalRequired {
		t.Errorf("DoneRequired=%d, TotalRequired=%d", result.DoneRequired, result.TotalRequired)
	}
	if result.Score != 100 {
		t.Errorf("expected score=100, got %d", result.Score)
	}
}

func TestBathhouseService_CheckCompleteness_MissingFields(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         "Баня",
		Description:  "Короткое",
		PricePerHour: 5000,
		MaxGuests:    10,
		MinDuration:  1,
		Status:       domain.BathhouseStatusPending,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	result, err := env.svc.CheckCompleteness(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Ready {
		t.Errorf("expected Ready=false for incomplete listing")
	}

	// Check specific incomplete fields
	incomplete := map[string]bool{}
	for _, item := range result.Items {
		if !item.Complete {
			incomplete[item.Field] = true
		}
	}
	for _, field := range []string{"description", "address", "city", "coordinates", "photos", "schedule"} {
		if !incomplete[field] {
			t.Errorf("expected field %q to be incomplete", field)
		}
	}
}

func TestBathhouseService_CheckCompleteness_Forbidden(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	otherID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Test", Address: "A", CityID: 1,
		PricePerHour: 1000, MinDuration: 1, MaxGuests: 5, Status: domain.BathhouseStatusPending,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	_, err := env.svc.CheckCompleteness(context.Background(), otherID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestBathhouseService_SubmitForModeration_IncompleteBlocked(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Test", Description: "Short",
		PricePerHour: 1000, MinDuration: 1, MaxGuests: 5, Status: domain.BathhouseStatusRejected,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	err := env.svc.SubmitForModeration(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrListingIncomplete) {
		t.Errorf("expected ErrListingIncomplete, got: %v", err)
	}
}

func TestBathhouseService_SubmitForModeration_Success(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         "Готовая баня",
		Description:  "Это описание бани длиной более пятидесяти символов для прохождения проверки",
		Address:      "ул. Тестовая 123",
		CityID:       1,
		Latitude:     55.75,
		Longitude:    37.62,
		PricePerHour: 5000,
		MaxGuests:    10,
		MinDuration:  1,
		WorkingHours: []domain.WorkingHours{{DayOfWeek: 0, OpenTime: "09:00", CloseTime: "22:00"}},
		Status:       domain.BathhouseStatusRejected,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	for i := 0; i < 3; i++ {
		_ = env.photoRepo.Create(context.Background(), &domain.BathhousePhoto{
			ID: uuid.New(), BathhouseID: bh.ID, URL: "https://example.com/photo.jpg",
			Status: domain.PhotoStatusVerified, Position: i,
		})
	}

	err := env.svc.SubmitForModeration(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify status changed to pending
	updated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	if updated.Status != domain.BathhouseStatusPending {
		t.Errorf("status = %q, want %q", updated.Status, domain.BathhouseStatusPending)
	}
}
