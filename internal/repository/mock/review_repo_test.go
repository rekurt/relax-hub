package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func TestReviewRepo_FullCoverage(t *testing.T) {
	ctx := context.Background()
	repo := NewReviewRepo()

	bhID := uuid.New()
	userID := uuid.New()
	bookingID := uuid.New()

	mkRating := func(v float64) *float64 { return &v }

	t.Run("CreateAndGetByID", func(t *testing.T) {
		rev := &domain.Review{
			UserID:        userID,
			BathhouseID:   bhID,
			BookingID:     bookingID,
			Rating:        5,
			Cleanliness:   mkRating(5),
			Accuracy:      mkRating(5),
			Communication: mkRating(4.5),
			ValueForMoney: mkRating(5),
			Status:        domain.ReviewStatusApproved,
		}
		if err := repo.Create(ctx, rev); err != nil {
			t.Fatal(err)
		}

		got, err := repo.GetByID(ctx, rev.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Rating != 5 {
			t.Errorf("Rating: got %d", got.Rating)
		}

		// Duplicate (same user+booking) should fail.
		dup := &domain.Review{
			UserID:    userID,
			BookingID: bookingID,
			Rating:    4,
			Status:    domain.ReviewStatusPending,
		}
		if err := repo.Create(ctx, dup); err != domain.ErrAlreadyExists {
			t.Errorf("duplicate: %v", err)
		}

		if _, err := repo.GetByID(ctx, uuid.New()); err != domain.ErrNotFound {
			t.Errorf("missing GetByID: %v", err)
		}
	})

	t.Run("UpdateAndDelete", func(t *testing.T) {
		rev := &domain.Review{
			UserID:      uuid.New(),
			BathhouseID: bhID,
			BookingID:   uuid.New(),
			Rating:      4,
			Status:      domain.ReviewStatusPending,
		}
		_ = repo.Create(ctx, rev)
		rev.Text = "updated"
		if err := repo.Update(ctx, rev); err != nil {
			t.Fatal(err)
		}
		got, _ := repo.GetByID(ctx, rev.ID)
		if got.Text != "updated" {
			t.Errorf("Update text: %q", got.Text)
		}

		ghost := &domain.Review{ID: uuid.New()}
		if err := repo.Update(ctx, ghost); err != domain.ErrNotFound {
			t.Errorf("Update missing: %v", err)
		}

		if err := repo.Delete(ctx, rev.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if err := repo.Delete(ctx, rev.ID); err != domain.ErrNotFound {
			t.Errorf("Delete twice: %v", err)
		}
	})

	t.Run("ListByBathhouseAndFilters", func(t *testing.T) {
		fresh := NewReviewRepo()
		bh := uuid.New()
		other := uuid.New()
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uuid.New(), BathhouseID: bh, BookingID: uuid.New(),
			Rating: 5, Status: domain.ReviewStatusApproved,
		})
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uuid.New(), BathhouseID: bh, BookingID: uuid.New(),
			Rating: 2, Status: domain.ReviewStatusPending,
		})
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uuid.New(), BathhouseID: other, BookingID: uuid.New(),
			Rating: 5, Status: domain.ReviewStatusApproved,
		})

		page, _ := fresh.ListByBathhouse(ctx, bh, 1, 10)
		if page.TotalCount != 1 {
			t.Errorf("ListByBathhouse approved-only: %d", page.TotalCount)
		}

		// Filtered list
		bhFilter := bh
		st := domain.ReviewStatusPending
		minR := 2
		filtered, _ := fresh.ListByBathhouseFiltered(ctx, domain.ReviewFilter{
			BathhouseID: &bhFilter, Status: &st, MinRating: &minR, Page: 1, PageSize: 10,
		})
		if filtered.TotalCount != 1 {
			t.Errorf("Filtered list: %d", filtered.TotalCount)
		}

		// MinRating excludes lower-rated
		highMin := 4
		filtered, _ = fresh.ListByBathhouseFiltered(ctx, domain.ReviewFilter{
			BathhouseID: &bhFilter, MinRating: &highMin, Page: 1, PageSize: 10,
		})
		if filtered.TotalCount != 1 {
			t.Errorf("HighMin: %d, want 1", filtered.TotalCount)
		}
	})

	t.Run("GetByBookingIDAndUpdateStatus", func(t *testing.T) {
		fresh := NewReviewRepo()
		bID := uuid.New()
		rev := &domain.Review{
			UserID: uuid.New(), BathhouseID: bhID, BookingID: bID,
			Rating: 4, Status: domain.ReviewStatusPending,
		}
		_ = fresh.Create(ctx, rev)

		got, err := fresh.GetByBookingID(ctx, bID)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != rev.ID {
			t.Errorf("GetByBookingID id mismatch")
		}
		if _, err := fresh.GetByBookingID(ctx, uuid.New()); err != domain.ErrNotFound {
			t.Errorf("missing GetByBookingID: %v", err)
		}

		if err := fresh.UpdateStatus(ctx, rev.ID, domain.ReviewStatusApproved); err != nil {
			t.Fatal(err)
		}
		got2, _ := fresh.GetByID(ctx, rev.ID)
		if got2.Status != domain.ReviewStatusApproved {
			t.Errorf("UpdateStatus did not persist")
		}
		if err := fresh.UpdateStatus(ctx, uuid.New(), domain.ReviewStatusApproved); err != domain.ErrNotFound {
			t.Errorf("UpdateStatus missing: %v", err)
		}

		respondedAt := time.Now()
		if err := fresh.AddOwnerResponse(ctx, rev.ID, "thanks!", respondedAt); err != nil {
			t.Fatal(err)
		}
		got3, _ := fresh.GetByID(ctx, rev.ID)
		if got3.OwnerResponse != "thanks!" || got3.OwnerResponseAt == nil {
			t.Error("AddOwnerResponse did not persist")
		}
		if err := fresh.AddOwnerResponse(ctx, uuid.New(), "x", respondedAt); err != domain.ErrNotFound {
			t.Errorf("AddOwnerResponse missing: %v", err)
		}

		if err := fresh.UpdateStatusWithReasons(ctx, rev.ID, domain.ReviewStatusRejected, []string{"spam"}); err != nil {
			t.Fatal(err)
		}
		got4, _ := fresh.GetByID(ctx, rev.ID)
		if len(got4.RejectionReasons) != 1 {
			t.Errorf("RejectionReasons not persisted: %v", got4.RejectionReasons)
		}
		if err := fresh.UpdateStatusWithReasons(ctx, uuid.New(), domain.ReviewStatusRejected, nil); err != domain.ErrNotFound {
			t.Errorf("UpdateStatusWithReasons missing: %v", err)
		}
	})

	t.Run("UserStatsAndPlatformAvg", func(t *testing.T) {
		fresh := NewReviewRepo()
		uID := uuid.New()
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uID, BathhouseID: bhID, BookingID: uuid.New(),
			Rating: 5, Status: domain.ReviewStatusApproved,
		})
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uID, BathhouseID: bhID, BookingID: uuid.New(),
			Rating: 3, Status: domain.ReviewStatusApproved,
		})
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uID, BathhouseID: bhID, BookingID: uuid.New(),
			Rating: 4, Status: domain.ReviewStatusPending,
		})

		stats, _ := fresh.GetUserReviewStats(ctx, uID)
		if stats.ReviewCount != 2 {
			t.Errorf("ReviewCount: %d", stats.ReviewCount)
		}
		if stats.AvgRating != 4 {
			t.Errorf("AvgRating: %v", stats.AvgRating)
		}

		empty, _ := fresh.GetUserReviewStats(ctx, uuid.New())
		if empty.ReviewCount != 0 {
			t.Errorf("Empty ReviewCount: %d", empty.ReviewCount)
		}

		avg, _ := fresh.GetPlatformAverageRating(ctx)
		if avg != 4 {
			t.Errorf("Platform avg: %v", avg)
		}

		// Platform avg with no approved reviews.
		empty2 := NewReviewRepo()
		emptyAvg, _ := empty2.GetPlatformAverageRating(ctx)
		if emptyAvg != 0 {
			t.Errorf("Empty platform avg: %v", emptyAvg)
		}
	})

	t.Run("CountAndAdminFilters", func(t *testing.T) {
		fresh := NewReviewRepo()
		base := time.Now()
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uuid.New(), BathhouseID: bhID, BookingID: uuid.New(),
			Rating: 5, Status: domain.ReviewStatusApproved,
		})
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uuid.New(), BathhouseID: bhID, BookingID: uuid.New(),
			Rating: 2, Status: domain.ReviewStatusPending,
		})

		count, _ := fresh.CountPendingReviews(ctx)
		if count != 1 {
			t.Errorf("CountPendingReviews: %d", count)
		}

		bhFilter := bhID
		st := domain.ReviewStatusApproved
		minR := 1
		maxR := 5
		from := base.Add(-time.Hour)
		to := base.Add(time.Hour)
		page, _ := fresh.ListAllReviews(ctx, domain.AdminReviewFilter{
			BathhouseID: &bhFilter, Status: &st, MinRating: &minR, MaxRating: &maxR,
			FromDate: &from, ToDate: &to, Page: 1, PageSize: 10,
		})
		if page.TotalCount != 1 {
			t.Errorf("ListAllReviews: %d, want 1", page.TotalCount)
		}

		strictMin := 100
		excluded, _ := fresh.ListAllReviews(ctx, domain.AdminReviewFilter{MinRating: &strictMin})
		if excluded.TotalCount != 0 {
			t.Errorf("StrictMin: %d", excluded.TotalCount)
		}
		strictMax := 0
		exMax, _ := fresh.ListAllReviews(ctx, domain.AdminReviewFilter{MaxRating: &strictMax})
		if exMax.TotalCount != 0 {
			t.Errorf("StrictMax: %d", exMax.TotalCount)
		}
		future := base.Add(24 * time.Hour)
		exFrom, _ := fresh.ListAllReviews(ctx, domain.AdminReviewFilter{FromDate: &future})
		if exFrom.TotalCount != 0 {
			t.Errorf("ExFrom: %d", exFrom.TotalCount)
		}
		past := base.Add(-24 * time.Hour)
		exTo, _ := fresh.ListAllReviews(ctx, domain.AdminReviewFilter{ToDate: &past})
		if exTo.TotalCount != 0 {
			t.Errorf("ExTo: %d", exTo.TotalCount)
		}
	})

	t.Run("CriteriaAveragesAndUnrevealed", func(t *testing.T) {
		fresh := NewReviewRepo()
		bh := uuid.New()
		// One review with criteria
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uuid.New(), BathhouseID: bh, BookingID: uuid.New(),
			Rating: 4, Status: domain.ReviewStatusApproved,
			Cleanliness: mkRating(5), Accuracy: mkRating(4), Communication: mkRating(4), ValueForMoney: mkRating(3),
		})

		avgs, _ := fresh.GetCriteriaAverages(ctx, bh)
		if avgs.AvgCleanliness != 5 || avgs.AvgValueForMoney != 3 {
			t.Errorf("Averages: %+v", avgs)
		}

		// Empty bathhouse
		empty, _ := fresh.GetCriteriaAverages(ctx, uuid.New())
		if empty.AvgCleanliness != 0 {
			t.Errorf("Empty averages: %+v", empty)
		}

		// Unrevealed past deadline
		past := time.Now().Add(-time.Hour)
		future := time.Now().Add(time.Hour)
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uuid.New(), BathhouseID: bh, BookingID: uuid.New(),
			Rating: 4, Status: domain.ReviewStatusApproved,
			RevealAt: &past, IsRevealed: false,
		})
		_ = fresh.Create(ctx, &domain.Review{
			UserID: uuid.New(), BathhouseID: bh, BookingID: uuid.New(),
			Rating: 4, Status: domain.ReviewStatusApproved,
			RevealAt: &future, IsRevealed: false,
		})

		past2 := time.Now()
		ready, _ := fresh.ListUnrevealedPastDeadline(ctx, past2)
		if len(ready) != 1 {
			t.Errorf("ListUnrevealedPastDeadline: %d, want 1", len(ready))
		}
	})
}
