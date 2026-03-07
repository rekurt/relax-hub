package mock

import (
	"context"
	"errors"
	"fmt"
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
	_ repository.NotificationRepository   = (*NotificationRepo)(nil)
	_ repository.SocialAccountRepository  = (*SocialAccountRepo)(nil)
	_ repository.RecommendationRepository = (*RecommendationRepo)(nil)
	_ repository.SubscriptionRepository   = (*SubscriptionRepo)(nil)
	_ repository.PromotionRepository      = (*PromotionRepo)(nil)
	_ repository.LoyaltyRepository        = (*LoyaltyRepo)(nil)
	_ repository.AnalyticsRepository      = (*AnalyticsRepo)(nil)
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

func TestUserRepo_ProfileFields(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepo()

	cityID := int64(42)
	user := &domain.User{
		Email:        "profile@example.com",
		PasswordHash: "hash",
		Name:         "Profile User",
		Phone:        "+7999111111",
		Role:         domain.RoleClient,
		IsActive:     true,
		AvatarURL:    "https://example.com/avatar.jpg",
		Bio:          "I love saunas",
		CityID:       &cityID,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.AvatarURL != "https://example.com/avatar.jpg" {
		t.Errorf("AvatarURL = %q, want %q", got.AvatarURL, "https://example.com/avatar.jpg")
	}
	if got.Bio != "I love saunas" {
		t.Errorf("Bio = %q, want %q", got.Bio, "I love saunas")
	}
	if got.CityID == nil || *got.CityID != 42 {
		t.Errorf("CityID = %v, want 42", got.CityID)
	}

	// Update profile fields
	user.Bio = "Updated bio"
	user.AvatarURL = "https://example.com/new-avatar.jpg"
	newCityID := int64(99)
	user.CityID = &newCityID
	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ = repo.GetByID(ctx, user.ID)
	if got.Bio != "Updated bio" {
		t.Errorf("Bio after update = %q, want %q", got.Bio, "Updated bio")
	}
	if got.AvatarURL != "https://example.com/new-avatar.jpg" {
		t.Errorf("AvatarURL after update = %q, want %q", got.AvatarURL, "https://example.com/new-avatar.jpg")
	}
	if got.CityID == nil || *got.CityID != 99 {
		t.Errorf("CityID after update = %v, want 99", got.CityID)
	}

	// User without profile fields (nil CityID)
	user2 := &domain.User{
		Email:        "noprofile@example.com",
		PasswordHash: "hash",
		Name:         "No Profile",
		Role:         domain.RoleClient,
		IsActive:     true,
	}
	if err := repo.Create(ctx, user2); err != nil {
		t.Fatalf("Create user2: %v", err)
	}
	got2, _ := repo.GetByID(ctx, user2.ID)
	if got2.CityID != nil {
		t.Errorf("CityID should be nil, got %v", got2.CityID)
	}
	if got2.AvatarURL != "" {
		t.Errorf("AvatarURL should be empty, got %q", got2.AvatarURL)
	}
	if got2.Bio != "" {
		t.Errorf("Bio should be empty, got %q", got2.Bio)
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
		Status:      domain.ReviewStatusApproved,
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

func TestNotificationRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewNotificationRepo()

	userID := uuid.New()

	notif := &domain.Notification{
		UserID: userID,
		Type:   domain.NotifBookingConfirmed,
		Title:  "Booking Confirmed",
		Body:   "Your booking has been confirmed",
		Data:   map[string]string{"booking_id": uuid.New().String()},
	}

	// Create
	if err := repo.Create(ctx, notif); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if notif.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}
	if notif.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should be set after Create")
	}

	// GetByID
	got, err := repo.GetByID(ctx, notif.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Booking Confirmed" {
		t.Errorf("Title = %q, want %q", got.Title, "Booking Confirmed")
	}
	if got.IsRead {
		t.Error("New notification should not be read")
	}
	if got.Data["booking_id"] == "" {
		t.Error("Data should contain booking_id")
	}

	// GetByID not found
	_, err = repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID not found: want ErrNotFound, got %v", err)
	}

	// ListByUser
	result, err := repo.ListByUser(ctx, userID, 1, 10)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.TotalCount)
	}

	// ListByUser other user should get 0
	result, err = repo.ListByUser(ctx, uuid.New(), 1, 10)
	if err != nil {
		t.Fatalf("ListByUser other: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("TotalCount for other user = %d, want 0", result.TotalCount)
	}

	// CountUnread
	count, err := repo.CountUnread(ctx, userID)
	if err != nil {
		t.Fatalf("CountUnread: %v", err)
	}
	if count != 1 {
		t.Errorf("CountUnread = %d, want 1", count)
	}

	// MarkAsRead
	if err := repo.MarkAsRead(ctx, notif.ID); err != nil {
		t.Fatalf("MarkAsRead: %v", err)
	}
	got, _ = repo.GetByID(ctx, notif.ID)
	if !got.IsRead {
		t.Error("Notification should be read after MarkAsRead")
	}
	if got.ReadAt == nil {
		t.Error("ReadAt should be set after MarkAsRead")
	}

	// MarkAsRead on already read -> not found
	if err := repo.MarkAsRead(ctx, notif.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("MarkAsRead already read: want ErrNotFound, got %v", err)
	}

	// MarkAsRead not found
	if err := repo.MarkAsRead(ctx, uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("MarkAsRead not found: want ErrNotFound, got %v", err)
	}

	// CountUnread after marking as read
	count, err = repo.CountUnread(ctx, userID)
	if err != nil {
		t.Fatalf("CountUnread after read: %v", err)
	}
	if count != 0 {
		t.Errorf("CountUnread after read = %d, want 0", count)
	}
}

