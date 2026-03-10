package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type photoVerifTestEnv struct {
	svc       service.PhotoVerificationService
	photoRepo *mock.BathhousePhotoRepo
	bhRepo    *mock.BathhouseRepo
	repRepo   *mock.RepresentativeRepo
}

func newPhotoVerifTestEnv() *photoVerifTestEnv {
	photoRepo := mock.NewBathhousePhotoRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	accessCheck := service.NewAccessChecker(repRepo, bhRepo)
	svc := service.NewPhotoVerificationService(photoRepo, bhRepo, accessCheck, &noopNotifService{}, logger.New(logger.LevelWarn))
	return &photoVerifTestEnv{
		svc:       svc,
		photoRepo: photoRepo,
		bhRepo:    bhRepo,
		repRepo:   repRepo,
	}
}

func createTestBathhouse(env *photoVerifTestEnv, ownerID uuid.UUID) *domain.Bathhouse {
	bh := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         "Test Bathhouse",
		Address:      "Test Address",
		CityID:       1,
		PricePerHour: 500000,
		MinDuration:  1,
		MaxGuests:    10,
		Status:       domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(context.Background(), bh)
	return bh
}

func TestPhotoVerification_UploadPhoto_Success(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	photo, err := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID:  bh.ID,
		URL:          "https://example.com/photo1.jpg",
		ThumbnailURL: "https://example.com/photo1_thumb.jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if photo.BathhouseID != bh.ID {
		t.Errorf("bathhouse_id = %v, want %v", photo.BathhouseID, bh.ID)
	}
	if photo.Status != domain.PhotoStatusPending {
		t.Errorf("status = %v, want pending", photo.Status)
	}
	if photo.Position != 0 {
		t.Errorf("position = %d, want 0", photo.Position)
	}
}

func TestPhotoVerification_UploadPhoto_Forbidden(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	otherUser := uuid.New()

	_, err := env.svc.UploadPhoto(context.Background(), otherUser, domain.RoleClient, service.UploadPhotoInput{
		BathhouseID:  bh.ID,
		URL:          "https://example.com/photo1.jpg",
		ThumbnailURL: "https://example.com/photo1_thumb.jpg",
	})
	if err != domain.ErrForbidden {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestPhotoVerification_UploadPhoto_AdminAllowed(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	photo, err := env.svc.UploadPhoto(context.Background(), adminID, domain.RoleAdmin, service.UploadPhotoInput{
		BathhouseID:  bh.ID,
		URL:          "https://example.com/photo1.jpg",
		ThumbnailURL: "https://example.com/photo1_thumb.jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if photo.Status != domain.PhotoStatusPending {
		t.Errorf("status = %v, want pending", photo.Status)
	}
}

func TestPhotoVerification_UploadPhoto_PositionAutoIncrement(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	for i := 0; i < 3; i++ {
		photo, err := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
			BathhouseID: bh.ID,
			URL:         "https://example.com/photo.jpg",
		})
		if err != nil {
			t.Fatalf("upload %d: %v", i, err)
		}
		if photo.Position != i {
			t.Errorf("photo %d: position = %d, want %d", i, photo.Position, i)
		}
	}
}

func TestPhotoVerification_DeletePhoto_ByOwner(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	photo, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID,
		URL:         "https://example.com/photo.jpg",
	})

	err := env.svc.DeletePhoto(context.Background(), photo.ID, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	photos, _ := env.svc.ListByBathhouse(context.Background(), bh.ID)
	if len(photos) != 0 {
		t.Errorf("photos count = %d, want 0", len(photos))
	}
}

