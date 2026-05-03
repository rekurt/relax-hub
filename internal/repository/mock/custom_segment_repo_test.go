package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func TestCustomSegmentRepo_FullCoverage(t *testing.T) {
	ctx := context.Background()

	gcRepoIface := NewGuestCardRepo()
	gcRepo := gcRepoIface.(*GuestCardRepo)
	repo := NewCustomSegmentRepo(gcRepo)

	ownerID := uuid.New()
	bathhouseID := uuid.New()

	t.Run("CRUD", func(t *testing.T) {
		segment := &domain.CustomSegment{
			OwnerID: ownerID,
			Name:    "VIPs",
		}
		if err := repo.Create(ctx, segment); err != nil {
			t.Fatal(err)
		}
		got, err := repo.GetByID(ctx, segment.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "VIPs" {
			t.Errorf("Name: got %q", got.Name)
		}
		if _, err := repo.GetByID(ctx, uuid.New()); err != domain.ErrNotFound {
			t.Errorf("GetByID missing: got %v", err)
		}

		got.Name = "VIP+"
		if err := repo.Update(ctx, got); err != nil {
			t.Fatalf("Update: %v", err)
		}

		ghost := &domain.CustomSegment{ID: uuid.New(), Name: "ghost"}
		if err := repo.Update(ctx, ghost); err != domain.ErrNotFound {
			t.Errorf("Update missing: %v", err)
		}

		if err := repo.Delete(ctx, segment.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if err := repo.Delete(ctx, segment.ID); err != domain.ErrNotFound {
			t.Errorf("Delete already-deleted: %v", err)
		}
	})

	t.Run("ListByOwner", func(t *testing.T) {
		s1 := &domain.CustomSegment{OwnerID: ownerID, Name: "B"}
		s2 := &domain.CustomSegment{OwnerID: ownerID, Name: "A", BathhouseID: &bathhouseID}
		s3 := &domain.CustomSegment{OwnerID: uuid.New(), Name: "C"}
		_ = repo.Create(ctx, s1)
		_ = repo.Create(ctx, s2)
		_ = repo.Create(ctx, s3)

		all, err := repo.ListByOwner(ctx, domain.CustomSegmentFilter{OwnerID: ownerID})
		if err != nil {
			t.Fatal(err)
		}
		if len(all) != 2 {
			t.Errorf("ListByOwner: got %d, want 2", len(all))
		}
		if all[0].Name != "A" {
			t.Errorf("Sort by name: first item %q", all[0].Name)
		}

		// BathhouseID filter excludes segments with non-matching bathhouse.
		bhFilter := bathhouseID
		filtered, _ := repo.ListByOwner(ctx, domain.CustomSegmentFilter{OwnerID: ownerID, BathhouseID: &bhFilter})
		if len(filtered) != 2 {
			t.Errorf("Bathhouse filter: got %d, want 2 (segment without bathhouse always matches)", len(filtered))
		}

		other := uuid.New()
		filtered, _ = repo.ListByOwner(ctx, domain.CustomSegmentFilter{OwnerID: ownerID, BathhouseID: &other})
		if len(filtered) != 1 {
			t.Errorf("Other bathhouse filter: got %d, want 1", len(filtered))
		}
	})

	t.Run("EvaluateAndCount", func(t *testing.T) {
		// Seed guest cards
		now := time.Now()
		_ = gcRepo.Upsert(ctx, &domain.GuestCard{
			OwnerID: ownerID, ClientID: uuid.New(), BathhouseID: bathhouseID,
			VisitCount: 5, AvgCheck: 200000, TotalSpent: 1000000,
			Tags: []string{"vip", "regular"}, LastVisitAt: now,
		})
		_ = gcRepo.Upsert(ctx, &domain.GuestCard{
			OwnerID: ownerID, ClientID: uuid.New(), BathhouseID: bathhouseID,
			VisitCount: 1, AvgCheck: 50000, TotalSpent: 50000,
			Tags: []string{"new"}, LastVisitAt: now.Add(-100 * 24 * time.Hour),
		})
		_ = gcRepo.Upsert(ctx, &domain.GuestCard{
			OwnerID: uuid.New(), ClientID: uuid.New(), BathhouseID: uuid.New(),
			VisitCount: 99, TotalSpent: 9999999,
		})

		minVisit := 3
		minSpent := int64(500000)
		maxVisit := 100
		segment := &domain.CustomSegment{
			OwnerID:     ownerID,
			BathhouseID: &bathhouseID,
			Name:        "Reactivation",
			Conditions: domain.CustomSegmentCondition{
				VisitCountMin: &minVisit,
				VisitCountMax: &maxVisit,
				TotalSpentMin: &minSpent,
				TagsInclude:   []string{"vip"},
				TagsExclude:   []string{"banned"},
			},
		}
		ownerFilter := domain.GuestCardFilter{OwnerID: ownerID}

		page, err := repo.EvaluateSegment(ctx, segment, ownerFilter, 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		if page.TotalCount != 1 {
			t.Errorf("EvaluateSegment: got %d, want 1", page.TotalCount)
		}

		count, err := repo.CountSegmentGuests(ctx, segment, ownerFilter)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("CountSegmentGuests: got %d, want 1", count)
		}

		// All bounds (max checks) and last_visit_days
		minDays := 0
		maxDays := 1000
		minVisit2 := 1
		zero := int64(0)
		hugeAvg := int64(99999999)
		hugeSpent := int64(99999999)
		bigVisit := 1000
		segWide := &domain.CustomSegment{
			OwnerID: ownerID,
			Name:    "Wide",
			Conditions: domain.CustomSegmentCondition{
				VisitCountMin:    &minVisit2,
				VisitCountMax:    &bigVisit,
				AvgCheckMin:      &zero,
				AvgCheckMax:      &hugeAvg,
				TotalSpentMin:    &zero,
				TotalSpentMax:    &hugeSpent,
				LastVisitDaysMin: &minDays,
				LastVisitDaysMax: &maxDays,
			},
		}
		count, _ = repo.CountSegmentGuests(ctx, segWide, ownerFilter)
		if count != 2 {
			t.Errorf("Wide: got %d, want 2", count)
		}

		// Excluded tag drops the vip guest.
		excludeSeg := &domain.CustomSegment{
			OwnerID: ownerID,
			Name:    "ExcludeVIP",
			Conditions: domain.CustomSegmentCondition{
				TagsExclude: []string{"vip"},
			},
		}
		count, _ = repo.CountSegmentGuests(ctx, excludeSeg, ownerFilter)
		if count != 1 {
			t.Errorf("ExcludeVIP: got %d, want 1", count)
		}

		// Min/max visit-days mismatch
		hi := 10
		lo := 0
		strictRecent := &domain.CustomSegment{
			OwnerID: ownerID,
			Name:    "StrictRecent",
			Conditions: domain.CustomSegmentCondition{
				LastVisitDaysMin: &lo,
				LastVisitDaysMax: &hi,
			},
		}
		count, _ = repo.CountSegmentGuests(ctx, strictRecent, ownerFilter)
		if count != 1 {
			t.Errorf("StrictRecent: got %d, want 1", count)
		}

		// Visit-count rejection
		impossible := 1000
		zeroM := 0
		segReject := &domain.CustomSegment{
			OwnerID: ownerID,
			Name:    "Reject",
			Conditions: domain.CustomSegmentCondition{
				VisitCountMin: &impossible,
				VisitCountMax: &zeroM,
			},
		}
		count, _ = repo.CountSegmentGuests(ctx, segReject, ownerFilter)
		if count != 0 {
			t.Errorf("Reject: got %d, want 0", count)
		}

		// AvgCheck rejection
		bigAvg := int64(99999999999)
		segAvg := &domain.CustomSegment{
			OwnerID: ownerID,
			Name:    "Avg",
			Conditions: domain.CustomSegmentCondition{
				AvgCheckMin: &bigAvg,
			},
		}
		count, _ = repo.CountSegmentGuests(ctx, segAvg, ownerFilter)
		if count != 0 {
			t.Errorf("Avg: got %d, want 0", count)
		}

		// TotalSpent rejection (max)
		hardCap := int64(0)
		segCap := &domain.CustomSegment{
			OwnerID: ownerID,
			Name:    "Cap",
			Conditions: domain.CustomSegmentCondition{
				TotalSpentMax: &hardCap,
			},
		}
		count, _ = repo.CountSegmentGuests(ctx, segCap, ownerFilter)
		if count != 0 {
			t.Errorf("Cap: got %d, want 0", count)
		}

		// TagsInclude rejection
		mustInclude := &domain.CustomSegment{
			OwnerID: ownerID,
			Name:    "MustHave",
			Conditions: domain.CustomSegmentCondition{
				TagsInclude: []string{"missing-tag"},
			},
		}
		count, _ = repo.CountSegmentGuests(ctx, mustInclude, ownerFilter)
		if count != 0 {
			t.Errorf("MustHave: got %d, want 0", count)
		}
	})
}