func TestNotificationRepo_MarkAllAsRead(t *testing.T) {
	ctx := context.Background()
	repo := NewNotificationRepo()

	userID := uuid.New()

	// Create 3 notifications
	for i := range 3 {
		n := &domain.Notification{
			UserID: userID,
			Type:   domain.NotifSystem,
			Title:  fmt.Sprintf("Notif %d", i),
			Body:   "Body",
		}
		if err := repo.Create(ctx, n); err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}

	// All should be unread
	count, _ := repo.CountUnread(ctx, userID)
	if count != 3 {
		t.Errorf("CountUnread = %d, want 3", count)
	}

	// Mark all as read
	if err := repo.MarkAllAsRead(ctx, userID); err != nil {
		t.Fatalf("MarkAllAsRead: %v", err)
	}

	// All should be read
	count, _ = repo.CountUnread(ctx, userID)
	if count != 0 {
		t.Errorf("CountUnread after MarkAllAsRead = %d, want 0", count)
	}
}

func TestNotificationRepo_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := NewNotificationRepo()

	userID := uuid.New()

	// Create 5 notifications
	for i := range 5 {
		n := &domain.Notification{
			UserID:    userID,
			Type:      domain.NotifSystem,
			Title:     fmt.Sprintf("Notif %d", i),
			Body:      "Body",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		}
		if err := repo.Create(ctx, n); err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}

	// Page 1, size 2
	result, err := repo.ListByUser(ctx, userID, 1, 2)
	if err != nil {
		t.Fatalf("ListByUser page 1: %v", err)
	}
	if result.TotalCount != 5 {
		t.Errorf("TotalCount = %d, want 5", result.TotalCount)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items len = %d, want 2", len(result.Items))
	}
	if result.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", result.TotalPages)
	}

	// Page beyond range
	result, err = repo.ListByUser(ctx, userID, 10, 2)
	if err != nil {
		t.Fatalf("ListByUser page 10: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("Items len for out-of-range page = %d, want 0", len(result.Items))
	}
}