func TestPhotoVerification_DeletePhoto_ByAdmin(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	photo, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID,
		URL:         "https://example.com/photo.jpg",
	})

	err := env.svc.DeletePhoto(context.Background(), photo.ID, adminID, domain.RoleAdmin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPhotoVerification_DeletePhoto_Forbidden(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	otherUser := uuid.New()

	photo, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID,
		URL:         "https://example.com/photo.jpg",
	})

	err := env.svc.DeletePhoto(context.Background(), photo.ID, otherUser, domain.RoleClient)
	if err != domain.ErrForbidden {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestPhotoVerification_DeletePhoto_NotFound(t *testing.T) {
	env := newPhotoVerifTestEnv()
	err := env.svc.DeletePhoto(context.Background(), uuid.New(), uuid.New(), domain.RoleAdmin)
	if err != domain.ErrPhotoNotFound {
		t.Errorf("err = %v, want ErrPhotoNotFound", err)
	}
}

func TestPhotoVerification_ReorderPhotos(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	p1, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/1.jpg",
	})
	p2, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/2.jpg",
	})

	// Reverse the order
	err := env.svc.ReorderPhotos(context.Background(), bh.ID, ownerID, domain.RoleOwner, []uuid.UUID{p2.ID, p1.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	photos, _ := env.svc.ListByBathhouse(context.Background(), bh.ID)
	if len(photos) != 2 {
		t.Fatalf("photos count = %d, want 2", len(photos))
	}
	if photos[0].ID != p2.ID {
		t.Errorf("first photo = %v, want %v", photos[0].ID, p2.ID)
	}
	if photos[1].ID != p1.ID {
		t.Errorf("second photo = %v, want %v", photos[1].ID, p1.ID)
	}
}

func TestPhotoVerification_ReorderPhotos_Forbidden(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	err := env.svc.ReorderPhotos(context.Background(), bh.ID, uuid.New(), domain.RoleClient, []uuid.UUID{uuid.New()})
	if err != domain.ErrForbidden {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestPhotoVerification_VerifyPhoto_Success(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	photo, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/photo.jpg",
	})

	verified, err := env.svc.VerifyPhoto(context.Background(), photo.ID, adminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verified.Status != domain.PhotoStatusVerified {
		t.Errorf("status = %v, want verified", verified.Status)
	}
	if verified.VerifiedByID == nil || *verified.VerifiedByID != adminID {
		t.Errorf("verified_by_id = %v, want %v", verified.VerifiedByID, adminID)
	}
	if verified.VerifiedAt == nil {
		t.Error("verified_at should not be nil")
	}
}

func TestPhotoVerification_VerifyPhoto_NotPending(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	photo, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/photo.jpg",
	})

	_, _ = env.svc.VerifyPhoto(context.Background(), photo.ID, adminID)

	// Try to verify again
	_, err := env.svc.VerifyPhoto(context.Background(), photo.ID, adminID)
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPhotoVerification_VerifyPhoto_NotFound(t *testing.T) {
	env := newPhotoVerifTestEnv()
	_, err := env.svc.VerifyPhoto(context.Background(), uuid.New(), uuid.New())
	if err != domain.ErrPhotoNotFound {
		t.Errorf("err = %v, want ErrPhotoNotFound", err)
	}
}

func TestPhotoVerification_RejectPhoto_Success(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	photo, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/photo.jpg",
	})

	rejected, err := env.svc.RejectPhoto(context.Background(), photo.ID, adminID, "Low quality image")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rejected.Status != domain.PhotoStatusRejected {
		t.Errorf("status = %v, want rejected", rejected.Status)
	}
	if rejected.RejectionReason != "Low quality image" {
		t.Errorf("rejection_reason = %q, want %q", rejected.RejectionReason, "Low quality image")
	}
}

func TestPhotoVerification_RejectPhoto_NotPending(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	photo, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/photo.jpg",
	})

	_, _ = env.svc.RejectPhoto(context.Background(), photo.ID, adminID, "Bad")

	_, err := env.svc.RejectPhoto(context.Background(), photo.ID, adminID, "Still bad")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPhotoVerification_AllPhotosVerified_SetsFlag(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	p1, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/1.jpg",
	})
	p2, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/2.jpg",
	})

	// Verify first photo
	_, _ = env.svc.VerifyPhoto(context.Background(), p1.ID, adminID)

	// Bathhouse should NOT be verified yet
	bhUpdated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	if bhUpdated.IsPhotoVerified {
		t.Error("bathhouse should not be photo-verified with only one of two photos verified")
	}

	// Verify second photo
	_, _ = env.svc.VerifyPhoto(context.Background(), p2.ID, adminID)

	// Now bathhouse should be verified
	bhUpdated, _ = env.bhRepo.GetByID(context.Background(), bh.ID)
	if !bhUpdated.IsPhotoVerified {
		t.Error("bathhouse should be photo-verified when all photos are verified")
	}
}

func TestPhotoVerification_NewUpload_ResetsFlag(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	p1, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/1.jpg",
	})
	_, _ = env.svc.VerifyPhoto(context.Background(), p1.ID, adminID)

	// Bathhouse should be verified with one photo verified
	bhUpdated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	if !bhUpdated.IsPhotoVerified {
		t.Error("bathhouse should be photo-verified")
	}

	// Upload new photo -> resets verification
	_, _ = env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/2.jpg",
	})

	bhUpdated, _ = env.bhRepo.GetByID(context.Background(), bh.ID)
	if bhUpdated.IsPhotoVerified {
		t.Error("bathhouse should not be photo-verified after new upload")
	}
}

