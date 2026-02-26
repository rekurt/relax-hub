package mock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// Compile-time interface checks
var (
	_ repository.UserRepository           = (*UserRepo)(nil)
	_ repository.CityRepository           = (*CityRepo)(nil)
	_ repository.BathhouseRepository      = (*BathhouseRepo)(nil)
	_ repository.BookingRepository        = (*BookingRepo)(nil)
	_ repository.ReviewRepository         = (*ReviewRepo)(nil)
	_ repository.RepresentativeRepository = (*RepresentativeRepo)(nil)
)

func TestUserRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepo()

	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Phone:        "+7999000000",
		Role:         domain.RoleClient,
		IsActive:     true,
	}

	// Create
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if user.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// GetByID
	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Email != user.Email {
		t.Errorf("Email = %q, want %q", got.Email, user.Email)
	}

	// GetByEmail
	got, err = repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("ID = %v, want %v", got.ID, user.ID)
	}

	// Duplicate email
	dup := &domain.User{Email: "test@example.com", Name: "Dup", Role: domain.RoleClient}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create duplicate: want ErrAlreadyExists, got %v", err)
	}

	// Update
	user.Name = "Updated Name"
	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = repo.GetByID(ctx, user.ID)
	if got.Name != "Updated Name" {
		t.Errorf("Name after update = %q, want %q", got.Name, "Updated Name")
	}

	// SetActive
	if err := repo.SetActive(ctx, user.ID, false); err != nil {
		t.Fatalf("SetActive: %v", err)
	}
	got, _ = repo.GetByID(ctx, user.ID)
	if got.IsActive {
		t.Error("IsActive should be false after SetActive(false)")
	}

	// Not found
	_, err = repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID not found: want ErrNotFound, got %v", err)
	}

	// List
	result, err := repo.List(ctx, 1, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.TotalCount)
	}
}

func TestCityRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewCityRepo()

	city := &domain.City{Name: "Moscow", Slug: "moscow", Latitude: 55.75, Longitude: 37.62}

	// Create
	if err := repo.Create(ctx, city); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if city.ID == 0 {
		t.Fatal("ID should be assigned after Create")
	}

	// GetByID
	got, err := repo.GetByID(ctx, city.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Moscow" {
		t.Errorf("Name = %q, want %q", got.Name, "Moscow")
	}

	// GetBySlug
	got, err = repo.GetBySlug(ctx, "moscow")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if got.ID != city.ID {
		t.Errorf("ID = %d, want %d", got.ID, city.ID)
	}

	// Duplicate slug
	dup := &domain.City{Name: "Moscow 2", Slug: "moscow"}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create duplicate slug: want ErrAlreadyExists, got %v", err)
	}

	// GetAll
	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("GetAll len = %d, want 1", len(all))
	}

	// Update
	city.Name = "Москва"
	if err := repo.Update(ctx, city); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = repo.GetByID(ctx, city.ID)
	if got.Name != "Москва" {
		t.Errorf("Name after update = %q, want %q", got.Name, "Москва")
	}

	// Delete
	if err := repo.Delete(ctx, city.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = repo.GetByID(ctx, city.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID after delete: want ErrNotFound, got %v", err)
	}

	// Delete not found
	if err := repo.Delete(ctx, 999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete not found: want ErrNotFound, got %v", err)
	}
}

func TestBookingRepo_Availability(t *testing.T) {
	ctx := context.Background()
	repo := NewBookingRepo()

	bhID := uuid.New()
	start := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	booking := &domain.Booking{
		UserID:      uuid.New(),
		BathhouseID: bhID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  2,
		TotalPrice:  5000,
		Status:      domain.BookingConfirmed,
	}
	if err := repo.Create(ctx, booking); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Same time slot should be unavailable
	avail, err := repo.CheckAvailability(ctx, bhID, start, end)
	if err != nil {
		t.Fatalf("CheckAvailability: %v", err)
	}
	if avail {
		t.Error("Expected slot to be unavailable")
	}

	// Overlapping slot
	avail, err = repo.CheckAvailability(ctx, bhID,
		start.Add(-time.Hour), start.Add(time.Hour))
	if err != nil {
		t.Fatalf("CheckAvailability overlap: %v", err)
	}
	if avail {
		t.Error("Expected overlapping slot to be unavailable")
	}

	// Non-overlapping slot
	avail, err = repo.CheckAvailability(ctx, bhID,
		end, end.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("CheckAvailability no overlap: %v", err)
	}
	if !avail {
		t.Error("Expected non-overlapping slot to be available")
	}

	// Different bathhouse
	avail, err = repo.CheckAvailability(ctx, uuid.New(), start, end)
	if err != nil {
		t.Fatalf("CheckAvailability different bh: %v", err)
	}
	if !avail {
		t.Error("Expected different bathhouse to be available")
	}

	// GetOverlapping
	overlapping, err := repo.GetOverlapping(ctx, bhID, start.Add(-time.Hour), end.Add(time.Hour))
	if err != nil {
		t.Fatalf("GetOverlapping: %v", err)
	}
	if len(overlapping) != 1 {
		t.Errorf("GetOverlapping len = %d, want 1", len(overlapping))
	}

	// Cancelled booking should not block availability
	if err := repo.UpdateStatus(ctx, booking.ID, domain.BookingCancelled); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	avail, err = repo.CheckAvailability(ctx, bhID, start, end)
	if err != nil {
		t.Fatalf("CheckAvailability after cancel: %v", err)
	}
	if !avail {
		t.Error("Expected cancelled booking to not block availability")
	}
}