func TestNotificationRepo_Preferences(t *testing.T) {
	ctx := context.Background()
	repo := NewNotificationRepo()

	userID := uuid.New()

	// GetPreferences should return defaults when no custom prefs exist
	prefs, err := repo.GetPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("GetPreferences defaults: %v", err)
	}
	if !prefs.InApp {
		t.Error("Default InApp should be true")
	}
	if !prefs.Email {
		t.Error("Default Email should be true")
	}
	if prefs.Push {
		t.Error("Default Push should be false")
	}
	if !prefs.BookingEvents {
		t.Error("Default BookingEvents should be true")
	}

	// UpdatePreferences
	prefs.Push = true
	prefs.PromoEvents = false
	if err := repo.UpdatePreferences(ctx, prefs); err != nil {
		t.Fatalf("UpdatePreferences: %v", err)
	}

	// GetPreferences should return updated values
	got, err := repo.GetPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("GetPreferences after update: %v", err)
	}
	if !got.Push {
		t.Error("Push should be true after update")
	}
	if got.PromoEvents {
		t.Error("PromoEvents should be false after update")
	}
	if !got.InApp {
		t.Error("InApp should still be true")
	}
}

func TestNotificationRepo_DataIsolation(t *testing.T) {
	ctx := context.Background()
	repo := NewNotificationRepo()

	notif := &domain.Notification{
		UserID: uuid.New(),
		Type:   domain.NotifPromo,
		Title:  "Promo",
		Body:   "Special offer",
		Data:   map[string]string{"promo_id": "123"},
	}
	if err := repo.Create(ctx, notif); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Modify the original data - should not affect stored copy
	notif.Data["promo_id"] = "modified"

	got, _ := repo.GetByID(ctx, notif.ID)
	if got.Data["promo_id"] != "123" {
		t.Errorf("Data should be isolated, got promo_id=%q, want %q", got.Data["promo_id"], "123")
	}
}

func TestSocialAccountRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewSocialAccountRepo()

	userID := uuid.New()

	account := &domain.SocialAccount{
		UserID:     userID,
		Provider:   domain.OAuthProviderVK,
		ProviderID: "vk-123",
		Email:      "test@vk.com",
		Name:       "VK User",
		AvatarURL:  "https://vk.com/photo.jpg",
		LinkedAt:   time.Now(),
	}

	// Create
	if err := repo.Create(ctx, account); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if account.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// GetByProviderAndID
	got, err := repo.GetByProviderAndID(ctx, domain.OAuthProviderVK, "vk-123")
	if err != nil {
		t.Fatalf("GetByProviderAndID: %v", err)
	}
	if got.Email != "test@vk.com" {
		t.Errorf("Email = %q, want %q", got.Email, "test@vk.com")
	}

	// GetByProviderAndID not found
	_, err = repo.GetByProviderAndID(ctx, domain.OAuthProviderGoogle, "nonexistent")
	if !errors.Is(err, domain.ErrSocialAccountNotFound) {
		t.Errorf("GetByProviderAndID not found: want ErrSocialAccountNotFound, got %v", err)
	}

	// Duplicate provider+provider_id
	dup := &domain.SocialAccount{
		UserID:     uuid.New(),
		Provider:   domain.OAuthProviderVK,
		ProviderID: "vk-123",
		LinkedAt:   time.Now(),
	}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrSocialAccountAlreadyLinked) {
		t.Errorf("Create duplicate: want ErrSocialAccountAlreadyLinked, got %v", err)
	}

	// ListByUser
	accounts, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("ListByUser len = %d, want 1", len(accounts))
	}

	// Delete
	if err := repo.Delete(ctx, userID, domain.OAuthProviderVK); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	accounts, _ = repo.ListByUser(ctx, userID)
	if len(accounts) != 0 {
		t.Errorf("ListByUser after delete: len = %d, want 0", len(accounts))
	}

	// Delete not found
	if err := repo.Delete(ctx, userID, domain.OAuthProviderVK); !errors.Is(err, domain.ErrSocialAccountNotFound) {
		t.Errorf("Delete not found: want ErrSocialAccountNotFound, got %v", err)
	}
}

