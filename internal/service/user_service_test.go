package service_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/nikitaaldaev/bani/internal/storage"
)

var testLogger = logger.New(logger.LevelError)

// newUserService is a test helper that creates a UserService with mock repos.
func newUserService(userRepo *mock.UserRepo, fileStorage storage.FileStorage) service.UserService {
	return service.NewUserService(userRepo, mock.NewBookingRepo(), mock.NewReviewRepo(), fileStorage, testLogger)
}

// createTestJPEG creates a valid test JPEG image
func createTestJPEG(t *testing.T, width, height int) *bytes.Buffer {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return &buf
}

// createTestPNG creates a valid test PNG image
func createTestPNG(t *testing.T, width, height int) *bytes.Buffer {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func TestUserService_GetByID(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	found, err := svc.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Email != "test@example.com" {
		t.Errorf("email = %q, want %q", found.Email, "test@example.com")
	}
}

func TestUserService_Update(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	newName := "Updated Name"
	newPhone := "+79001234567"
	updated, err := svc.Update(context.Background(), user.ID, service.UpdateUserInput{
		Name: &newName, Phone: &newPhone,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("name = %q, want %q", updated.Name, newName)
	}
	if updated.Phone != newPhone {
		t.Errorf("phone = %q, want %q", updated.Phone, newPhone)
	}
}

func TestUserService_Update_Bio(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	bio := "Люблю русскую баню"
	updated, err := svc.Update(context.Background(), user.ID, service.UpdateUserInput{
		Bio: &bio,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Bio != bio {
		t.Errorf("bio = %q, want %q", updated.Bio, bio)
	}
}

func TestUserService_Update_CityID(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	cityID := int64(1)
	cityIDPtr := &cityID
	updated, err := svc.Update(context.Background(), user.ID, service.UpdateUserInput{
		CityID: &cityIDPtr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.CityID == nil || *updated.CityID != cityID {
		t.Errorf("city_id = %v, want %d", updated.CityID, cityID)
	}

	// Clear city_id
	var nilCityID *int64
	updated2, err := svc.Update(context.Background(), user.ID, service.UpdateUserInput{
		CityID: &nilCityID,
	})
	if err != nil {
		t.Fatalf("unexpected error clearing city: %v", err)
	}
	if updated2.CityID != nil {
		t.Errorf("city_id should be nil after clearing, got %v", updated2.CityID)
	}
}

func TestUserService_Update_EmptyName(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	emptyName := ""
	_, err := svc.Update(context.Background(), user.ID, service.UpdateUserInput{
		Name: &emptyName,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty name, got: %v", err)
	}
}

func TestUserService_UploadAvatar(t *testing.T) {
	userRepo := mock.NewUserRepo()
	mockStore := storage.NewMockStorage()
	svc := newUserService(userRepo, mockStore)

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	data := createTestJPEG(t, 200, 200)
	updated, err := svc.UploadAvatar(context.Background(), user.ID, service.UploadAvatarInput{
		Data:        data,
		ContentType: "image/jpeg",
		Ext:         ".jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.AvatarURL == "" {
		t.Error("avatar_url should not be empty after upload")
	}
	if mockStore.Len() != 1 {
		t.Errorf("expected 1 file in storage, got %d", mockStore.Len())
	}
}

func TestUserService_UploadAvatar_ReplacesOld(t *testing.T) {
	userRepo := mock.NewUserRepo()
	mockStore := storage.NewMockStorage()
	svc := newUserService(userRepo, mockStore)

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: true, AvatarURL: "http://mock-storage/old-avatar.jpg",
	}
	_ = userRepo.Create(context.Background(), user)

	data := createTestPNG(t, 300, 300)
	updated, err := svc.UploadAvatar(context.Background(), user.ID, service.UploadAvatarInput{
		Data:        data,
		ContentType: "image/png",
		Ext:         ".png",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.AvatarURL == "" {
		t.Error("avatar_url should not be empty")
	}
}

func TestUserService_DeleteAvatar(t *testing.T) {
	userRepo := mock.NewUserRepo()
	mockStore := storage.NewMockStorage()
	svc := newUserService(userRepo, mockStore)

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: true, AvatarURL: "http://mock-storage/avatar.jpg",
	}
	_ = userRepo.Create(context.Background(), user)

	updated, err := svc.DeleteAvatar(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.AvatarURL != "" {
		t.Errorf("avatar_url should be empty after delete, got %q", updated.AvatarURL)
	}
}

func TestUserService_GetPublicProfile(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test User",
		Bio: "Hello", Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	profile, err := svc.GetPublicProfile(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Name != "Test User" {
		t.Errorf("name = %q, want %q", profile.Name, "Test User")
	}
	if profile.Bio != "Hello" {
		t.Errorf("bio = %q, want %q", profile.Bio, "Hello")
	}
}

func TestUserService_GetPublicProfile_NotFound(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	_, err := svc.GetPublicProfile(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestUserService_Block(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	err := svc.Block(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := svc.GetByID(context.Background(), user.ID)
	if found.IsActive {
		t.Error("user should be blocked (inactive)")
	}
}

func TestUserService_Unblock(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	user := &domain.User{
		ID: uuid.New(), Email: "test@example.com", Name: "Test",
		Role: domain.RoleClient, IsActive: false,
	}
	_ = userRepo.Create(context.Background(), user)

	err := svc.Unblock(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := svc.GetByID(context.Background(), user.ID)
	if !found.IsActive {
		t.Error("user should be unblocked (active)")
	}
}

func TestUserService_List(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	for i := 0; i < 5; i++ {
		_ = userRepo.Create(context.Background(), &domain.User{
			ID: uuid.New(), Email: "user" + string(rune('a'+i)) + "@example.com",
			Name: "User", Role: domain.RoleClient, IsActive: true,
		})
	}

	result, err := svc.List(context.Background(), 1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 5 {
		t.Errorf("totalCount = %d, want 5", result.TotalCount)
	}
	if len(result.Items) != 3 {
		t.Errorf("items len = %d, want 3", len(result.Items))
	}
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newUserService(userRepo, storage.NewMockStorage())

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("should return ErrNotFound, got: %v", err)
	}
}

func TestUserService_GetMyStats(t *testing.T) {
	userRepo := mock.NewUserRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	svc := service.NewUserService(userRepo, bookingRepo, reviewRepo, storage.NewMockStorage(), testLogger)

	userID := uuid.New()
	bathhouseID := uuid.New()

	_ = userRepo.Create(context.Background(), &domain.User{
		ID: userID, Email: "stats@example.com", Name: "Stats User",
		Role: domain.RoleClient, IsActive: true,
	})

	// Create completed bookings
	for i := 0; i < 3; i++ {
		b := &domain.Booking{
			UserID: userID, BathhouseID: bathhouseID,
			GuestCount: 2, TotalPrice: 300000, // 3000 rubles in kopecks
			Status: domain.BookingCompleted,
		}
		_ = bookingRepo.Create(context.Background(), b)
	}
	// Create a cancelled booking (should not count)
	_ = bookingRepo.Create(context.Background(), &domain.Booking{
		UserID: userID, BathhouseID: bathhouseID,
		GuestCount: 2, TotalPrice: 500000,
		Status: domain.BookingCancelled,
	})

	// Create approved reviews
	for i := 0; i < 2; i++ {
		bookingID := uuid.New()
		_ = reviewRepo.Create(context.Background(), &domain.Review{
			UserID: userID, BathhouseID: bathhouseID, BookingID: bookingID,
			Rating: 4 + i, Text: "Great", Status: domain.ReviewStatusApproved,
		})
	}
	// Create a pending review (should not count)
	_ = reviewRepo.Create(context.Background(), &domain.Review{
		UserID: userID, BathhouseID: bathhouseID, BookingID: uuid.New(),
		Rating: 1, Text: "Bad", Status: domain.ReviewStatusPending,
	})

	stats, err := svc.GetMyStats(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalVisits != 3 {
		t.Errorf("total_visits = %d, want 3", stats.TotalVisits)
	}
	if stats.TotalSpent != 900000 {
		t.Errorf("total_spent = %d, want 900000", stats.TotalSpent)
	}
	if stats.AvgCheck != 300000 {
		t.Errorf("avg_check = %d, want 300000", stats.AvgCheck)
	}
	if stats.ReviewCount != 2 {
		t.Errorf("review_count = %d, want 2", stats.ReviewCount)
	}
	if stats.AvgRating != 4.5 {
		t.Errorf("avg_rating = %f, want 4.5", stats.AvgRating)
	}
}

func TestUserService_GetMyStats_Empty(t *testing.T) {
	userRepo := mock.NewUserRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	svc := service.NewUserService(userRepo, bookingRepo, reviewRepo, storage.NewMockStorage(), testLogger)

	userID := uuid.New()
	_ = userRepo.Create(context.Background(), &domain.User{
		ID: userID, Email: "empty@example.com", Name: "Empty User",
		Role: domain.RoleClient, IsActive: true,
	})

	stats, err := svc.GetMyStats(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalVisits != 0 {
		t.Errorf("total_visits = %d, want 0", stats.TotalVisits)
	}
	if stats.TotalSpent != 0 {
		t.Errorf("total_spent = %d, want 0", stats.TotalSpent)
	}
	if stats.AvgCheck != 0 {
		t.Errorf("avg_check = %d, want 0", stats.AvgCheck)
	}
	if stats.ReviewCount != 0 {
		t.Errorf("review_count = %d, want 0", stats.ReviewCount)
	}
	if stats.AvgRating != 0 {
		t.Errorf("avg_rating = %f, want 0", stats.AvgRating)
	}
}