func TestBookingRepo_ListByUser(t *testing.T) {
	ctx := context.Background()
	repo := NewBookingRepo()

	userID := uuid.New()
	bhID := uuid.New()

	for i := range 3 {
		b := &domain.Booking{
			UserID:      userID,
			BathhouseID: bhID,
			StartTime:   time.Now().Add(time.Duration(i) * time.Hour),
			EndTime:     time.Now().Add(time.Duration(i+1) * time.Hour),
			GuestCount:  2,
			TotalPrice:  5000,
			Status:      domain.BookingPending,
		}
		if err := repo.Create(ctx, b); err != nil {
			t.Fatalf("Create booking %d: %v", i, err)
		}
	}

	result, err := repo.ListByUser(ctx, userID, 1, 10)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", result.TotalCount)
	}
	if len(result.Items) != 3 {
		t.Errorf("Items len = %d, want 3", len(result.Items))
	}

	// Different user should get 0
	result, err = repo.ListByUser(ctx, uuid.New(), 1, 10)
	if err != nil {
		t.Fatalf("ListByUser other: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("TotalCount for other user = %d, want 0", result.TotalCount)
	}
}

func TestRepresentativeRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewRepresentativeRepo()

	userID := uuid.New()
	bhID := uuid.New()
	ownerID := uuid.New()

	rep := &domain.Representative{
		UserID:      userID,
		BathhouseID: bhID,
		OwnerID:     ownerID,
	}

	// Create
	if err := repo.Create(ctx, rep); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rep.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// Duplicate
	dup := &domain.Representative{UserID: userID, BathhouseID: bhID, OwnerID: ownerID}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create duplicate: want ErrAlreadyExists, got %v", err)
	}

	// GetByUserAndBathhouse
	got, err := repo.GetByUserAndBathhouse(ctx, userID, bhID)
	if err != nil {
		t.Fatalf("GetByUserAndBathhouse: %v", err)
	}
	if got.ID != rep.ID {
		t.Errorf("ID = %v, want %v", got.ID, rep.ID)
	}

	// ListByBathhouse
	list, err := repo.ListByBathhouse(ctx, bhID)
	if err != nil {
		t.Fatalf("ListByBathhouse: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListByBathhouse len = %d, want 1", len(list))
	}

	// ListByUser
	list, err = repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListByUser len = %d, want 1", len(list))
	}

	// ListBathhouseIDsByUser
	ids, err := repo.ListBathhouseIDsByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListBathhouseIDsByUser: %v", err)
	}
	if len(ids) != 1 || ids[0] != bhID {
		t.Errorf("ListBathhouseIDsByUser = %v, want [%v]", ids, bhID)
	}

	// Delete
	if err := repo.Delete(ctx, rep.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = repo.GetByUserAndBathhouse(ctx, userID, bhID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByUserAndBathhouse after delete: want ErrNotFound, got %v", err)
	}
}

func TestReviewRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewReviewRepo()

	userID := uuid.New()
	bhID := uuid.New()
	bookingID := uuid.New()

	review := &domain.Review{
		UserID:      userID,
		BathhouseID: bhID,
		BookingID:   bookingID,
		Rating:      5,
		Text:        "Excellent!",
	}

	// Create
	if err := repo.Create(ctx, review); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if review.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// Duplicate user+booking
	dup := &domain.Review{UserID: userID, BathhouseID: bhID, BookingID: bookingID, Rating: 3}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create duplicate: want ErrAlreadyExists, got %v", err)
	}

	// GetByBookingID
	got, err := repo.GetByBookingID(ctx, bookingID)
	if err != nil {
		t.Fatalf("GetByBookingID: %v", err)
	}
	if got.Rating != 5 {
		t.Errorf("Rating = %d, want 5", got.Rating)
	}

	// ListByBathhouse
	result, err := repo.ListByBathhouse(ctx, bhID, 1, 10)
	if err != nil {
		t.Fatalf("ListByBathhouse: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.TotalCount)
	}

	// Not found
	_, err = repo.GetByBookingID(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByBookingID not found: want ErrNotFound, got %v", err)
	}
}

func TestBathhouseRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewBathhouseRepo()

	ownerID := uuid.New()
	bh := &domain.Bathhouse{
		OwnerID:      ownerID,
		Name:         "Test Banya",
		Address:      "123 Street",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
		Status:       domain.BathhouseStatusActive,
	}

	// Create
	if err := repo.Create(ctx, bh); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if bh.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// GetByID
	got, err := repo.GetByID(ctx, bh.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Test Banya" {
		t.Errorf("Name = %q, want %q", got.Name, "Test Banya")
	}

	// Update
	bh.Name = "Updated Banya"
	if err := repo.Update(ctx, bh); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = repo.GetByID(ctx, bh.ID)
	if got.Name != "Updated Banya" {
		t.Errorf("Name after update = %q, want %q", got.Name, "Updated Banya")
	}

	// UpdateStatus
	if err := repo.UpdateStatus(ctx, bh.ID, domain.BathhouseStatusPending); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ = repo.GetByID(ctx, bh.ID)
	if got.Status != domain.BathhouseStatusPending {
		t.Errorf("Status = %v, want pending", got.Status)
	}

	// ListByOwner
	result, err := repo.ListByOwner(ctx, ownerID, 1, 10)
	if err != nil {
		t.Fatalf("ListByOwner: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.TotalCount)
	}

	// List with filter (active only, our bh is now pending)
	listResult, err := repo.List(ctx, domain.BathhouseFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if listResult.TotalCount != 0 {
		t.Errorf("List active TotalCount = %d, want 0 (bh is pending)", listResult.TotalCount)
	}

	// List with status filter
	pending := domain.BathhouseStatusPending
	listResult, err = repo.List(ctx, domain.BathhouseFilter{Status: &pending, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List pending: %v", err)
	}
	if listResult.TotalCount != 1 {
		t.Errorf("List pending TotalCount = %d, want 1", listResult.TotalCount)
	}

	// Delete
	if err := repo.Delete(ctx, bh.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = repo.GetByID(ctx, bh.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID after delete: want ErrNotFound, got %v", err)
	}

	// Delete not found
	if err := repo.Delete(ctx, uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete not found: want ErrNotFound, got %v", err)
	}
}

func TestBathhouseRepo_ListWithFilters(t *testing.T) {
	ctx := context.Background()
	repo := NewBathhouseRepo()

	ownerID := uuid.New()

	// Create bathhouses with different attributes
	bh1 := &domain.Bathhouse{
		OwnerID: ownerID, Name: "Cheap Banya", CityID: 1,
		PricePerHour: 2000, MinDuration: 1, MaxGuests: 5,
		Address: "addr1", Status: domain.BathhouseStatusActive,
	}
	bh2 := &domain.Bathhouse{
		OwnerID: ownerID, Name: "Expensive Banya", CityID: 2,
		PricePerHour: 10000, MinDuration: 2, MaxGuests: 15,
		Address: "addr2", Status: domain.BathhouseStatusActive,
	}

	_ = repo.Create(ctx, bh1)
	_ = repo.Create(ctx, bh2)

	// Filter by city
	cityID := int64(1)
	result, _ := repo.List(ctx, domain.BathhouseFilter{CityID: &cityID, Page: 1, PageSize: 10})
	if result.TotalCount != 1 {
		t.Errorf("Filter by city: TotalCount = %d, want 1", result.TotalCount)
	}

	// Filter by price range
	priceMin := int64(3000)
	result, _ = repo.List(ctx, domain.BathhouseFilter{PriceMin: &priceMin, Page: 1, PageSize: 10})
	if result.TotalCount != 1 {
		t.Errorf("Filter by priceMin: TotalCount = %d, want 1", result.TotalCount)
	}

	priceMax := int64(5000)
	result, _ = repo.List(ctx, domain.BathhouseFilter{PriceMax: &priceMax, Page: 1, PageSize: 10})
	if result.TotalCount != 1 {
		t.Errorf("Filter by priceMax: TotalCount = %d, want 1", result.TotalCount)
	}
}

func TestBathhouseRepo_ListWithGuestCount(t *testing.T) {
	ctx := context.Background()
	repo := NewBathhouseRepo()

	ownerID := uuid.New()
	_ = repo.Create(ctx, &domain.Bathhouse{
		OwnerID: ownerID, Name: "Small", CityID: 1,
		PricePerHour: 2000, MinDuration: 1, MaxGuests: 4,
		Address: "a1", Status: domain.BathhouseStatusActive,
	})
	_ = repo.Create(ctx, &domain.Bathhouse{
		OwnerID: ownerID, Name: "Large", CityID: 1,
		PricePerHour: 5000, MinDuration: 1, MaxGuests: 12,
		Address: "a2", Status: domain.BathhouseStatusActive,
	})

	gc := 10
	result, err := repo.List(ctx, domain.BathhouseFilter{GuestCount: &gc, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List with GuestCount: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("GuestCount filter: TotalCount = %d, want 1", result.TotalCount)
	}
	if len(result.Items) == 1 && result.Items[0].Name != "Large" {
		t.Errorf("Expected Large bathhouse, got %q", result.Items[0].Name)
	}
}

func TestBathhouseRepo_ListWithSearchQuery(t *testing.T) {
	ctx := context.Background()
	repo := NewBathhouseRepo()

	ownerID := uuid.New()
	_ = repo.Create(ctx, &domain.Bathhouse{
		OwnerID: ownerID, Name: "Русская баня", Description: "Традиционная парная",
		CityID: 1, PricePerHour: 3000, MinDuration: 1, MaxGuests: 6,
		Address: "a1", Status: domain.BathhouseStatusActive,
	})
	_ = repo.Create(ctx, &domain.Bathhouse{
		OwnerID: ownerID, Name: "Финская сауна", Description: "Современная финская",
		CityID: 1, PricePerHour: 5000, MinDuration: 1, MaxGuests: 8,
		Address: "a2", Status: domain.BathhouseStatusActive,
	})

	// Search by name
	q := "сауна"
	result, err := repo.List(ctx, domain.BathhouseFilter{SearchQuery: &q, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List with SearchQuery: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("SearchQuery by name: TotalCount = %d, want 1", result.TotalCount)
	}

	// Search by description
	q2 := "парная"
	result, err = repo.List(ctx, domain.BathhouseFilter{SearchQuery: &q2, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List with SearchQuery desc: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("SearchQuery by description: TotalCount = %d, want 1", result.TotalCount)
	}

	// Search with no match
	q3 := "хаммам"
	result, err = repo.List(ctx, domain.BathhouseFilter{SearchQuery: &q3, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List with SearchQuery no match: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("SearchQuery no match: TotalCount = %d, want 0", result.TotalCount)
	}
}

func TestBathhouseRepo_ListWithOpenNow(t *testing.T) {
	ctx := context.Background()
	repo := NewBathhouseRepo()

	ownerID := uuid.New()
	now := time.Now()
	d := now.Weekday()
	dayOfWeek := int(d) - 1
	if d == time.Sunday {
		dayOfWeek = 6
	}

	// Bathhouse that is open now
	_ = repo.Create(ctx, &domain.Bathhouse{
		OwnerID: ownerID, Name: "Open Now", CityID: 1,
		PricePerHour: 3000, MinDuration: 1, MaxGuests: 6,
		Address: "a1", Status: domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: dayOfWeek, OpenTime: "00:00", CloseTime: "23:59"},
		},
	})
	// Bathhouse that is closed now (different day)
	closedDay := (dayOfWeek + 1) % 7
	_ = repo.Create(ctx, &domain.Bathhouse{
		OwnerID: ownerID, Name: "Closed Now", CityID: 1,
		PricePerHour: 5000, MinDuration: 1, MaxGuests: 8,
		Address: "a2", Status: domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: closedDay, OpenTime: "00:00", CloseTime: "23:59"},
		},
	})

	openNow := true
	result, err := repo.List(ctx, domain.BathhouseFilter{OpenNow: &openNow, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List with OpenNow: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("OpenNow filter: TotalCount = %d, want 1", result.TotalCount)
	}
	if len(result.Items) == 1 && result.Items[0].Name != "Open Now" {
		t.Errorf("Expected 'Open Now', got %q", result.Items[0].Name)
	}
}