func TestRecommendationRepo_UserPreferences(t *testing.T) {
	ctx := context.Background()
	repo := NewRecommendationRepo()

	userID := uuid.New()
	cityID := int64(1)
	minPrice := int64(1000)
	maxPrice := int64(5000)

	prefs := &domain.UserPreferences{
		UserID:          userID,
		PreferredCityID: &cityID,
		PriceRangeMin:   &minPrice,
		PriceRangeMax:   &maxPrice,
		PreferPool:      true,
		PreferSauna:     true,
		PreferSteamRoom: false,
	}

	// Save preferences
	if err := repo.SaveUserPreferences(ctx, prefs); err != nil {
		t.Fatalf("SaveUserPreferences: %v", err)
	}

	// Get preferences
	got, err := repo.GetUserPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserPreferences: %v", err)
	}
	if got.PreferPool != true {
		t.Error("PreferPool should be true")
	}
	if got.PreferSauna != true {
		t.Error("PreferSauna should be true")
	}
	if got.PreferSteamRoom != false {
		t.Error("PreferSteamRoom should be false")
	}

	// Not found
	_, err = repo.GetUserPreferences(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetUserPreferences not found: want ErrNotFound, got %v", err)
	}

	// Update preferences
	newCityID := int64(2)
	prefs.PreferredCityID = &newCityID
	if err := repo.SaveUserPreferences(ctx, prefs); err != nil {
		t.Fatalf("SaveUserPreferences update: %v", err)
	}

	got, _ = repo.GetUserPreferences(ctx, userID)
	if got.PreferredCityID == nil || *got.PreferredCityID != 2 {
		t.Errorf("PreferredCityID after update = %v, want 2", got.PreferredCityID)
	}
}

func TestRecommendationRepo_UserActivity(t *testing.T) {
	ctx := context.Background()
	repo := NewRecommendationRepo()

	userID := uuid.New()
	bhID := uuid.New()

	activity := &domain.UserActivity{
		UserID:      userID,
		BathhouseID: bhID,
		Type:        domain.ActivityTypeView,
	}

	// Record activity
	if err := repo.RecordActivity(ctx, activity); err != nil {
		t.Fatalf("RecordActivity: %v", err)
	}
	if activity.ID == uuid.Nil {
		t.Fatal("ID should be assigned after RecordActivity")
	}
	if activity.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should be assigned after RecordActivity")
	}

	// Invalid activity
	invalidActivity := &domain.UserActivity{
		UserID: uuid.Nil, // invalid
		Type:   domain.ActivityTypeView,
	}
	if err := repo.RecordActivity(ctx, invalidActivity); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("RecordActivity invalid: want ErrInvalidInput, got %v", err)
	}
}

func TestRecommendationRepo_BookedBathhouses(t *testing.T) {
	ctx := context.Background()
	repo := NewRecommendationRepo()

	userID := uuid.New()
	bh1 := uuid.New()
	bh2 := uuid.New()

	// Manually add bookings
	repo.mu.Lock()
	repo.bookings[userID] = []uuid.UUID{bh1, bh2}
	repo.mu.Unlock()

	// Get booked bathhouses
	result, err := repo.GetUserBookedBathhouses(ctx, userID, 10)
	if err != nil {
		t.Fatalf("GetUserBookedBathhouses: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Booked bathhouses len = %d, want 2", len(result))
	}

	// Empty list for user with no bookings
	result, err = repo.GetUserBookedBathhouses(ctx, uuid.New(), 10)
	if err != nil {
		t.Fatalf("GetUserBookedBathhouses empty: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Empty user len = %d, want 0", len(result))
	}
}

func TestRecommendationRepo_SimilarUsers(t *testing.T) {
	ctx := context.Background()
	repo := NewRecommendationRepo()

	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()
	bh1 := uuid.New()
	bh2 := uuid.New()

	// Manually setup overlapping bookings
	repo.mu.Lock()
	repo.bookings[user1] = []uuid.UUID{bh1, bh2}
	repo.bookings[user2] = []uuid.UUID{bh1, bh2} // Same as user1 (max overlap)
	repo.bookings[user3] = []uuid.UUID{bh1}      // Partial overlap
	repo.mu.Unlock()

	// Get similar users for user1
	similar, err := repo.GetSimilarUsers(ctx, user1, 10)
	if err != nil {
		t.Fatalf("GetSimilarUsers: %v", err)
	}

	// Should find user2 and user3, with user2 first (higher score)
	if len(similar) < 1 {
		t.Errorf("Similar users len = %d, want at least 1", len(similar))
	}
	if len(similar) > 0 && similar[0] != user2 {
		t.Errorf("Most similar user should be user2, got %v", similar[0])
	}
}

