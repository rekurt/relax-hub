package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/antifraud"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type bathhouseTestEnv struct {
	svc          service.BathhouseService
	bhRepo       *mock.BathhouseRepo
	repRepo      *mock.RepresentativeRepo
	bookingRepo  *mock.BookingRepo
	photoRepo    *mock.BathhousePhotoRepo
	kycRepo      *mock.KYCRepo
	offerRepo    *mock.OfferRepo
	pdRepo       *mock.PaymentDetailsRepo
	stoplistRepo *mock.StoplistRepo
	fraudFlagRepo *mock.FraudFlagRepo
	userRepo     *mock.UserRepo
}

func newBathhouseTestEnv() *bathhouseTestEnv {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	bookingRepo := mock.NewBookingRepo()
	photoRepo := mock.NewBathhousePhotoRepo()
	kycRepo := mock.NewKYCRepo()
	offerRepo := mock.NewOfferRepo()
	pdRepo := mock.NewPaymentDetailsRepo()
	userRepo := mock.NewUserRepo()
	stoplistRepo := mock.NewStoplistRepo().(*mock.StoplistRepo)
	fraudFlagRepo := mock.NewFraudFlagRepo().(*mock.FraudFlagRepo)
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	kycSvc := service.NewKYCService(kycRepo, &noopNotifService{}, log)
	offerSvc := service.NewOfferService(offerRepo, log)
	pdSvc := service.NewPaymentDetailsService(pdRepo, log)
	auditLogRepo := mock.NewAuditLogRepo()
	auditSvc := service.NewAuditLogService(auditLogRepo, log)
	subRepo := mock.NewSubscriptionRepo()
	fraudEngine := antifraud.NewFraudEngine(fraudFlagRepo, log)
	svc := service.NewBathhouseService(bhRepo, bookingRepo, photoRepo, subRepo, access, kycSvc, offerSvc, pdSvc, auditSvc, fraudEngine, stoplistRepo, userRepo, pdRepo, log)
	return &bathhouseTestEnv{
		svc: svc, bhRepo: bhRepo, repRepo: repRepo, bookingRepo: bookingRepo,
		photoRepo: photoRepo, kycRepo: kycRepo, offerRepo: offerRepo, pdRepo: pdRepo,
		stoplistRepo: stoplistRepo, fraudFlagRepo: fraudFlagRepo, userRepo: userRepo,
	}
}

// setupOnboardingGate prepares KYC, offer, payment details, and user for a user so they pass the gate.
func (e *bathhouseTestEnv) setupOnboardingGate(t *testing.T, userID uuid.UUID) {
	t.Helper()

	// User record (needed for antifraud checks)
	user := &domain.User{
		ID:    userID,
		Email: fmt.Sprintf("owner-%s@test.com", userID.String()[:8]),
		Phone: "+79001234567",
		Role:  domain.RoleOwner,
		Name:  "Test User",
	}
	if err := e.userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("setup user: %v", err)
	}

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

func TestBathhouseService_Create_BlockedByStoplist(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	env.setupOnboardingGate(t, ownerID)

	// Add the owner's phone to stoplist
	user, _ := env.userRepo.GetByID(context.Background(), ownerID)
	_ = env.stoplistRepo.Create(context.Background(), &domain.StoplistEntry{
		Phone:  user.Phone,
		Reason: "fraud detected",
	})

	_, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name:         "Blocked Bath",
		Address:      "456 Street",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
	})

	if !errors.Is(err, domain.ErrFraudDetected) {
		t.Errorf("expected ErrFraudDetected, got: %v", err)
	}
}

