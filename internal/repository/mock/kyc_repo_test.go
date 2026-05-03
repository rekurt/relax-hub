package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func TestKYCRepo_FullCoverage(t *testing.T) {
	ctx := context.Background()
	repo := NewKYCRepo()

	userID := uuid.New()
	reviewer := uuid.New()

	t.Run("CRUD", func(t *testing.T) {
		app := &domain.KYCApplication{
			UserID:       userID,
			Status:       domain.KYCStatusPending,
			EntityType:   domain.KYCEntityIndividual,
			FullName:     "Иван Иванов",
			INN:          "1234567890",
			DocumentURLs: []string{"doc.pdf"},
		}
		if err := repo.Create(ctx, app); err != nil {
			t.Fatal(err)
		}
		got, err := repo.GetByID(ctx, app.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.FullName != "Иван Иванов" {
			t.Errorf("FullName: %q", got.FullName)
		}
		if _, err := repo.GetByID(ctx, uuid.New()); err != domain.ErrKYCNotFound {
			t.Errorf("GetByID missing: %v", err)
		}

		got.INN = "0000000000"
		if err := repo.Update(ctx, got); err != nil {
			t.Fatal(err)
		}
		ghost := &domain.KYCApplication{ID: uuid.New()}
		if err := repo.Update(ctx, ghost); err != domain.ErrKYCNotFound {
			t.Errorf("Update missing: %v", err)
		}
	})

	t.Run("GetByUserIDLatest", func(t *testing.T) {
		fresh := NewKYCRepo()
		uID := uuid.New()
		old := &domain.KYCApplication{
			UserID: uID, Status: domain.KYCStatusRejected,
			EntityType: domain.KYCEntityIndividual,
			FullName:   "old",
		}
		_ = fresh.Create(ctx, old)
		// Force older CreatedAt
		old.CreatedAt = time.Now().Add(-time.Hour)
		_ = fresh.Update(ctx, old)

		newer := &domain.KYCApplication{
			UserID: uID, Status: domain.KYCStatusPending,
			EntityType: domain.KYCEntityIndividual,
			FullName:   "newer",
		}
		_ = fresh.Create(ctx, newer)

		got, err := fresh.GetByUserID(ctx, uID)
		if err != nil {
			t.Fatal(err)
		}
		if got.FullName != "newer" {
			t.Errorf("Should pick latest by CreatedAt; got %q", got.FullName)
		}

		if _, err := fresh.GetByUserID(ctx, uuid.New()); err != domain.ErrKYCNotFound {
			t.Errorf("GetByUserID missing: %v", err)
		}
	})

	t.Run("ListPending", func(t *testing.T) {
		fresh := NewKYCRepo()
		_ = fresh.Create(ctx, &domain.KYCApplication{UserID: uuid.New(), Status: domain.KYCStatusPending, EntityType: domain.KYCEntityIndividual, FullName: "p1"})
		_ = fresh.Create(ctx, &domain.KYCApplication{UserID: uuid.New(), Status: domain.KYCStatusPending, EntityType: domain.KYCEntityIndividual, FullName: "p2"})
		_ = fresh.Create(ctx, &domain.KYCApplication{UserID: uuid.New(), Status: domain.KYCStatusApproved, EntityType: domain.KYCEntityIndividual, FullName: "a1"})

		page, err := fresh.ListPending(ctx, 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		if page.TotalCount != 2 {
			t.Errorf("ListPending: %d, want 2", page.TotalCount)
		}

		// pagination defaults
		page, _ = fresh.ListPending(ctx, 0, 0)
		if page.Page != 1 || page.PageSize != 20 {
			t.Errorf("Defaults: %+v", page)
		}

		// out of range
		oor, _ := fresh.ListPending(ctx, 99, 1)
		if len(oor.Items) != 0 {
			t.Errorf("Out of range: %d", len(oor.Items))
		}
	})

	t.Run("ApproveAndReject", func(t *testing.T) {
		fresh := NewKYCRepo()
		toApprove := &domain.KYCApplication{
			UserID: uuid.New(), Status: domain.KYCStatusPending,
			EntityType: domain.KYCEntityIndividual, FullName: "ok",
		}
		_ = fresh.Create(ctx, toApprove)
		expires := time.Now().Add(365 * 24 * time.Hour)
		if err := fresh.Approve(ctx, toApprove.ID, reviewer, expires); err != nil {
			t.Fatal(err)
		}
		got, _ := fresh.GetByID(ctx, toApprove.ID)
		if got.Status != domain.KYCStatusApproved {
			t.Errorf("Status: %v", got.Status)
		}
		if got.ReviewedBy == nil || *got.ReviewedBy != reviewer {
			t.Error("ReviewedBy not set")
		}
		// Cannot approve already-approved
		if err := fresh.Approve(ctx, toApprove.ID, reviewer, expires); err != domain.ErrKYCNotFound {
			t.Errorf("re-approve: %v", err)
		}
		if err := fresh.Approve(ctx, uuid.New(), reviewer, expires); err != domain.ErrKYCNotFound {
			t.Errorf("Approve missing: %v", err)
		}

		toReject := &domain.KYCApplication{
			UserID: uuid.New(), Status: domain.KYCStatusPending,
			EntityType: domain.KYCEntityIndividual, FullName: "no",
		}
		_ = fresh.Create(ctx, toReject)
		if err := fresh.Reject(ctx, toReject.ID, reviewer, "missing docs"); err != nil {
			t.Fatal(err)
		}
		got2, _ := fresh.GetByID(ctx, toReject.ID)
		if got2.Status != domain.KYCStatusRejected || got2.RejectionReason != "missing docs" {
			t.Errorf("Reject: %+v", got2)
		}
		if err := fresh.Reject(ctx, toReject.ID, reviewer, "x"); err != domain.ErrKYCNotFound {
			t.Errorf("re-reject: %v", err)
		}
		if err := fresh.Reject(ctx, uuid.New(), reviewer, "x"); err != domain.ErrKYCNotFound {
			t.Errorf("Reject missing: %v", err)
		}
	})

	t.Run("ListExpiredApproved", func(t *testing.T) {
		fresh := NewKYCRepo()
		past := time.Now().Add(-time.Hour)
		future := time.Now().Add(time.Hour)
		expiring := &domain.KYCApplication{
			UserID: uuid.New(), Status: domain.KYCStatusApproved,
			EntityType: domain.KYCEntityIndividual, FullName: "x",
			ExpiresAt: &past,
		}
		_ = fresh.Create(ctx, expiring)
		valid := &domain.KYCApplication{
			UserID: uuid.New(), Status: domain.KYCStatusApproved,
			EntityType: domain.KYCEntityIndividual, FullName: "y",
			ExpiresAt: &future,
		}
		_ = fresh.Create(ctx, valid)

		expired, err := fresh.ListExpiredApproved(ctx, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		if len(expired) != 1 {
			t.Errorf("Expired: %d, want 1", len(expired))
		}
	})
}