func TestRecommendationRepo_SimilarBathhouses(t *testing.T) {
	ctx := context.Background()
	repo := NewRecommendationRepo()

	bhID := uuid.New()

	// Mock implementation returns empty - behavior verified by interface test
	result, err := repo.GetSimilarBathhouses(ctx, bhID, 10)
	if err != nil {
		t.Fatalf("GetSimilarBathhouses: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Mock GetSimilarBathhouses should return empty, got %d items", len(result))
	}
}

func TestRecommendationRepo_PopularBathhouses(t *testing.T) {
	ctx := context.Background()
	repo := NewRecommendationRepo()

	// Mock implementation returns empty - behavior verified by interface test
	result, err := repo.GetPopularBathhouses(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetPopularBathhouses: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Mock GetPopularBathhouses should return empty, got %d items", len(result))
	}
}

func TestSubscriptionRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewSubscriptionRepo()

	bathhouseID := uuid.New()
	ownerID := uuid.New()

	sub := &domain.Subscription{
		BathhouseID:  bathhouseID,
		OwnerID:      ownerID,
		Plan:         domain.PlanPremium,
		Status:       domain.SubscriptionActive,
		StartDate:    time.Now(),
		EndDate:      nil,
		AutoRenew:    true,
		PriceKopecks: 299900,
	}

	// Create
	if err := repo.Create(ctx, sub); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sub.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// GetByID
	got, err := repo.GetByID(ctx, sub.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Plan != domain.PlanPremium {
		t.Errorf("Got plan %v, want %v", got.Plan, domain.PlanPremium)
	}

	// GetActiveBybathhouse
	active, err := repo.GetActiveBybathhouse(ctx, bathhouseID)
	if err != nil {
		t.Fatalf("GetActiveBybathhouse: %v", err)
	}
	if active.ID != sub.ID {
		t.Errorf("Got subscription %v, want %v", active.ID, sub.ID)
	}

	// Update
	sub.Plan = domain.PlanPromoted
	if err := repo.Update(ctx, sub); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := repo.GetByID(ctx, sub.ID)
	if err != nil {
		t.Fatalf("GetByID after Update: %v", err)
	}
	if updated.Plan != domain.PlanPromoted {
		t.Errorf("After update, got plan %v, want %v", updated.Plan, domain.PlanPromoted)
	}

	// ListByOwner
	result, err := repo.ListByOwner(ctx, ownerID, 1, 20)
	if err != nil {
		t.Fatalf("ListByOwner: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("Expected 1 subscription, got %d", result.TotalCount)
	}
}

func TestSubscriptionRepo_GetExpiring(t *testing.T) {
	ctx := context.Background()
	repo := NewSubscriptionRepo()

	now := time.Now()
	bathhouseID := uuid.New()
	ownerID := uuid.New()

	// Create subscription that expires tomorrow
	tomorrow := now.Add(24 * time.Hour)
	sub := &domain.Subscription{
		BathhouseID:  bathhouseID,
		OwnerID:      ownerID,
		Plan:         domain.PlanPremium,
		Status:       domain.SubscriptionActive,
		StartDate:    now,
		EndDate:      &tomorrow,
		AutoRenew:    false,
		PriceKopecks: 299900,
	}

	if err := repo.Create(ctx, sub); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Check with date before expiration - should not return
	beforeExpiration := now.Add(12 * time.Hour)
	result, err := repo.GetExpiring(ctx, beforeExpiration)
	if err != nil {
		t.Fatalf("GetExpiring: %v", err)
	}
	if len(result) > 0 {
		t.Error("GetExpiring should not return subscriptions not yet expiring")
	}

	// Check with date after expiration - should return
	afterExpiration := tomorrow.Add(1 * time.Hour)
	result, err = repo.GetExpiring(ctx, afterExpiration)
	if err != nil {
		t.Fatalf("GetExpiring: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("GetExpiring should return 1 subscription, got %d", len(result))
	}
}

func TestPromotionRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewPromotionRepo()

	bathhouseID := uuid.New()
	cityID := int64(1)

	promo := &domain.Promotion{
		BathhouseID:     bathhouseID,
		BudgetKopecks:   100000,
		SpentKopecks:    50000,
		StartDate:       time.Now(),
		EndDate:         time.Now().Add(30 * 24 * time.Hour),
		TargetCityID:    &cityID,
		Status:          domain.PromotionActive,
		ImpressionCount: 1000,
		ClickCount:      50,
	}

	// Create
	if err := repo.Create(ctx, promo); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if promo.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// GetByID
	got, err := repo.GetByID(ctx, promo.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.BudgetKopecks != 100000 {
		t.Errorf("Got budget %d, want 100000", got.BudgetKopecks)
	}

	// GetActiveBybathhouse
	active, err := repo.GetActiveBybathhouse(ctx, bathhouseID)
	if err != nil {
		t.Fatalf("GetActiveBybathhouse: %v", err)
	}
	if active.ID != promo.ID {
		t.Errorf("Got promotion %v, want %v", active.ID, promo.ID)
	}

	// Update
	promo.SpentKopecks = 75000
	promo.ImpressionCount = 1500
	if err := repo.Update(ctx, promo); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := repo.GetByID(ctx, promo.ID)
	if err != nil {
		t.Fatalf("GetByID after Update: %v", err)
	}
	if updated.SpentKopecks != 75000 {
		t.Errorf("After update, got spent %d, want 75000", updated.SpentKopecks)
	}
}

func TestSubscriptionRepo_DuplicateActive(t *testing.T) {
	ctx := context.Background()
	repo := NewSubscriptionRepo()

	bathhouseID := uuid.New()
	ownerID := uuid.New()

	sub1 := &domain.Subscription{
		BathhouseID:  bathhouseID,
		OwnerID:      ownerID,
		Plan:         domain.PlanPremium,
		Status:       domain.SubscriptionActive,
		StartDate:    time.Now(),
		PriceKopecks: 299900,
	}

	// Create first subscription
	if err := repo.Create(ctx, sub1); err != nil {
		t.Fatalf("Create first: %v", err)
	}

	// Try to create another active subscription for same bathhouse
	sub2 := &domain.Subscription{
		BathhouseID:  bathhouseID,
		OwnerID:      ownerID,
		Plan:         domain.PlanPromoted,
		Status:       domain.SubscriptionActive,
		StartDate:    time.Now(),
		PriceKopecks: 499900,
	}

	if err := repo.Create(ctx, sub2); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Expected ErrAlreadyExists, got %v", err)
	}
}

func TestLoyaltyRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewLoyaltyRepo()
	userID := uuid.New()

	// GetAccount not found
	_, err := repo.GetAccount(ctx, userID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}

	// CreateAccount
	account := &domain.LoyaltyAccount{
		UserID: userID,
		Level:  domain.LoyaltyBronze,
		Points: 0,
	}
	if err := repo.CreateAccount(ctx, account); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}

	// GetAccount
	got, err := repo.GetAccount(ctx, userID)
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Level != domain.LoyaltyBronze {
		t.Errorf("Level = %v, want %v", got.Level, domain.LoyaltyBronze)
	}

	// Duplicate create
	if err := repo.CreateAccount(ctx, account); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Expected ErrAlreadyExists, got %v", err)
	}

	// AddPoints
	if err := repo.AddPoints(ctx, userID, 100); err != nil {
		t.Fatalf("AddPoints: %v", err)
	}
	got, _ = repo.GetAccount(ctx, userID)
	if got.Points != 100 {
		t.Errorf("Points = %d, want 100", got.Points)
	}
	if got.TotalEarned != 100 {
		t.Errorf("TotalEarned = %d, want 100", got.TotalEarned)
	}

	// SpendPoints
	if err := repo.SpendPoints(ctx, userID, 30); err != nil {
		t.Fatalf("SpendPoints: %v", err)
	}
	got, _ = repo.GetAccount(ctx, userID)
	if got.Points != 70 {
		t.Errorf("Points = %d, want 70", got.Points)
	}
	if got.TotalSpent != 30 {
		t.Errorf("TotalSpent = %d, want 30", got.TotalSpent)
	}

	// SpendPoints insufficient
	if err := repo.SpendPoints(ctx, userID, 1000); !errors.Is(err, domain.ErrInsufficientPoints) {
		t.Errorf("Expected ErrInsufficientPoints, got %v", err)
	}

	// UpdateLevel
	if err := repo.UpdateLevel(ctx, userID, domain.LoyaltySilver); err != nil {
		t.Fatalf("UpdateLevel: %v", err)
	}
	got, _ = repo.GetAccount(ctx, userID)
	if got.Level != domain.LoyaltySilver {
		t.Errorf("Level = %v, want %v", got.Level, domain.LoyaltySilver)
	}

	// AddPoints/SpendPoints on non-existent account
	fakeID := uuid.New()
	if err := repo.AddPoints(ctx, fakeID, 10); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("AddPoints non-existent: expected ErrNotFound, got %v", err)
	}
	if err := repo.SpendPoints(ctx, fakeID, 10); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("SpendPoints non-existent: expected ErrNotFound, got %v", err)
	}
	if err := repo.UpdateLevel(ctx, fakeID, domain.LoyaltyGold); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("UpdateLevel non-existent: expected ErrNotFound, got %v", err)
	}
}

