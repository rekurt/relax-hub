package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// TestGuestCardRepo_FullCoverage exercises every GuestCardRepo method.
func TestGuestCardRepo_FullCoverage(t *testing.T) {
	ctx := context.Background()

	t.Run("UpsertCreatesAndUpdates", func(t *testing.T) {
		repo := NewGuestCardRepo()
		ownerID := uuid.New()
		clientID := uuid.New()
		bathhouseID := uuid.New()

		first := &domain.GuestCard{
			OwnerID:     ownerID,
			ClientID:    clientID,
			BathhouseID: bathhouseID,
			TotalSpent:  100000,
		}
		if err := repo.Upsert(ctx, first); err != nil {
			t.Fatal(err)
		}

		got, err := repo.GetByID(ctx, first.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.OwnerID != ownerID {
			t.Errorf("OwnerID mismatch")
		}

		// Second upsert with same triple should aggregate.
		second := &domain.GuestCard{
			OwnerID:     ownerID,
			ClientID:    clientID,
			BathhouseID: bathhouseID,
			TotalSpent:  50000,
			LastVisitAt: time.Now().Add(time.Hour),
		}
		_ = repo.Upsert(ctx, second)

		got2, _ := repo.GetByOwnerAndClient(ctx, ownerID, clientID, bathhouseID)
		if got2.TotalSpent != 150000 {
			t.Errorf("TotalSpent: got %d, want 150000", got2.TotalSpent)
		}
		if got2.VisitCount != 1 {
			t.Errorf("VisitCount: got %d, want 1 (started at 0, one upsert increment)", got2.VisitCount)
		}
		if got2.AvgCheck != got2.TotalSpent/int64(got2.VisitCount) {
			t.Errorf("AvgCheck inconsistent with TotalSpent/VisitCount")
		}

		// Missing IDs
		if _, err := repo.GetByID(ctx, uuid.New()); err != domain.ErrNotFound {
			t.Errorf("GetByID missing: got %v", err)
		}
		if _, err := repo.GetByOwnerAndClient(ctx, uuid.New(), uuid.New(), uuid.New()); err != domain.ErrNotFound {
			t.Errorf("GetByOwnerAndClient missing: got %v", err)
		}
	})

	t.Run("ListByOwnerWithFiltersAndSorts", func(t *testing.T) {
		repo := NewGuestCardRepo()
		ownerID := uuid.New()
		bh1 := uuid.New()
		bh2 := uuid.New()
		now := time.Now()

		makeCard := func(spent int64, visits int, tags []string, last time.Time, otherOwner bool, bh uuid.UUID) *domain.GuestCard {
			oID := ownerID
			if otherOwner {
				oID = uuid.New()
			}
			return &domain.GuestCard{
				OwnerID:     oID,
				ClientID:    uuid.New(),
				BathhouseID: bh,
				TotalSpent:  spent,
				VisitCount:  visits,
				AvgCheck:    spent / int64(maxInt(1, visits)),
				Tags:        tags,
				LastVisitAt: last,
			}
		}

		_ = repo.Upsert(ctx, makeCard(10000000, 5, []string{"vip"}, now, false, bh1))
		_ = repo.Upsert(ctx, makeCard(2000000, 1, []string{"new"}, now.Add(-100*24*time.Hour), false, bh1))
		_ = repo.Upsert(ctx, makeCard(500000, 3, nil, now.Add(-30*24*time.Hour), false, bh2))
		_ = repo.Upsert(ctx, makeCard(999999, 10, nil, now, true, bh1)) // other owner

		base := domain.GuestCardFilter{OwnerID: ownerID, Page: 1, PageSize: 10}
		page, _ := repo.ListByOwner(ctx, base)
		if page.TotalCount != 3 {
			t.Errorf("ListByOwner: got %d, want 3", page.TotalCount)
		}

		// Filter by bathhouse
		bhFilter := bh1
		page, _ = repo.ListByOwner(ctx, domain.GuestCardFilter{OwnerID: ownerID, BathhouseID: &bhFilter, Page: 1, PageSize: 10})
		if page.TotalCount != 2 {
			t.Errorf("Bathhouse filter: got %d, want 2", page.TotalCount)
		}

		// Filter by segment VIP
		seg := domain.SegmentVIP
		page, _ = repo.ListByOwner(ctx, domain.GuestCardFilter{OwnerID: ownerID, Segment: &seg, Page: 1, PageSize: 10})
		if page.TotalCount != 1 {
			t.Errorf("VIP segment: got %d, want 1", page.TotalCount)
		}

		// Filter by lost segment
		lost := domain.SegmentLost
		page, _ = repo.ListByOwner(ctx, domain.GuestCardFilter{OwnerID: ownerID, Segment: &lost, Page: 1, PageSize: 10})
		if page.TotalCount != 1 {
			t.Errorf("Lost segment: got %d, want 1", page.TotalCount)
		}

		// Filter by tag
		vipTag := "vip"
		page, _ = repo.ListByOwner(ctx, domain.GuestCardFilter{OwnerID: ownerID, Tag: &vipTag, Page: 1, PageSize: 10})
		if page.TotalCount != 1 {
			t.Errorf("Tag filter: got %d, want 1", page.TotalCount)
		}

		missingTag := "missing"
		page, _ = repo.ListByOwner(ctx, domain.GuestCardFilter{OwnerID: ownerID, Tag: &missingTag, Page: 1, PageSize: 10})
		if page.TotalCount != 0 {
			t.Errorf("Missing tag: got %d, want 0", page.TotalCount)
		}

		// Date filters
		from := now.Add(-50 * 24 * time.Hour)
		page, _ = repo.ListByOwner(ctx, domain.GuestCardFilter{OwnerID: ownerID, DateFrom: &from, Page: 1, PageSize: 10})
		if page.TotalCount != 2 {
			t.Errorf("DateFrom: got %d, want 2", page.TotalCount)
		}
		to := now.Add(-10 * 24 * time.Hour)
		page, _ = repo.ListByOwner(ctx, domain.GuestCardFilter{OwnerID: ownerID, DateTo: &to, Page: 1, PageSize: 10})
		if page.TotalCount != 2 {
			t.Errorf("DateTo: got %d, want 2", page.TotalCount)
		}

		// Sort variants — just ensure no panic and rows are returned.
		for _, sort := range []string{"total_spent", "visit_count", "avg_check", "last_visit"} {
			f := base
			f.SortBy = sort
			res, _ := repo.ListByOwner(ctx, f)
			if res.TotalCount == 0 {
				t.Errorf("sort %q produced no results", sort)
			}
		}
	})

	t.Run("UpdateNotes", func(t *testing.T) {
		repo := NewGuestCardRepo()
		card := &domain.GuestCard{
			OwnerID:     uuid.New(),
			ClientID:    uuid.New(),
			BathhouseID: uuid.New(),
			Tags:        []string{"old"},
		}
		_ = repo.Upsert(ctx, card)

		if err := repo.UpdateNotes(ctx, card.ID, "vip client", []string{"vip", "regular"}); err != nil {
			t.Fatalf("UpdateNotes: %v", err)
		}
		got, _ := repo.GetByID(ctx, card.ID)
		if got.Notes != "vip client" {
			t.Errorf("Notes: got %q", got.Notes)
		}
		if len(got.Tags) != 2 {
			t.Errorf("Tags: got %v, want 2 entries", got.Tags)
		}

		if err := repo.UpdateNotes(ctx, uuid.New(), "x", nil); err != domain.ErrNotFound {
			t.Errorf("UpdateNotes missing: got %v", err)
		}
	})

	t.Run("CountBySegmentAndStats", func(t *testing.T) {
		repo := NewGuestCardRepo()
		ownerID := uuid.New()
		bh := uuid.New()

		_ = repo.Upsert(ctx, &domain.GuestCard{OwnerID: ownerID, ClientID: uuid.New(), BathhouseID: bh, VisitCount: 1, TotalSpent: 1000})
		_ = repo.Upsert(ctx, &domain.GuestCard{OwnerID: ownerID, ClientID: uuid.New(), BathhouseID: bh, VisitCount: 5, TotalSpent: 9000000})
		_ = repo.Upsert(ctx, &domain.GuestCard{OwnerID: uuid.New(), ClientID: uuid.New(), BathhouseID: uuid.New(), VisitCount: 1})

		filter := domain.GuestCardFilter{OwnerID: ownerID}

		count, err := repo.CountBySegment(ctx, filter, domain.SegmentNew)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("CountBySegment new: got %d, want 1", count)
		}

		count, _ = repo.CountBySegment(ctx, filter, domain.SegmentVIP)
		if count != 1 {
			t.Errorf("CountBySegment vip: got %d, want 1", count)
		}

		// NoOwnerFilter aggregates everything.
		count, _ = repo.CountBySegment(ctx, domain.GuestCardFilter{NoOwnerFilter: true}, domain.SegmentNew)
		if count != 2 {
			t.Errorf("CountBySegment unfiltered: got %d, want 2", count)
		}

		// BathhouseIDs filter
		count, _ = repo.CountBySegment(ctx, domain.GuestCardFilter{BathhouseIDs: []uuid.UUID{bh}}, domain.SegmentNew)
		if count != 1 {
			t.Errorf("CountBySegment by ids: got %d, want 1", count)
		}
		count, _ = repo.CountBySegment(ctx, domain.GuestCardFilter{BathhouseIDs: []uuid.UUID{uuid.New()}}, domain.SegmentNew)
		if count != 0 {
			t.Errorf("CountBySegment by unknown ids: got %d, want 0", count)
		}

		stats, err := repo.GetStats(ctx, filter)
		if err != nil {
			t.Fatal(err)
		}
		if stats.TotalGuests != 2 {
			t.Errorf("TotalGuests: got %d, want 2", stats.TotalGuests)
		}
		if stats.AvgVisitCount != 3 {
			t.Errorf("AvgVisitCount: got %d, want 3", stats.AvgVisitCount)
		}

		emptyStats, _ := repo.GetStats(ctx, domain.GuestCardFilter{OwnerID: uuid.New()})
		if emptyStats.TotalGuests != 0 {
			t.Errorf("Empty stats: got %v", emptyStats)
		}
	})

	t.Run("GetRFMScores", func(t *testing.T) {
		repo := NewGuestCardRepo()
		empty, err := repo.GetRFMScores(ctx, domain.GuestCardFilter{NoOwnerFilter: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(empty.Guests) != 0 || len(empty.Matrix) != 0 {
			t.Error("Expected empty RFM result")
		}

		now := time.Now()
		ownerID := uuid.New()
		for i := 0; i < 6; i++ {
			_ = repo.Upsert(ctx, &domain.GuestCard{
				OwnerID:     ownerID,
				ClientID:    uuid.New(),
				BathhouseID: uuid.New(),
				VisitCount:  i + 1,
				TotalSpent:  int64(1000 * (i + 1)),
				LastVisitAt: now.Add(time.Duration(i) * time.Hour),
			})
		}

		result, err := repo.GetRFMScores(ctx, domain.GuestCardFilter{OwnerID: ownerID})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Guests) != 6 {
			t.Errorf("Guests: got %d, want 6", len(result.Guests))
		}
		if len(result.Matrix) == 0 {
			t.Error("Matrix should not be empty")
		}
		for _, g := range result.Guests {
			if g.RFM.Recency < 1 || g.RFM.Recency > 5 {
				t.Errorf("Recency out of bounds: %d", g.RFM.Recency)
			}
			if g.RFM.Frequency < 1 || g.RFM.Frequency > 5 {
				t.Errorf("Frequency out of bounds: %d", g.RFM.Frequency)
			}
			if g.RFM.Monetary < 1 || g.RFM.Monetary > 5 {
				t.Errorf("Monetary out of bounds: %d", g.RFM.Monetary)
			}
		}
	})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
