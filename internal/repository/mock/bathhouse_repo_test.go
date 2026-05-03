package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

func TestBathhouseRepo_FullCoverage(t *testing.T) {
	ctx := context.Background()
	repo := NewBathhouseRepo()

	ownerID := uuid.New()

	mkBh := func(name, slug string, status domain.BathhouseStatus) *domain.Bathhouse {
		return &domain.Bathhouse{
			OwnerID:      ownerID,
			Name:         name,
			Slug:         slug,
			Description:  "desc " + name,
			Address:      "addr " + name,
			CityID:       1,
			PricePerHour: 200000,
			MaxGuests:    6,
			Status:       status,
			BookingMode:  domain.BookingModeInstant,
			ApiKey:       "key-" + slug,
		}
	}

	t.Run("CreateGetUpdateDelete", func(t *testing.T) {
		bh := mkBh("Sauna", "sauna", domain.BathhouseStatusActive)
		if err := repo.Create(ctx, bh); err != nil {
			t.Fatal(err)
		}
		got, err := repo.GetByID(ctx, bh.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != bh.Name {
			t.Errorf("Name: got %q", got.Name)
		}
		if _, err := repo.GetByID(ctx, uuid.New()); err != domain.ErrNotFound {
			t.Errorf("GetByID missing: %v", err)
		}

		got.Description = "updated"
		if err := repo.Update(ctx, got); err != nil {
			t.Fatal(err)
		}
		ghost := &domain.Bathhouse{ID: uuid.New()}
		if err := repo.Update(ctx, ghost); err != domain.ErrNotFound {
			t.Errorf("Update missing: %v", err)
		}

		// Lookups
		bySlug, err := repo.GetBySlug(ctx, "sauna")
		if err != nil {
			t.Fatalf("GetBySlug: %v", err)
		}
		if bySlug.ID != bh.ID {
			t.Errorf("GetBySlug mismatch")
		}
		if _, err := repo.GetBySlug(ctx, "nope"); err != domain.ErrNotFound {
			t.Errorf("GetBySlug missing: %v", err)
		}

		exists, _ := repo.SlugExists(ctx, "sauna")
		if !exists {
			t.Error("SlugExists should return true")
		}
		exists, _ = repo.SlugExists(ctx, "missing")
		if exists {
			t.Error("SlugExists missing should be false")
		}

		byApi, err := repo.GetByAPIKey(ctx, "key-sauna")
		if err != nil {
			t.Fatal(err)
		}
		if byApi.ID != bh.ID {
			t.Errorf("GetByAPIKey mismatch")
		}
		if _, err := repo.GetByAPIKey(ctx, "no-such-key"); err != domain.ErrNotFound {
			t.Errorf("GetByAPIKey missing: %v", err)
		}

		// Delete
		if err := repo.Delete(ctx, bh.ID); err != nil {
			t.Fatal(err)
		}
		if err := repo.Delete(ctx, bh.ID); err != domain.ErrNotFound {
			t.Errorf("Delete twice: %v", err)
		}
	})

	t.Run("ListWithFilters", func(t *testing.T) {
		fresh := NewBathhouseRepo()
		_ = fresh.Create(ctx, mkBh("Active 1", "a1", domain.BathhouseStatusActive))
		_ = fresh.Create(ctx, mkBh("Active 2", "a2", domain.BathhouseStatusActive))
		_ = fresh.Create(ctx, mkBh("Pending", "p1", domain.BathhouseStatusPending))

		// Default: only active
		page, err := fresh.List(ctx, domain.BathhouseFilter{Page: 1, PageSize: 10})
		if err != nil {
			t.Fatal(err)
		}
		if page.TotalCount != 2 {
			t.Errorf("Default list: %d, want 2", page.TotalCount)
		}

		// ShowAllStatuses
		page, _ = fresh.List(ctx, domain.BathhouseFilter{ShowAllStatuses: true, Page: 1, PageSize: 10})
		if page.TotalCount != 3 {
			t.Errorf("ShowAllStatuses: %d", page.TotalCount)
		}

		// Status filter
		st := domain.BathhouseStatusPending
		page, _ = fresh.List(ctx, domain.BathhouseFilter{Status: &st})
		if page.TotalCount != 1 {
			t.Errorf("Status pending: %d", page.TotalCount)
		}

		// CityID filter
		other := int64(999)
		page, _ = fresh.List(ctx, domain.BathhouseFilter{CityID: &other})
		if page.TotalCount != 0 {
			t.Errorf("Other city: %d", page.TotalCount)
		}

		// PriceMin / PriceMax
		min := int64(100000)
		max := int64(500000)
		page, _ = fresh.List(ctx, domain.BathhouseFilter{PriceMin: &min, PriceMax: &max})
		if page.TotalCount != 2 {
			t.Errorf("Price range: %d", page.TotalCount)
		}
		impossibleMin := int64(99999999)
		page, _ = fresh.List(ctx, domain.BathhouseFilter{PriceMin: &impossibleMin})
		if page.TotalCount != 0 {
			t.Errorf("Impossible min: %d", page.TotalCount)
		}
		zeroMax := int64(0)
		page, _ = fresh.List(ctx, domain.BathhouseFilter{PriceMax: &zeroMax})
		if page.TotalCount != 0 {
			t.Errorf("Zero max: %d", page.TotalCount)
		}

		// GuestCount
		need := 100
		page, _ = fresh.List(ctx, domain.BathhouseFilter{GuestCount: &need})
		if page.TotalCount != 0 {
			t.Errorf("GuestCount: %d", page.TotalCount)
		}

		// SearchQuery
		q := "Active"
		page, _ = fresh.List(ctx, domain.BathhouseFilter{SearchQuery: &q})
		if page.TotalCount != 2 {
			t.Errorf("Search by Active: %d, want 2", page.TotalCount)
		}
		// Prefix matching
		qPrefix := "Activ"
		page, _ = fresh.List(ctx, domain.BathhouseFilter{SearchQuery: &qPrefix})
		if page.TotalCount != 2 {
			t.Errorf("Prefix search: %d, want 2", page.TotalCount)
		}
		qNone := "xyzzy"
		page, _ = fresh.List(ctx, domain.BathhouseFilter{SearchQuery: &qNone})
		if page.TotalCount != 0 {
			t.Errorf("No-match search: %d", page.TotalCount)
		}

		// OpenNow filter — bathhouse without working hours is closed
		openNow := true
		page, _ = fresh.List(ctx, domain.BathhouseFilter{OpenNow: &openNow})
		if page.TotalCount != 0 {
			t.Errorf("OpenNow w/o hours: %d", page.TotalCount)
		}
	})

	t.Run("ListByOwnerAndIDsByOwner", func(t *testing.T) {
		fresh := NewBathhouseRepo()
		oID := uuid.New()
		for i := 0; i < 2; i++ {
			bh := mkBh("X", "x", domain.BathhouseStatusActive)
			bh.OwnerID = oID
			bh.Slug = "x" + string(rune('a'+i))
			_ = fresh.Create(ctx, bh)
		}
		_ = fresh.Create(ctx, mkBh("Other", "other", domain.BathhouseStatusActive))

		page, _ := fresh.ListByOwner(ctx, oID, 1, 10)
		if page.TotalCount != 2 {
			t.Errorf("ListByOwner: %d", page.TotalCount)
		}

		ids, _ := fresh.ListIDsByOwner(ctx, oID)
		if len(ids) != 2 {
			t.Errorf("ListIDsByOwner: %d", len(ids))
		}
	})

	t.Run("RatingsAndStatusUpdates", func(t *testing.T) {
		fresh := NewBathhouseRepo()
		bh := mkBh("R", "rating", domain.BathhouseStatusActive)
		_ = fresh.Create(ctx, bh)

		if err := fresh.UpdateRating(ctx, bh.ID); err != nil {
			t.Errorf("UpdateRating: %v", err)
		}

		if err := fresh.UpdateBayesianRating(ctx, bh.ID, 4.5); err != nil {
			t.Fatal(err)
		}
		got, _ := fresh.GetByID(ctx, bh.ID)
		if got.BayesianRating != 4.5 {
			t.Errorf("BayesianRating: %v", got.BayesianRating)
		}
		if err := fresh.UpdateBayesianRating(ctx, uuid.New(), 0); err != domain.ErrNotFound {
			t.Errorf("UpdateBayesianRating missing: %v", err)
		}

		if err := fresh.UpdateStatus(ctx, bh.ID, domain.BathhouseStatusInactive); err != nil {
			t.Fatal(err)
		}
		if err := fresh.UpdateStatus(ctx, uuid.New(), domain.BathhouseStatusInactive); err != domain.ErrNotFound {
			t.Errorf("UpdateStatus missing: %v", err)
		}

		if err := fresh.UpdatePhotoVerified(ctx, bh.ID, true); err != nil {
			t.Fatal(err)
		}
		got, _ = fresh.GetByID(ctx, bh.ID)
		if !got.IsPhotoVerified {
			t.Error("IsPhotoVerified not set")
		}
		if err := fresh.UpdatePhotoVerified(ctx, uuid.New(), true); err != domain.ErrNotFound {
			t.Errorf("UpdatePhotoVerified missing: %v", err)
		}

		if err := fresh.IncrementViewCount(ctx, bh.ID); err != nil {
			t.Fatal(err)
		}
		if err := fresh.IncrementViewCount(ctx, uuid.New()); err != domain.ErrNotFound {
			t.Errorf("IncrementViewCount missing: %v", err)
		}

		if err := fresh.UpdateRankingFields(ctx, bh.ID, 0.6, 0.8); err != nil {
			t.Fatal(err)
		}
		if err := fresh.UpdateRankingFields(ctx, uuid.New(), 0, 0); err != domain.ErrNotFound {
			t.Errorf("UpdateRankingFields missing: %v", err)
		}

		when := time.Now()
		if err := fresh.UpdateResponseRate(ctx, bh.ID, 0.5, 30, &when); err != nil {
			t.Fatal(err)
		}
		if err := fresh.UpdateResponseRate(ctx, uuid.New(), 0, 0, nil); err != domain.ErrNotFound {
			t.Errorf("UpdateResponseRate missing: %v", err)
		}
	})

	t.Run("CalendarTokensAndSuggest", func(t *testing.T) {
		fresh := NewBathhouseRepo()
		bh := mkBh("Cal", "cal", domain.BathhouseStatusActive)
		_ = fresh.Create(ctx, bh)

		_, err := fresh.GetCalendarToken(ctx, bh.ID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fresh.GetCalendarToken(ctx, uuid.New()); err != domain.ErrNotFound {
			t.Errorf("GetCalendarToken missing: %v", err)
		}

		if err := fresh.SetCalendarToken(ctx, bh.ID, "tok"); err != nil {
			t.Fatal(err)
		}
		if err := fresh.SetCalendarToken(ctx, uuid.New(), "x"); err != domain.ErrNotFound {
			t.Errorf("SetCalendarToken missing: %v", err)
		}
		got, _ := fresh.GetByCalendarToken(ctx, "tok")
		if got.ID != bh.ID {
			t.Errorf("GetByCalendarToken mismatch")
		}
		if _, err := fresh.GetByCalendarToken(ctx, "nope"); err != domain.ErrNotFound {
			t.Errorf("GetByCalendarToken missing: %v", err)
		}

		// SuggestNames
		_ = fresh.Create(ctx, mkBh("Cal Two", "cal2", domain.BathhouseStatusActive))
		_ = fresh.Create(ctx, mkBh("Inactive", "inact", domain.BathhouseStatusInactive))
		names, err := fresh.SuggestNames(ctx, repository.SuggestionFilter{Query: "cal", Limit: 5})
		if err != nil {
			t.Fatal(err)
		}
		if len(names) != 2 {
			t.Errorf("SuggestNames: %v, want 2", names)
		}

		// Limit
		names, _ = fresh.SuggestNames(ctx, repository.SuggestionFilter{Query: "cal", Limit: 1})
		if len(names) != 1 {
			t.Errorf("SuggestNames limit: %v", names)
		}
	})

	t.Run("RequestModeAndAreaAvg", func(t *testing.T) {
		fresh := NewBathhouseRepo()
		req := mkBh("Req", "req", domain.BathhouseStatusActive)
		req.BookingMode = domain.BookingModeRequest
		req.PricePerHour = 100000
		_ = fresh.Create(ctx, req)
		inst := mkBh("Inst", "inst", domain.BathhouseStatusActive)
		inst.PricePerHour = 300000
		_ = fresh.Create(ctx, inst)

		reqList, err := fresh.ListRequestModeBathhouses(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(reqList) != 1 {
			t.Errorf("RequestMode list: %d, want 1", len(reqList))
		}

		avg, err := fresh.GetAreaAvgPrice(ctx, 1, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if avg != 200000 {
			t.Errorf("AreaAvgPrice: %d, want 200000", avg)
		}

		// No bathhouses for unknown city
		none, _ := fresh.GetAreaAvgPrice(ctx, 999, 0, 0)
		if none != 0 {
			t.Errorf("AreaAvgPrice empty: %d", none)
		}
	})

	t.Run("OpenNowVariants", func(t *testing.T) {
		fresh := NewBathhouseRepo()
		bh := mkBh("Open", "open", domain.BathhouseStatusActive)

		// Add normal-hours schedule covering full day, every day.
		for i := 0; i < 7; i++ {
			bh.WorkingHours = append(bh.WorkingHours, domain.WorkingHours{
				DayOfWeek: i,
				OpenTime:  "00:00",
				CloseTime: "23:59",
			})
		}
		_ = fresh.Create(ctx, bh)

		openNow := true
		page, _ := fresh.List(ctx, domain.BathhouseFilter{OpenNow: &openNow})
		if page.TotalCount != 1 {
			t.Errorf("Always-open: %d, want 1", page.TotalCount)
		}

		// Overnight schedule: 22:00 - 06:00 (covers midnight wraparound)
		bh2 := mkBh("Night", "night", domain.BathhouseStatusActive)
		for i := 0; i < 7; i++ {
			bh2.WorkingHours = append(bh2.WorkingHours, domain.WorkingHours{
				DayOfWeek: i,
				OpenTime:  "22:00",
				CloseTime: "06:00",
			})
		}
		_ = fresh.Create(ctx, bh2)
		page, _ = fresh.List(ctx, domain.BathhouseFilter{OpenNow: &openNow})
		// At least bh1 is open; bh2 may or may not be depending on current time.
		if page.TotalCount < 1 {
			t.Errorf("OpenNow with overnight: got %d", page.TotalCount)
		}
	})
}
