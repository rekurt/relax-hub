package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// TestBookingRepo_FullCoverage exercises every BookingRepo method end-to-end so
// the mock implementation has parity coverage with the postgres impl.
func TestBookingRepo_FullCoverage(t *testing.T) {
	ctx := context.Background()
	repo := NewBookingRepo()

	bhID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	t.Run("CreateAndGetByID", func(t *testing.T) {
		b := &domain.Booking{
			UserID:      userID,
			BathhouseID: bhID,
			StartTime:   now.Add(time.Hour),
			EndTime:     now.Add(2 * time.Hour),
			GuestCount:  2,
			TotalPrice:  10000,
			Status:      domain.BookingPending,
		}
		if err := repo.Create(ctx, b); err != nil {
			t.Fatalf("Create: %v", err)
		}
		got, err := repo.GetByID(ctx, b.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.ID != b.ID {
			t.Errorf("got ID %v, want %v", got.ID, b.ID)
		}
		// Missing ID
		if _, err := repo.GetByID(ctx, uuid.New()); err != domain.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("UpdateStatusAndUpdate", func(t *testing.T) {
		b := &domain.Booking{
			UserID: userID, BathhouseID: bhID,
			StartTime: now, EndTime: now.Add(time.Hour),
			TotalPrice: 1000, Status: domain.BookingPending,
		}
		_ = repo.Create(ctx, b)
		if err := repo.UpdateStatus(ctx, b.ID, domain.BookingConfirmed); err != nil {
			t.Fatalf("UpdateStatus: %v", err)
		}
		got, _ := repo.GetByID(ctx, b.ID)
		if got.Status != domain.BookingConfirmed {
			t.Errorf("status: got %v, want confirmed", got.Status)
		}
		if err := repo.UpdateStatus(ctx, uuid.New(), domain.BookingCancelled); err != domain.ErrNotFound {
			t.Errorf("UpdateStatus missing: got %v", err)
		}

		got.GuestCount = 5
		if err := repo.Update(ctx, got); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got2, _ := repo.GetByID(ctx, b.ID)
		if got2.GuestCount != 5 {
			t.Errorf("Update did not persist GuestCount: got %d", got2.GuestCount)
		}
		ghost := *got
		ghost.ID = uuid.New()
		if err := repo.Update(ctx, &ghost); err != domain.ErrNotFound {
			t.Errorf("Update missing: got %v", err)
		}
	})

	t.Run("CheckAvailability", func(t *testing.T) {
		freshBh := uuid.New()
		start := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
		end := start.Add(2 * time.Hour)
		conflict := &domain.Booking{
			UserID: userID, BathhouseID: freshBh,
			StartTime: start, EndTime: end,
			Status: domain.BookingConfirmed,
		}
		_ = repo.Create(ctx, conflict)

		ok, _ := repo.CheckAvailability(ctx, freshBh, start, end)
		if ok {
			t.Error("overlap should be unavailable")
		}
		ok, _ = repo.CheckAvailability(ctx, freshBh, end, end.Add(time.Hour))
		if !ok {
			t.Error("non-overlap should be available")
		}

		// Excluding the conflicting booking should free the slot.
		ok, _ = repo.CheckAvailabilityExcluding(ctx, freshBh, start, end, conflict.ID)
		if !ok {
			t.Error("excluding conflict should be available")
		}
		ok, _ = repo.CheckAvailabilityExcluding(ctx, freshBh, start, end, uuid.New())
		if ok {
			t.Error("excluding unrelated id should still see conflict")
		}
	})

	t.Run("CreateWithAvailabilityCheck", func(t *testing.T) {
		freshBh := uuid.New()
		start := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)
		end := start.Add(time.Hour)
		first := &domain.Booking{
			UserID: userID, BathhouseID: freshBh,
			StartTime: start, EndTime: end,
			Status: domain.BookingConfirmed,
		}
		if err := repo.CreateWithAvailabilityCheck(ctx, first, start, end); err != nil {
			t.Fatalf("first create: %v", err)
		}
		second := &domain.Booking{
			UserID: userID, BathhouseID: freshBh,
			StartTime: start, EndTime: end,
			Status: domain.BookingPending,
		}
		err := repo.CreateWithAvailabilityCheck(ctx, second, start, end)
		if err != domain.ErrSlotUnavailable {
			t.Errorf("expected ErrSlotUnavailable, got %v", err)
		}
	})

	t.Run("GetOverlappingAndCounts", func(t *testing.T) {
		freshBh := uuid.New()
		uID := uuid.New()
		start := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
		_ = repo.Create(ctx, &domain.Booking{
			UserID: uID, BathhouseID: freshBh,
			StartTime: start, EndTime: start.Add(time.Hour),
			Status: domain.BookingConfirmed,
		})
		_ = repo.Create(ctx, &domain.Booking{
			UserID: uID, BathhouseID: freshBh,
			StartTime: start.Add(2 * time.Hour), EndTime: start.Add(3 * time.Hour),
			Status: domain.BookingPending,
		})

		overlaps, _ := repo.GetOverlapping(ctx, freshBh, start, start.Add(time.Hour))
		if len(overlaps) != 1 {
			t.Errorf("GetOverlapping: got %d, want 1", len(overlaps))
		}
		cnt, _ := repo.CountActiveByBathhouse(ctx, freshBh)
		if cnt != 2 {
			t.Errorf("CountActiveByBathhouse: got %d, want 2", cnt)
		}
		cnt, _ = repo.CountActiveByUser(ctx, uID)
		if cnt != 2 {
			t.Errorf("CountActiveByUser: got %d, want 2", cnt)
		}
	})

	t.Run("ListAllAndFilters", func(t *testing.T) {
		fresh := NewBookingRepo()
		uID := uuid.New()
		bID := uuid.New()
		for i := 0; i < 3; i++ {
			_ = fresh.Create(ctx, &domain.Booking{
				UserID: uID, BathhouseID: bID,
				StartTime: now.Add(time.Duration(i) * time.Hour),
				EndTime:   now.Add(time.Duration(i+1) * time.Hour),
				Status:    domain.BookingConfirmed, TotalPrice: 1000,
			})
		}

		page, _ := fresh.ListAll(ctx, domain.AdminBookingFilter{Page: 1, PageSize: 10})
		if page.TotalCount != 3 {
			t.Errorf("ListAll total: got %d, want 3", page.TotalCount)
		}

		uIDFilter := uID
		page, _ = fresh.ListAll(ctx, domain.AdminBookingFilter{UserID: &uIDFilter, Page: 1, PageSize: 2})
		if page.TotalCount != 3 || page.PageSize != 2 {
			t.Errorf("ListAll w/ user: got total=%d size=%d", page.TotalCount, page.PageSize)
		}

		bIDFilter := bID
		page, _ = fresh.ListAll(ctx, domain.AdminBookingFilter{BathhouseID: &bIDFilter})
		if page.TotalCount != 3 {
			t.Errorf("ListAll w/ bathhouse: got %d", page.TotalCount)
		}

		stat := domain.BookingPending
		page, _ = fresh.ListAll(ctx, domain.AdminBookingFilter{Status: &stat})
		if page.TotalCount != 0 {
			t.Errorf("ListAll w/ pending status: got %d, want 0", page.TotalCount)
		}

		past := now.Add(-24 * time.Hour)
		future := now.Add(24 * time.Hour)
		page, _ = fresh.ListAll(ctx, domain.AdminBookingFilter{FromDate: &past})
		if page.TotalCount != 3 {
			t.Errorf("ListAll FromDate: got %d", page.TotalCount)
		}
		page, _ = fresh.ListAll(ctx, domain.AdminBookingFilter{ToDate: &future})
		if page.TotalCount != 3 {
			t.Errorf("ListAll ToDate: got %d", page.TotalCount)
		}

		// Page out of range
		page, _ = fresh.ListAll(ctx, domain.AdminBookingFilter{Page: 99, PageSize: 1})
		if len(page.Items) != 0 {
			t.Errorf("ListAll out of range: got %d items", len(page.Items))
		}
	})

	t.Run("UserStats", func(t *testing.T) {
		fresh := NewBookingRepo()
		uID := uuid.New()
		_ = fresh.Create(ctx, &domain.Booking{UserID: uID, BathhouseID: bhID, Status: domain.BookingCompleted, TotalPrice: 1000})
		_ = fresh.Create(ctx, &domain.Booking{UserID: uID, BathhouseID: bhID, Status: domain.BookingCompleted, TotalPrice: 3000})
		_ = fresh.Create(ctx, &domain.Booking{UserID: uID, BathhouseID: bhID, Status: domain.BookingPending, TotalPrice: 999})
		stats, err := fresh.GetUserStats(ctx, uID)
		if err != nil {
			t.Fatal(err)
		}
		if stats.TotalVisits != 2 {
			t.Errorf("TotalVisits: got %d, want 2", stats.TotalVisits)
		}
		if stats.TotalSpent != 4000 {
			t.Errorf("TotalSpent: got %d, want 4000", stats.TotalSpent)
		}
		if stats.AvgCheck != 2000 {
			t.Errorf("AvgCheck: got %d, want 2000", stats.AvgCheck)
		}

		empty, _ := fresh.GetUserStats(ctx, uuid.New())
		if empty.TotalVisits != 0 {
			t.Errorf("empty stats: got %v", empty)
		}
	})

	t.Run("CheckinCheckoutCancellation", func(t *testing.T) {
		fresh := NewBookingRepo()
		b := &domain.Booking{
			UserID: userID, BathhouseID: bhID,
			StartTime: now, EndTime: now.Add(time.Hour),
			Status: domain.BookingConfirmed,
		}
		_ = fresh.Create(ctx, b)

		ci := now.Add(time.Minute)
		if err := fresh.UpdateCheckin(ctx, b.ID, &ci); err != nil {
			t.Fatalf("UpdateCheckin: %v", err)
		}
		got, _ := fresh.GetByID(ctx, b.ID)
		if got.CheckedInAt == nil {
			t.Error("CheckedInAt not set")
		}
		if err := fresh.UpdateCheckin(ctx, uuid.New(), &ci); err != domain.ErrNotFound {
			t.Errorf("UpdateCheckin missing: %v", err)
		}

		co := now.Add(2 * time.Hour)
		if err := fresh.UpdateCheckout(ctx, b.ID, &co, domain.BookingCompleted); err != nil {
			t.Fatalf("UpdateCheckout: %v", err)
		}
		if err := fresh.UpdateCheckout(ctx, uuid.New(), &co, domain.BookingCompleted); err != domain.ErrNotFound {
			t.Errorf("UpdateCheckout missing: %v", err)
		}

		if err := fresh.UpdateCancelledByOwner(ctx, b.ID); err != nil {
			t.Fatalf("UpdateCancelledByOwner: %v", err)
		}
		if err := fresh.UpdateCancelledByOwner(ctx, uuid.New()); err != domain.ErrNotFound {
			t.Errorf("UpdateCancelledByOwner missing: %v", err)
		}
	})

	t.Run("ListConfirmedWithoutCheckinAndUpcoming", func(t *testing.T) {
		fresh := NewBookingRepo()
		past := now.Add(-time.Hour)
		future := now.Add(time.Hour)
		_ = fresh.Create(ctx, &domain.Booking{
			UserID: userID, BathhouseID: bhID,
			StartTime: past, EndTime: now,
			Status: domain.BookingConfirmed,
		})
		ci := now
		_ = fresh.Create(ctx, &domain.Booking{
			UserID: userID, BathhouseID: bhID,
			StartTime: past, EndTime: now,
			Status:      domain.BookingConfirmed,
			CheckedInAt: &ci,
		})

		noShow, _ := fresh.ListConfirmedWithoutCheckin(ctx, now)
		if len(noShow) != 1 {
			t.Errorf("ListConfirmedWithoutCheckin: got %d, want 1", len(noShow))
		}

		_ = fresh.Create(ctx, &domain.Booking{
			UserID: userID, BathhouseID: bhID,
			StartTime: future, EndTime: future.Add(time.Hour),
			Status: domain.BookingConfirmed,
		})
		upcoming, _ := fresh.ListUpcoming(ctx, now, future.Add(2*time.Hour))
		if len(upcoming) != 1 {
			t.Errorf("ListUpcoming: got %d, want 1", len(upcoming))
		}
	})

	t.Run("UpdateModificationAndEndTime", func(t *testing.T) {
		fresh := NewBookingRepo()
		b := &domain.Booking{
			UserID: userID, BathhouseID: bhID,
			StartTime:  now,
			EndTime:    now.Add(time.Hour),
			TotalPrice: 1000,
			Status:     domain.BookingConfirmed,
		}
		_ = fresh.Create(ctx, b)

		newStart := now.Add(2 * time.Hour)
		newEnd := newStart.Add(2 * time.Hour)
		err := fresh.UpdateModification(ctx, b.ID, newStart, newEnd, 4, 2000, 0, 1500, 0, 0, 0, 200, 1)
		if err != nil {
			t.Fatalf("UpdateModification: %v", err)
		}
		got, _ := fresh.GetByID(ctx, b.ID)
		if got.GuestCount != 4 || got.TotalPrice != 2000 || got.ModificationCount != 1 {
			t.Errorf("UpdateModification mismatch: %+v", got)
		}

		if err := fresh.UpdateModification(ctx, uuid.New(), newStart, newEnd, 1, 1, 0, 0, 0, 0, 0, 0, 0); err != domain.ErrNotFound {
			t.Errorf("UpdateModification missing: got %v", err)
		}

		// UpdateEndTime: success and concurrent-update detection.
		if err := fresh.UpdateEndTime(ctx, b.ID, newEnd, newEnd.Add(time.Hour), 3000); err != nil {
			t.Fatalf("UpdateEndTime: %v", err)
		}
		if err := fresh.UpdateEndTime(ctx, b.ID, newEnd /* stale */, newEnd.Add(2*time.Hour), 4000); err != domain.ErrWalletConcurrentUpdate {
			t.Errorf("UpdateEndTime concurrent: %v", err)
		}
		if err := fresh.UpdateEndTime(ctx, uuid.New(), newEnd, newEnd, 0); err != domain.ErrNotFound {
			t.Errorf("UpdateEndTime missing: %v", err)
		}
	})

	t.Run("RegionAndDeposits", func(t *testing.T) {
		fresh := NewBookingRepo()
		regionBh := uuid.New()
		_ = fresh.Create(ctx, &domain.Booking{
			UserID: userID, BathhouseID: regionBh,
			StartTime: now, EndTime: now.Add(time.Hour),
			Status: domain.BookingConfirmed,
		})
		// Without region map, all confirmed match.
		all, _ := fresh.ListConfirmedByRegionAndDateRange(ctx, "ru", now.Add(-time.Hour), now.Add(2*time.Hour))
		if len(all) != 1 {
			t.Errorf("ListConfirmedByRegionAndDateRange (no map): %d", len(all))
		}
		fresh.SetBathhouseRegion(regionBh, "ru")
		ru, _ := fresh.ListConfirmedByRegionAndDateRange(ctx, "ru", now.Add(-time.Hour), now.Add(2*time.Hour))
		if len(ru) != 1 {
			t.Errorf("ListConfirmedByRegionAndDateRange ru: %d", len(ru))
		}
		by, _ := fresh.ListConfirmedByRegionAndDateRange(ctx, "by", now.Add(-time.Hour), now.Add(2*time.Hour))
		if len(by) != 0 {
			t.Errorf("ListConfirmedByRegionAndDateRange by: %d", len(by))
		}

		// Deposit fields
		b := &domain.Booking{
			UserID: userID, BathhouseID: regionBh,
			Status:     domain.BookingConfirmed,
			StartTime:  now,
			EndTime:    now.Add(time.Hour),
			TotalPrice: 1000,
		}
		_ = fresh.Create(ctx, b)
		if err := fresh.UpdateDeposit(ctx, b.ID, 500, domain.DepositHeld, "ext-1"); err != nil {
			t.Fatalf("UpdateDeposit: %v", err)
		}
		got, _ := fresh.GetByID(ctx, b.ID)
		if got.DepositAmount != 500 || got.DepositStatus != domain.DepositHeld || got.DepositExternalID != "ext-1" {
			t.Errorf("UpdateDeposit mismatch: %+v", got)
		}
		if err := fresh.UpdateDeposit(ctx, uuid.New(), 0, domain.DepositHeld, ""); err != domain.ErrNotFound {
			t.Errorf("UpdateDeposit missing: %v", err)
		}

		// Mark check-out so the deposit becomes ready for release.
		co := now
		_ = fresh.UpdateCheckout(ctx, b.ID, &co, domain.BookingCompleted)
		ready, _ := fresh.ListHeldDepositsReadyForRelease(ctx, now.Add(time.Minute))
		if len(ready) != 1 {
			t.Errorf("ListHeldDepositsReadyForRelease: got %d, want 1", len(ready))
		}

		released := now
		if err := fresh.UpdateDepositStatus(ctx, b.ID, domain.DepositReleased, &released); err != nil {
			t.Fatalf("UpdateDepositStatus: %v", err)
		}
		if err := fresh.UpdateDepositStatus(ctx, uuid.New(), domain.DepositReleased, &released); err != domain.ErrNotFound {
			t.Errorf("UpdateDepositStatus missing: %v", err)
		}
	})

	t.Run("OwnerCancellationsAndResponseStats", func(t *testing.T) {
		fresh := NewBookingRepo()
		oBh := uuid.New()
		oneHour := time.Hour

		// 2 owner-cancelled bookings within window, 1 outside window.
		within := now.Add(-time.Hour)
		outside := now.Add(-72 * time.Hour)
		bIn := &domain.Booking{
			UserID: userID, BathhouseID: oBh,
			Status:           domain.BookingCancelled,
			CancelledByOwner: true,
			StartTime:        within, EndTime: within.Add(oneHour),
			TotalPrice: 1000,
			UpdatedAt:  within,
		}
		bOut := &domain.Booking{
			UserID: userID, BathhouseID: oBh,
			Status:           domain.BookingCancelled,
			CancelledByOwner: true,
			StartTime:        outside, EndTime: outside.Add(oneHour),
			TotalPrice: 1000,
			UpdatedAt:  outside,
		}
		_ = fresh.Create(ctx, bIn)
		_ = fresh.Create(ctx, bOut)

		count, err := fresh.CountOwnerCancellations(ctx, uuid.New(), now.Add(-2*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("CountOwnerCancellations: got %d, want 1", count)
		}

		// Response stats: one held request that got confirmed.
		hold := uuid.New()
		req := &domain.Booking{
			UserID: userID, BathhouseID: oBh,
			Status:    domain.BookingConfirmed,
			HoldID:    &hold,
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now.Add(-time.Hour + 30*time.Minute),
		}
		_ = fresh.Create(ctx, req)
		// Force CreatedAt because Create stamps it; reach in via Update.
		req.CreatedAt = now.Add(-time.Hour)
		req.UpdatedAt = now.Add(-time.Hour + 30*time.Minute)
		_ = fresh.Update(ctx, req)

		total, responded, avg, err := fresh.GetResponseStats(ctx, oBh, now.Add(-2*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if total != 1 || responded != 1 {
			t.Errorf("GetResponseStats counts: total=%d responded=%d", total, responded)
		}
		if avg <= 0 {
			t.Errorf("GetResponseStats avg should be positive, got %d", avg)
		}
	})

	t.Run("ListCompletedForReviewRequestsAndLastBooking", func(t *testing.T) {
		fresh := NewBookingRepo()
		uID := uuid.New()
		bh := uuid.New()
		ago := now.Add(-2 * time.Hour)
		recent := now
		_ = fresh.Create(ctx, &domain.Booking{
			UserID: uID, BathhouseID: bh,
			Status:       domain.BookingCompleted,
			CheckedOutAt: &ago,
			StartTime:    ago, EndTime: ago.Add(time.Hour),
		})
		_ = fresh.Create(ctx, &domain.Booking{
			UserID: uID, BathhouseID: bh,
			Status:       domain.BookingCompleted,
			CheckedOutAt: &recent,
			StartTime:    recent, EndTime: recent.Add(time.Hour),
		})

		old, _ := fresh.ListCompletedForReviewRequests(ctx, now.Add(-time.Hour))
		if len(old) != 1 {
			t.Errorf("ListCompletedForReviewRequests: got %d, want 1", len(old))
		}

		latest, err := fresh.GetLastBookingDateByUser(ctx, uID)
		if err != nil {
			t.Fatal(err)
		}
		if latest == nil {
			t.Fatal("GetLastBookingDateByUser returned nil")
		}

		none, _ := fresh.GetLastBookingDateByUser(ctx, uuid.New())
		if none != nil {
			t.Error("GetLastBookingDateByUser for unknown user should be nil")
		}
	})

	t.Run("ListTimedOutRequests", func(t *testing.T) {
		fresh := NewBookingRepo()

		old := &domain.Booking{
			UserID: userID, BathhouseID: bhID,
			Status:    domain.BookingPendingOwner,
			CreatedAt: time.Now().Add(-48 * time.Hour),
			StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		}
		_ = fresh.Create(ctx, old)
		// Re-stamp CreatedAt, since Create resets if zero.
		old.CreatedAt = time.Now().Add(-48 * time.Hour)
		_ = fresh.Update(ctx, old)

		newReq := &domain.Booking{
			UserID: userID, BathhouseID: bhID,
			Status:    domain.BookingPendingOwner,
			StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		}
		_ = fresh.Create(ctx, newReq)

		timed, err := fresh.ListTimedOutRequests(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(timed) != 1 {
			t.Errorf("ListTimedOutRequests: got %d, want 1", len(timed))
		}
	})
}