func TestBathhouseService_Create_DuplicateOwnerFlagged(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	env.setupOnboardingGate(t, ownerID)

	// Set duplicate count > 0 (another owner shares same phone/email)
	env.stoplistRepo.SetDuplicateCount(1)

	bh, err := env.svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name:         "Dup Owner Bath",
		Address:      "789 Street",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
	})

	// Duplicate creates a flag but does NOT block creation
	if err != nil {
		t.Fatalf("expected no error (flag action, not block), got: %v", err)
	}
	if bh == nil {
		t.Fatal("expected bathhouse to be created")
	}

	// Verify a fraud flag was created
	result, _ := env.fraudFlagRepo.ListByUser(context.Background(), ownerID, 1, 10)
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 fraud flag, got %d", len(result.Items))
	}
	if result.Items[0].Rule != domain.FraudRuleListingDuplicate {
		t.Errorf("expected rule %s, got %s", domain.FraudRuleListingDuplicate, result.Items[0].Rule)
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
		Role:        domain.RepRoleManager,
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

func TestBathhouseService_Duplicate_Success(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	dup, err := env.svc.DuplicateBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dup.ID == bh.ID {
		t.Error("duplicate should have a new ID")
	}
	if dup.Name != bh.Name+" (копия)" {
		t.Errorf("name = %q, want %q", dup.Name, bh.Name+" (копия)")
	}
	if dup.Slug == bh.Slug {
		t.Error("duplicate should have a different slug")
	}
	if dup.Status != domain.BathhouseStatusInactive {
		t.Errorf("status = %q, want %q", dup.Status, domain.BathhouseStatusInactive)
	}
	if dup.OwnerID != ownerID {
		t.Errorf("ownerID = %v, want %v", dup.OwnerID, ownerID)
	}
	if dup.PricePerHour != bh.PricePerHour {
		t.Errorf("price = %d, want %d", dup.PricePerHour, bh.PricePerHour)
	}
	if dup.ApiKey == bh.ApiKey {
		t.Error("duplicate should have a new API key")
	}
	if len(dup.WorkingHours) != len(bh.WorkingHours) {
		t.Errorf("working hours count = %d, want %d", len(dup.WorkingHours), len(bh.WorkingHours))
	}
}

func TestBathhouseService_Duplicate_Forbidden(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	_, err := env.svc.DuplicateBathhouse(context.Background(), otherOwnerID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestBathhouseService_Duplicate_NotFound(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	_, err := env.svc.DuplicateBathhouse(context.Background(), ownerID, domain.RoleOwner, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestBathhouseService_Duplicate_RepresentativeAllowed(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	rep := &domain.Representative{
		ID: uuid.New(), UserID: repUserID, BathhouseID: bh.ID, OwnerID: ownerID,
		Role:        domain.RepRoleManager,
	}
	_ = env.repRepo.Create(context.Background(), rep)

	dup, err := env.svc.DuplicateBathhouse(context.Background(), repUserID, domain.RoleRepresentative, bh.ID)
	if err != nil {
		t.Fatalf("representative should be allowed to duplicate, got: %v", err)
	}
	if dup.Name != bh.Name+" (копия)" {
		t.Errorf("name = %q, want %q", dup.Name, bh.Name+" (копия)")
	}
}

// --- Deactivation & Archival Tests ---

func TestBathhouseService_Deactivate_Success(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID) // active status

	err := env.svc.DeactivateBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	if updated.Status != domain.BathhouseStatusInactive {
		t.Errorf("status = %q, want %q", updated.Status, domain.BathhouseStatusInactive)
	}
}

func TestBathhouseService_Deactivate_NotActive(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Test", Address: "A", CityID: 1,
		PricePerHour: 1000, MinDuration: 1, MaxGuests: 5, Status: domain.BathhouseStatusPending,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	err := env.svc.DeactivateBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for non-active bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Deactivate_Forbidden(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	otherID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	err := env.svc.DeactivateBathhouse(context.Background(), otherID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestBathhouseService_Activate_Success(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	// First deactivate
	_ = env.svc.DeactivateBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)

	// Then activate
	err := env.svc.ActivateBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	if updated.Status != domain.BathhouseStatusActive {
		t.Errorf("status = %q, want %q", updated.Status, domain.BathhouseStatusActive)
	}
}

func TestBathhouseService_Activate_NotInactive(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID) // active

	err := env.svc.ActivateBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for already-active bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Activate_FromArchived(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Test", Address: "A", CityID: 1,
		PricePerHour: 1000, MinDuration: 1, MaxGuests: 5, Status: domain.BathhouseStatusArchived,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	err := env.svc.ActivateBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for archived bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Archive_Success(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	err := env.svc.ArchiveBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	if updated.Status != domain.BathhouseStatusArchived {
		t.Errorf("status = %q, want %q", updated.Status, domain.BathhouseStatusArchived)
	}
}

func TestBathhouseService_Archive_AlreadyArchived(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Test", Address: "A", CityID: 1,
		PricePerHour: 1000, MinDuration: 1, MaxGuests: 5, Status: domain.BathhouseStatusArchived,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	err := env.svc.ArchiveBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for already-archived bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Archive_HasActiveBookings(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	// Create an active booking
	booking := &domain.Booking{
		ID:          uuid.New(),
		BathhouseID: bh.ID,
		UserID:      uuid.New(),
		Status:      domain.BookingConfirmed,
		StartTime:   time.Now().Add(24 * time.Hour),
		EndTime:     time.Now().Add(26 * time.Hour),
	}
	_ = env.bookingRepo.Create(context.Background(), booking)

	err := env.svc.ArchiveBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrBathhouseHasBookings) {
		t.Errorf("expected ErrBathhouseHasBookings, got: %v", err)
	}
}

func TestBathhouseService_Archive_Forbidden(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	otherID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	err := env.svc.ArchiveBathhouse(context.Background(), otherID, domain.RoleOwner, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestBathhouseService_Deactivate_HiddenFromSearch(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	_ = env.svc.DeactivateBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)

	// Public GetByID should not find inactive bathhouse
	_, err := env.svc.GetByID(context.Background(), bh.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("inactive bathhouse should not be visible via GetByID, got: %v", err)
	}
}

func TestBathhouseService_Archive_HiddenFromSearch(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	_ = env.svc.ArchiveBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)

	// Public GetByID should not find archived bathhouse
	_, err := env.svc.GetByID(context.Background(), bh.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("archived bathhouse should not be visible via GetByID, got: %v", err)
	}
}

func TestBathhouseService_Deactivate_RepresentativeAllowed(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	rep := &domain.Representative{
		ID: uuid.New(), UserID: repUserID, BathhouseID: bh.ID, OwnerID: ownerID,
		Role:        domain.RepRoleManager,
	}
	_ = env.repRepo.Create(context.Background(), rep)

	err := env.svc.DeactivateBathhouse(context.Background(), repUserID, domain.RoleRepresentative, bh.ID)
	if err != nil {
		t.Fatalf("representative should be allowed to deactivate, got: %v", err)
	}
}

func TestBathhouseService_Search_FullTextQuery(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	// Create two bathhouses with different names
	bh1 := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Русская баня на дровах",
		Description: "Лучшая баня в городе", Slug: "russkaya-banya",
		Address: "ул. Ленина 1", CityID: 1, PricePerHour: 3000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	bh2 := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Финская сауна",
		Description: "Настоящая финская сауна", Slug: "finskaya-sauna",
		Address: "ул. Мира 5", CityID: 1, PricePerHour: 4000,
		MaxGuests: 6, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(context.Background(), bh1)
	_ = env.bhRepo.Create(context.Background(), bh2)

	// Search for "баня" — should find only bh1
	q := "баня"
	result, err := env.svc.Search(context.Background(), domain.BathhouseFilter{
		SearchQuery: &q, Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 result for 'баня', got %d", result.TotalCount)
	}
	if result.TotalCount == 1 && result.Items[0].ID != bh1.ID {
		t.Errorf("expected bathhouse %s, got %s", bh1.ID, result.Items[0].ID)
	}

	// Search for "сауна" — should find only bh2
	q2 := "сауна"
	result, err = env.svc.Search(context.Background(), domain.BathhouseFilter{
		SearchQuery: &q2, Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 result for 'сауна', got %d", result.TotalCount)
	}

	// Empty search — should return all
	result, err = env.svc.Search(context.Background(), domain.BathhouseFilter{
		Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected 2 results for empty search, got %d", result.TotalCount)
	}
}

func TestBathhouseService_Search_SortByRelevance(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh1 := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Баня люкс",
		Description: "Описание", Slug: "banya-lux",
		Address: "ул. Ленина 1", CityID: 1, PricePerHour: 3000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(context.Background(), bh1)

	q := "баня"
	result, err := env.svc.Search(context.Background(), domain.BathhouseFilter{
		SearchQuery: &q, SortBy: "relevance", Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search with relevance sort: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 result, got %d", result.TotalCount)
	}
}

func TestBathhouseService_Search_NoResultsForNonMatching(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Русская баня",
		Description: "Классическая баня", Slug: "russkaya",
		Address: "ул. Мира 1", CityID: 1, PricePerHour: 2000,
		MaxGuests: 8, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	q := "бассейн"
	result, err := env.svc.Search(context.Background(), domain.BathhouseFilter{
		SearchQuery: &q, Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected 0 results for non-matching query, got %d", result.TotalCount)
	}
}

func TestBathhouseService_Search_DescriptionMatch(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Комплекс отдыха",
		Description: "Большой бассейн с подогревом", Slug: "kompleks",
		Address: "ул. Центральная 10", CityID: 1, PricePerHour: 5000,
		MaxGuests: 15, MinDuration: 2, Status: domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	q := "бассейн"
	result, err := env.svc.Search(context.Background(), domain.BathhouseFilter{
		SearchQuery: &q, Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 result matching description, got %d", result.TotalCount)
	}
}

func TestBathhouseService_Search_PrefixMatch(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh1 := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Русская баня на дровах",
		Description: "Традиционная парная", Slug: "russkaya-banya",
		Address: "ул. Ленина 1", CityID: 1, PricePerHour: 3000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	bh2 := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Финская сауна",
		Description: "Классическая сауна", Slug: "finskaya-sauna",
		Address: "ул. Мира 5", CityID: 1, PricePerHour: 4000,
		MaxGuests: 6, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(context.Background(), bh1)
	_ = env.bhRepo.Create(context.Background(), bh2)

	// Prefix "бан" should match "баня"
	q := "бан"
	result, err := env.svc.Search(context.Background(), domain.BathhouseFilter{
		SearchQuery: &q, Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search prefix: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 result for prefix 'бан', got %d", result.TotalCount)
	}
	if result.TotalCount == 1 && result.Items[0].ID != bh1.ID {
		t.Errorf("expected bathhouse %s, got %s", bh1.ID, result.Items[0].ID)
	}

	// Prefix "сау" should match "сауна"
	q2 := "сау"
	result, err = env.svc.Search(context.Background(), domain.BathhouseFilter{
		SearchQuery: &q2, Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search prefix: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 result for prefix 'сау', got %d", result.TotalCount)
	}
}

func TestBathhouseService_Search_AddressMatch(t *testing.T) {
	env := newBathhouseTestEnv()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Комплекс отдыха",
		Description: "Отличный отдых", Slug: "kompleks",
		Address: "Набережная улица 15", CityID: 1, PricePerHour: 5000,
		MaxGuests: 15, MinDuration: 2, Status: domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(context.Background(), bh)

	q := "Набережная"
	result, err := env.svc.Search(context.Background(), domain.BathhouseFilter{
		SearchQuery: &q, Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search address: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 result matching address, got %d", result.TotalCount)
	}
}

// --- Badge Tests ---

func containsBadge(badges []string, badge string) bool {
	for _, b := range badges {
		if b == badge {
			return true
		}
	}
	return false
}

func TestComputeBadges_New(t *testing.T) {
	env := newBathhouseTestEnv()
	ctx := context.Background()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: uuid.New(), Name: "Новая Баня", Slug: "novaya",
		Address: "ул. Новая 1", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		ReviewCount: 0, CreatedAt: time.Now(),
	}
	_ = env.bhRepo.Create(ctx, bh)

	badges := env.svc.ComputeBadges(ctx, bh)
	if !containsBadge(badges, "new") {
		t.Error("expected 'new' badge for recently created bathhouse with 0 reviews")
	}
	if containsBadge(badges, "top") {
		t.Error("unexpected 'top' badge")
	}
}

func TestComputeBadges_Top(t *testing.T) {
	env := newBathhouseTestEnv()
	ctx := context.Background()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: uuid.New(), Name: "Топ Баня", Slug: "top",
		Address: "ул. Топовая 1", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		BayesianRating: 4.7, ReviewCount: 15,
		CreatedAt: time.Now().Add(-60 * 24 * time.Hour),
	}
	_ = env.bhRepo.Create(ctx, bh)

	badges := env.svc.ComputeBadges(ctx, bh)
	if !containsBadge(badges, "top") {
		t.Error("expected 'top' badge for bathhouse with Bayesian >= 4.5 and 15 reviews")
	}
	if containsBadge(badges, "new") {
		t.Error("unexpected 'new' badge for 60-day-old bathhouse")
	}
}

func TestComputeBadges_Verified(t *testing.T) {
	env := newBathhouseTestEnv()
	ctx := context.Background()
	ownerID := uuid.New()
	env.setupOnboardingGate(t, ownerID)

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Верифицированная Баня", Slug: "verified",
		Address: "ул. Верная 1", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		IsPhotoVerified: true,
		CreatedAt:       time.Now().Add(-60 * 24 * time.Hour),
	}
	_ = env.bhRepo.Create(ctx, bh)

	badges := env.svc.ComputeBadges(ctx, bh)
	if !containsBadge(badges, "verified") {
		t.Error("expected 'verified' badge for photo-verified bathhouse with KYC approved")
	}
}

func TestComputeBadges_VerifiedNoKYC(t *testing.T) {
	env := newBathhouseTestEnv()
	ctx := context.Background()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: uuid.New(), Name: "Фото Баня", Slug: "photo",
		Address: "ул. Фотовая 1", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		IsPhotoVerified: true,
		CreatedAt:       time.Now().Add(-60 * 24 * time.Hour),
	}
	_ = env.bhRepo.Create(ctx, bh)

	badges := env.svc.ComputeBadges(ctx, bh)
	if containsBadge(badges, "verified") {
		t.Error("unexpected 'verified' badge - owner has no KYC")
	}
}

func TestComputeBadges_Premium(t *testing.T) {
	env := newBathhouseTestEnv()
	ctx := context.Background()
	bathhouseID := uuid.New()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: ownerID, Name: "Премиум Баня", Slug: "premium",
		Address: "ул. Премиум 1", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		CreatedAt: time.Now().Add(-60 * 24 * time.Hour),
	}
	_ = env.bhRepo.Create(ctx, bh)

	// Create active premium subscription
	sub := &domain.Subscription{
		ID: uuid.New(), BathhouseID: bathhouseID, OwnerID: ownerID,
		Plan: domain.PlanPremium, Status: domain.SubscriptionActive,
		StartDate: time.Now().Add(-24 * time.Hour), PriceKopecks: 100000,
	}
	subRepo := mock.NewSubscriptionRepo()
	_ = subRepo.Create(ctx, sub)

	// Create a service with subscription repo
	access := service.NewAccessChecker(env.repRepo, env.bhRepo)
	log := logger.New(logger.LevelWarn)
	kycSvc := service.NewKYCService(env.kycRepo, &noopNotifService{}, log)
	offerSvc := service.NewOfferService(env.offerRepo, log)
	pdSvc := service.NewPaymentDetailsService(env.pdRepo, log)
	auditLogRepo := mock.NewAuditLogRepo()
	auditSvc := service.NewAuditLogService(auditLogRepo, log)
	fraudEngine := antifraud.NewFraudEngine(env.fraudFlagRepo, log)
	svcWithSub := service.NewBathhouseService(env.bhRepo, env.bookingRepo, env.photoRepo, subRepo, access, kycSvc, offerSvc, pdSvc, auditSvc, fraudEngine, env.stoplistRepo, env.userRepo, env.pdRepo, log)

	badges := svcWithSub.ComputeBadges(ctx, bh)
	if !containsBadge(badges, "premium") {
		t.Error("expected 'premium' badge for bathhouse with active premium subscription")
	}
}

func TestComputeBadges_NoPremiumWithoutSubscription(t *testing.T) {
	env := newBathhouseTestEnv()
	ctx := context.Background()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: uuid.New(), Name: "Обычная Баня", Slug: "usual",
		Address: "ул. Обычная 1", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		CreatedAt: time.Now().Add(-60 * 24 * time.Hour),
	}
	_ = env.bhRepo.Create(ctx, bh)

	badges := env.svc.ComputeBadges(ctx, bh)
	if containsBadge(badges, "premium") {
		t.Error("unexpected 'premium' badge - no active subscription")
	}
}

func TestIsLastMinuteActive_Enabled_SlotInWindow(t *testing.T) {
	env := newBathhouseTestEnv()
	now := time.Now()
	wd := now.Weekday()
	dayOfWeek := int(wd) - 1
	if wd == time.Sunday {
		dayOfWeek = 6
	}
	// Open time 1 hour from now
	openTime := now.Add(1 * time.Hour).Format("15:04")

	bh := &domain.Bathhouse{
		LastMinuteEnabled:        true,
		LastMinuteDiscountPercent: 20,
		LastMinuteHoursThreshold: 6,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: dayOfWeek, OpenTime: openTime, CloseTime: "23:00"},
		},
	}
	if !env.svc.IsLastMinuteActive(bh) {
		t.Error("expected last-minute active for slot starting 1 hour from now with 6h threshold")
	}
}

func TestIsLastMinuteActive_Enabled_NoSlotInWindow(t *testing.T) {
	env := newBathhouseTestEnv()
	now := time.Now()
	wd := now.Weekday()
	dayOfWeek := int(wd) - 1
	if wd == time.Sunday {
		dayOfWeek = 6
	}
	// Open time 10 hours from now (beyond 6h threshold)
	openTime := now.Add(10 * time.Hour).Format("15:04")

	bh := &domain.Bathhouse{
		LastMinuteEnabled:        true,
		LastMinuteDiscountPercent: 20,
		LastMinuteHoursThreshold: 6,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: dayOfWeek, OpenTime: openTime, CloseTime: "23:59"},
		},
	}
	if env.svc.IsLastMinuteActive(bh) {
		t.Error("expected last-minute NOT active for slot starting 10 hours from now with 6h threshold")
	}
}

func TestIsLastMinuteActive_Disabled(t *testing.T) {
	env := newBathhouseTestEnv()
	bh := &domain.Bathhouse{
		LastMinuteEnabled:        false,
		LastMinuteDiscountPercent: 20,
		LastMinuteHoursThreshold: 6,
	}
	if env.svc.IsLastMinuteActive(bh) {
		t.Error("expected last-minute NOT active when feature is disabled")
	}
}

func TestComputeBadges_LastMinute(t *testing.T) {
	env := newBathhouseTestEnv()
	ctx := context.Background()
	now := time.Now()
	wd := now.Weekday()
	dayOfWeek := int(wd) - 1
	if wd == time.Sunday {
		dayOfWeek = 6
	}
	openTime := now.Add(1 * time.Hour).Format("15:04")

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: uuid.New(), Name: "Last Minute Баня", Slug: "lm",
		Address: "ул. Скидки 1", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		CreatedAt:                time.Now().Add(-60 * 24 * time.Hour),
		LastMinuteEnabled:        true,
		LastMinuteDiscountPercent: 20,
		LastMinuteHoursThreshold: 6,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: dayOfWeek, OpenTime: openTime, CloseTime: "23:00"},
		},
	}
	_ = env.bhRepo.Create(ctx, bh)

	badges := env.svc.ComputeBadges(ctx, bh)
	if !containsBadge(badges, "last_minute") {
		t.Error("expected 'last_minute' badge for bathhouse with active last-minute discount")
	}
}

func TestSearch_PromotedLimit(t *testing.T) {
	env := newBathhouseTestEnv()
	ctx := context.Background()

	// Create 5 bathhouses, all will show as "promoted" in mock
	// The mock repo doesn't check promotions table so we test the service capping logic
	// by verifying that Search() caps promoted count to 3
	for i := 0; i < 5; i++ {
		bh := &domain.Bathhouse{
			ID: uuid.New(), OwnerID: uuid.New(),
			Name: fmt.Sprintf("Баня %d", i), Slug: fmt.Sprintf("banya-%d", i),
			Address: "ул. Тест 1", CityID: 1, PricePerHour: 5000,
			MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
			IsPromoted: true,
		}
		_ = env.bhRepo.Create(ctx, bh)
	}

	result, err := env.svc.Search(ctx, domain.BathhouseFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	promotedCount := 0
	for _, bh := range result.Items {
		if bh.IsPromoted {
			promotedCount++
		}
	}
	if promotedCount > 3 {
		t.Errorf("expected max 3 promoted per page, got %d", promotedCount)
	}
}

func TestGetAreaAvgPrice(t *testing.T) {
	env := newBathhouseTestEnv()
	ctx := context.Background()

	// Create 3 active bathhouses in city 1
	for i := 0; i < 3; i++ {
		bh := &domain.Bathhouse{
			ID: uuid.New(), OwnerID: uuid.New(),
			Name: fmt.Sprintf("Баня avg %d", i), Slug: fmt.Sprintf("avg-%d", i),
			Address: "ул. Средняя 1", CityID: 1, PricePerHour: int64((i + 1) * 100000),
			MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		}
		_ = env.bhRepo.Create(ctx, bh)
	}

	// Prices: 100000, 200000, 300000 -> avg = 200000
	avgPrice, err := env.svc.GetAreaAvgPrice(ctx, 1, 55.7, 37.6)
	if err != nil {
		t.Fatalf("GetAreaAvgPrice failed: %v", err)
	}
	if avgPrice != 200000 {
		t.Errorf("expected avg price 200000, got %d", avgPrice)
	}

	// City with no bathhouses
	avgPrice, err = env.svc.GetAreaAvgPrice(ctx, 999, 55.7, 37.6)
	if err != nil {
		t.Fatalf("GetAreaAvgPrice for empty city failed: %v", err)
	}
	if avgPrice != 0 {
		t.Errorf("expected avg price 0 for empty city, got %d", avgPrice)
	}
}
