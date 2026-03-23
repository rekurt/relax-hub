package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type kycTestEnv struct {
	svc     service.KYCService
	kycRepo *mock.KYCRepo
}

func newKYCTestEnv() *kycTestEnv {
	kycRepo := mock.NewKYCRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewKYCService(kycRepo, &noopNotifService{}, log)
	return &kycTestEnv{
		svc:     svc,
		kycRepo: kycRepo,
	}
}

func TestKYCService_Submit_Success(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()

	app, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов Иван Иванович",
		DocumentURLs: []string{"https://example.com/passport.jpg"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.UserID != userID {
		t.Errorf("user_id = %v, want %v", app.UserID, userID)
	}
	if app.Status != domain.KYCStatusPending {
		t.Errorf("status = %v, want pending", app.Status)
	}
	if app.EntityType != domain.KYCEntityIndividual {
		t.Errorf("entity_type = %v, want individual", app.EntityType)
	}
}

func TestKYCService_Submit_SoleProprietorRequiresINN(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()

	_, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntitySoleProprietor,
		FullName:     "Петров Пётр",
		DocumentURLs: []string{"https://example.com/doc.jpg"},
		// Missing INN and OGRNIP
	})
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestKYCService_Submit_LegalEntityRequiresCompanyName(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()

	_, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityLegalEntity,
		FullName:     "Директор ООО",
		INN:          "1234567890",
		DocumentURLs: []string{"https://example.com/doc.jpg"},
		// Missing CompanyName
	})
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestKYCService_Submit_DuplicatePending(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()

	_, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов",
		DocumentURLs: []string{"https://example.com/doc.jpg"},
	})
	if err != nil {
		t.Fatalf("first submit failed: %v", err)
	}

	_, err = env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов",
		DocumentURLs: []string{"https://example.com/doc2.jpg"},
	})
	if err != domain.ErrKYCPending {
		t.Fatalf("expected ErrKYCPending, got: %v", err)
	}
}

func TestKYCService_Submit_NoDocuments(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()

	_, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов",
		DocumentURLs: []string{},
	})
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestKYCService_Approve(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()
	adminID := uuid.New()

	app, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов",
		DocumentURLs: []string{"https://example.com/doc.jpg"},
	})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	err = env.svc.Approve(context.Background(), app.ID, adminID)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	// Check status
	updated, err := env.svc.GetStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}
	if updated.Status != domain.KYCStatusApproved {
		t.Errorf("status = %v, want approved", updated.Status)
	}
	if updated.ExpiresAt == nil {
		t.Error("expires_at should be set after approval")
	}
}

func TestKYCService_Reject(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()
	adminID := uuid.New()

	app, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов",
		DocumentURLs: []string{"https://example.com/doc.jpg"},
	})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	err = env.svc.Reject(context.Background(), app.ID, adminID, "Нечитаемый документ")
	if err != nil {
		t.Fatalf("reject failed: %v", err)
	}

	updated, err := env.svc.GetStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}
	if updated.Status != domain.KYCStatusRejected {
		t.Errorf("status = %v, want rejected", updated.Status)
	}
	if updated.RejectionReason != "Нечитаемый документ" {
		t.Errorf("rejection_reason = %v, want 'Нечитаемый документ'", updated.RejectionReason)
	}
}

func TestKYCService_Reject_EmptyReason(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()
	adminID := uuid.New()

	app, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов",
		DocumentURLs: []string{"https://example.com/doc.jpg"},
	})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	err = env.svc.Reject(context.Background(), app.ID, adminID, "")
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for empty reason, got: %v", err)
	}
}

func TestKYCService_IsApproved(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()
	adminID := uuid.New()

	// Not found -> false
	approved, err := env.svc.IsApproved(context.Background(), userID)
	if err == nil {
		t.Fatalf("expected error for non-existent user, got approved=%v", approved)
	}

	// Submit + pending -> false
	app, _ := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов",
		DocumentURLs: []string{"https://example.com/doc.jpg"},
	})

	approved, err = env.svc.IsApproved(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved {
		t.Error("expected not approved for pending KYC")
	}

	// Approve -> true
	_ = env.svc.Approve(context.Background(), app.ID, adminID)
	approved, err = env.svc.IsApproved(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved {
		t.Error("expected approved after approval")
	}
}

func TestKYCService_ListPending(t *testing.T) {
	env := newKYCTestEnv()

	// Submit 3 applications
	for i := 0; i < 3; i++ {
		_, err := env.svc.Submit(context.Background(), uuid.New(), service.SubmitKYCInput{
			EntityType:   domain.KYCEntityIndividual,
			FullName:     "User",
			DocumentURLs: []string{"https://example.com/doc.jpg"},
		})
		if err != nil {
			t.Fatalf("submit %d failed: %v", i, err)
		}
	}

	result, err := env.svc.ListPending(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("list pending failed: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
	if len(result.Items) != 3 {
		t.Errorf("items count = %d, want 3", len(result.Items))
	}
}

func TestKYCService_Submit_AfterRejection(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()
	adminID := uuid.New()

	app, _ := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов",
		DocumentURLs: []string{"https://example.com/doc.jpg"},
	})

	_ = env.svc.Reject(context.Background(), app.ID, adminID, "Плохое качество")

	// Should be able to submit again after rejection
	newApp, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityIndividual,
		FullName:     "Иванов",
		DocumentURLs: []string{"https://example.com/better_doc.jpg"},
	})
	if err != nil {
		t.Fatalf("re-submit after rejection failed: %v", err)
	}
	if newApp.Status != domain.KYCStatusPending {
		t.Errorf("status = %v, want pending", newApp.Status)
	}
}

func TestKYCService_Submit_SoleProprietorSuccess(t *testing.T) {
	env := newKYCTestEnv()
	userID := uuid.New()

	app, err := env.svc.Submit(context.Background(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntitySoleProprietor,
		FullName:     "ИП Петров",
		INN:          "123456789012",
		OGRNIP:       "315000000000000",
		DocumentURLs: []string{"https://example.com/doc.jpg"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.EntityType != domain.KYCEntitySoleProprietor {
		t.Errorf("entity_type = %v, want sole_proprietor", app.EntityType)
	}
}