func TestPhotoVerification_GetPendingPhotos(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	for i := 0; i < 3; i++ {
		_, _ = env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
			BathhouseID: bh.ID, URL: "https://example.com/photo.jpg",
		})
	}

	result, err := env.svc.GetPendingPhotos(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
}

func TestPhotoVerification_ListByBathhouse(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	for i := 0; i < 2; i++ {
		_, _ = env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
			BathhouseID: bh.ID, URL: "https://example.com/photo.jpg",
		})
	}

	photos, err := env.svc.ListByBathhouse(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(photos) != 2 {
		t.Errorf("photos count = %d, want 2", len(photos))
	}
}

func TestPhotoVerification_DeletePhoto_RecalcsVerification(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	p1, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/1.jpg",
	})
	p2, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/2.jpg",
	})

	// Verify only p1
	_, _ = env.svc.VerifyPhoto(context.Background(), p1.ID, adminID)

	// Not verified because p2 is still pending
	bhUpdated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	if bhUpdated.IsPhotoVerified {
		t.Error("should not be verified")
	}

	// Delete pending p2 -> only verified p1 remains
	_ = env.svc.DeletePhoto(context.Background(), p2.ID, ownerID, domain.RoleOwner)

	bhUpdated, _ = env.bhRepo.GetByID(context.Background(), bh.ID)
	if !bhUpdated.IsPhotoVerified {
		t.Error("should be verified after deleting the unverified photo")
	}
}

func TestPhotoVerification_ReorderPhotos_IncompleteList(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	p1, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/1.jpg",
	})
	_, _ = env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/2.jpg",
	})

	// Try to reorder with only one of two photos
	err := env.svc.ReorderPhotos(context.Background(), bh.ID, ownerID, domain.RoleOwner, []uuid.UUID{p1.ID})
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPhotoVerification_ReorderPhotos_EmptyList(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	err := env.svc.ReorderPhotos(context.Background(), bh.ID, ownerID, domain.RoleOwner, []uuid.UUID{})
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPhotoVerification_ListVerifiedByBathhouse(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	p1, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/1.jpg",
	})
	_, _ = env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/2.jpg",
	})

	// Verify only p1
	_, _ = env.svc.VerifyPhoto(context.Background(), p1.ID, adminID)

	// ListVerifiedByBathhouse should return only verified photos
	photos, err := env.svc.ListVerifiedByBathhouse(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(photos) != 1 {
		t.Errorf("photos count = %d, want 1", len(photos))
	}
	if len(photos) > 0 && photos[0].ID != p1.ID {
		t.Errorf("photo ID = %v, want %v", photos[0].ID, p1.ID)
	}

	// ListByBathhouse should return all photos
	allPhotos, err := env.svc.ListByBathhouse(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(allPhotos) != 2 {
		t.Errorf("all photos count = %d, want 2", len(allPhotos))
	}
}

func TestPhotoVerification_RejectPhoto_EmptyReason(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)
	adminID := uuid.New()

	photo, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/photo.jpg",
	})

	_, err := env.svc.RejectPhoto(context.Background(), photo.ID, adminID, "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for empty rejection reason", err)
	}
}

func TestPhotoVerification_UploadPhoto_PositionAfterDeletion(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	p1, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/1.jpg",
	})
	p2, _ := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/2.jpg",
	})

	// Delete first photo (position 0)
	_ = env.svc.DeletePhoto(context.Background(), p1.ID, ownerID, domain.RoleOwner)

	// Upload new photo - should not collide with p2's position (1)
	p3, err := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/3.jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// p2 is at position 1, so p3 should be at position 2
	if p3.Position != 2 {
		t.Errorf("position = %d, want 2 (p2 is at 1)", p3.Position)
	}

	_ = p2 // used above in setup
}

func TestPhotoVerification_UploadPhoto_MaxPhotosExceeded(t *testing.T) {
	env := newPhotoVerifTestEnv()
	ownerID := uuid.New()
	bh := createTestBathhouse(env, ownerID)

	// Upload 30 photos (the maximum)
	for i := 0; i < 30; i++ {
		_, err := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
			BathhouseID: bh.ID, URL: fmt.Sprintf("https://example.com/photo%d.jpg", i),
		})
		if err != nil {
			t.Fatalf("upload photo %d: %v", i, err)
		}
	}

	// 31st upload should fail
	_, err := env.svc.UploadPhoto(context.Background(), ownerID, domain.RoleOwner, service.UploadPhotoInput{
		BathhouseID: bh.ID, URL: "https://example.com/overflow.jpg",
	})
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for exceeding max photos, got %v", err)
	}
}