func TestLoyaltyRepo_Transactions(t *testing.T) {
	ctx := context.Background()
	repo := NewLoyaltyRepo()
	userID := uuid.New()
	bookingID := uuid.New()

	// Create some transactions
	for i := 0; i < 5; i++ {
		tx := &domain.LoyaltyTransaction{
			UserID:      userID,
			Type:        domain.LoyaltyTransactionEarn,
			Amount:      int64(100 + i*10),
			BookingID:   &bookingID,
			Description: fmt.Sprintf("earn %d", i),
		}
		if err := repo.CreateTransaction(ctx, tx); err != nil {
			t.Fatalf("CreateTransaction %d: %v", i, err)
		}
		if tx.ID == uuid.Nil {
			t.Fatal("transaction ID should be assigned")
		}
	}

	// Add a transaction for another user
	otherUser := uuid.New()
	otherTx := &domain.LoyaltyTransaction{
		UserID:      otherUser,
		Type:        domain.LoyaltyTransactionSpend,
		Amount:      50,
		Description: "other user",
	}
	if err := repo.CreateTransaction(ctx, otherTx); err != nil {
		t.Fatalf("CreateTransaction other: %v", err)
	}

	// ListTransactions for userID
	result, err := repo.ListTransactions(ctx, userID, 1, 3)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if result.TotalCount != 5 {
		t.Errorf("TotalCount = %d, want 5", result.TotalCount)
	}
	if len(result.Items) != 3 {
		t.Errorf("len(Items) = %d, want 3", len(result.Items))
	}
	if result.TotalPages != 2 {
		t.Errorf("TotalPages = %d, want 2", result.TotalPages)
	}

	// Page 2
	result, err = repo.ListTransactions(ctx, userID, 2, 3)
	if err != nil {
		t.Fatalf("ListTransactions page 2: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("len(Items) page 2 = %d, want 2", len(result.Items))
	}

	// Other user has only 1 transaction
	result, err = repo.ListTransactions(ctx, otherUser, 1, 20)
	if err != nil {
		t.Fatalf("ListTransactions other: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount other = %d, want 1", result.TotalCount)
	}
}

