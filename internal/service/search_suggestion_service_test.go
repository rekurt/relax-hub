package service_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
	"github.com/redis/go-redis/v9"
)

type searchSuggestionTestEnv struct {
	svc       service.SearchSuggestionService
	bhRepo    *mock.BathhouseRepo
	cityRepo  *mock.CityRepo
	redis     *redis.Client
	miniredis *miniredis.Miniredis
}

func newSearchSuggestionTestEnv(t *testing.T) *searchSuggestionTestEnv {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	bhRepo := mock.NewBathhouseRepo()
	cityRepo := mock.NewCityRepo()
	log := logger.New(logger.LevelWarn)

	svc := service.NewSearchSuggestionService(bhRepo, cityRepo, rdb, log)

	t.Cleanup(func() {
		rdb.Close()
		mr.Close()
	})

	return &searchSuggestionTestEnv{
		svc:       svc,
		bhRepo:    bhRepo,
		cityRepo:  cityRepo,
		redis:     rdb,
		miniredis: mr,
	}
}

func TestSearchSuggestion_GetSuggestions_ShortQuery(t *testing.T) {
	env := newSearchSuggestionTestEnv(t)
	ctx := context.Background()

	suggestions, err := env.svc.GetSuggestions(ctx, "а", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for short query, got %d", len(suggestions))
	}
}

func TestSearchSuggestion_GetSuggestions_EmptyQuery(t *testing.T) {
	env := newSearchSuggestionTestEnv(t)
	ctx := context.Background()

	suggestions, err := env.svc.GetSuggestions(ctx, "", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for empty query, got %d", len(suggestions))
	}
}

func TestSearchSuggestion_GetSuggestions_BathhouseNames(t *testing.T) {
	env := newSearchSuggestionTestEnv(t)
	ctx := context.Background()

	// Create active bathhouses with matching names
	bh1 := &domain.Bathhouse{
		ID:           mustUUID(),
		OwnerID:      mustUUID(),
		Name:         "Русская баня на дровах",
		Address:      "ул. Пушкина 1",
		CityID:       1,
		PricePerHour: 500000,
		MinDuration:  1,
		MaxGuests:    10,
		Status:       domain.BathhouseStatusActive,
	}
	bh2 := &domain.Bathhouse{
		ID:           mustUUID(),
		OwnerID:      mustUUID(),
		Name:         "Финская сауна премиум",
		Address:      "ул. Ленина 2",
		CityID:       1,
		PricePerHour: 800000,
		MinDuration:  1,
		MaxGuests:    6,
		Status:       domain.BathhouseStatusActive,
	}
	// Inactive bathhouse should not appear
	bh3 := &domain.Bathhouse{
		ID:           mustUUID(),
		OwnerID:      mustUUID(),
		Name:         "Русская баня Закрытая",
		Address:      "ул. Мира 3",
		CityID:       1,
		PricePerHour: 300000,
		MinDuration:  1,
		MaxGuests:    8,
		Status:       domain.BathhouseStatusInactive,
	}

	_ = env.bhRepo.Create(ctx, bh1)
	_ = env.bhRepo.Create(ctx, bh2)
	_ = env.bhRepo.Create(ctx, bh3)

	suggestions, err := env.svc.GetSuggestions(ctx, "баня", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find "Русская баня на дровах" but not the inactive one
	found := false
	foundInactive := false
	for _, s := range suggestions {
		if s.Text == "Русская баня на дровах" && s.Type == "bathhouse" {
			found = true
		}
		if s.Text == "Русская баня Закрытая" {
			foundInactive = true
		}
	}
	if !found {
		t.Error("expected to find active bathhouse suggestion")
	}
	if foundInactive {
		t.Error("should not find inactive bathhouse")
	}
}

func TestSearchSuggestion_GetSuggestions_CityNames(t *testing.T) {
	env := newSearchSuggestionTestEnv(t)
	ctx := context.Background()

	_ = env.cityRepo.Create(ctx, &domain.City{Name: "Москва", Slug: "moscow"})
	_ = env.cityRepo.Create(ctx, &domain.City{Name: "Санкт-Петербург", Slug: "spb"})
	_ = env.cityRepo.Create(ctx, &domain.City{Name: "Новосибирск", Slug: "novosibirsk"})

	suggestions, err := env.svc.GetSuggestions(ctx, "Моск", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, s := range suggestions {
		if s.Text == "Москва" && s.Type == "city" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find city suggestion for Москва")
	}
}

func TestSearchSuggestion_GetSuggestions_PopularQueries(t *testing.T) {
	env := newSearchSuggestionTestEnv(t)
	ctx := context.Background()

	// Record queries multiple times to make them popular
	for i := 0; i < 5; i++ {
		_ = env.svc.RecordQuery(ctx, "баня с бассейном")
	}
	for i := 0; i < 3; i++ {
		_ = env.svc.RecordQuery(ctx, "баня на дровах")
	}
	// Only once - should not appear (minimum 2)
	_ = env.svc.RecordQuery(ctx, "баня редкая")

	suggestions, err := env.svc.GetSuggestions(ctx, "баня", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundPool := false
	foundWood := false
	foundRare := false
	for _, s := range suggestions {
		if s.Text == "баня с бассейном" && s.Type == "popular" {
			foundPool = true
		}
		if s.Text == "баня на дровах" && s.Type == "popular" {
			foundWood = true
		}
		if s.Text == "баня редкая" {
			foundRare = true
		}
	}
	if !foundPool {
		t.Error("expected to find popular suggestion 'баня с бассейном'")
	}
	if !foundWood {
		t.Error("expected to find popular suggestion 'баня на дровах'")
	}
	if foundRare {
		t.Error("should not find query with only 1 search")
	}
}

func TestSearchSuggestion_GetSuggestions_LimitResults(t *testing.T) {
	env := newSearchSuggestionTestEnv(t)
	ctx := context.Background()

	// Create many cities matching
	for i := 0; i < 15; i++ {
		_ = env.cityRepo.Create(ctx, &domain.City{
			Name: "Город " + string(rune('А'+i)),
			Slug: "city-" + string(rune('a'+i)),
		})
	}

	suggestions, err := env.svc.GetSuggestions(ctx, "Город", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(suggestions) > 5 {
		t.Errorf("expected at most 5 suggestions, got %d", len(suggestions))
	}
}

func TestSearchSuggestion_RecordQuery_ShortQuery(t *testing.T) {
	env := newSearchSuggestionTestEnv(t)
	ctx := context.Background()

	err := env.svc.RecordQuery(ctx, "а")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not be recorded
	score, _ := env.miniredis.ZScore("search:popular_queries", "а")
	if score != 0 {
		t.Error("short query should not be recorded")
	}
}

func TestSearchSuggestion_RecordQuery_Success(t *testing.T) {
	env := newSearchSuggestionTestEnv(t)
	ctx := context.Background()

	err := env.svc.RecordQuery(ctx, "Баня с бассейном")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be recorded in lowercase
	score, _ := env.miniredis.ZScore("search:popular_queries", "баня с бассейном")
	if score != 1 {
		t.Errorf("expected score 1, got %v", score)
	}

	// Record again
	_ = env.svc.RecordQuery(ctx, "Баня с бассейном")
	score, _ = env.miniredis.ZScore("search:popular_queries", "баня с бассейном")
	if score != 2 {
		t.Errorf("expected score 2, got %v", score)
	}
}

func TestSearchSuggestion_GetSuggestions_Deduplication(t *testing.T) {
	env := newSearchSuggestionTestEnv(t)
	ctx := context.Background()

	// Create a city and a popular query with the same text
	_ = env.cityRepo.Create(ctx, &domain.City{Name: "Москва", Slug: "moscow"})

	for i := 0; i < 5; i++ {
		_ = env.svc.RecordQuery(ctx, "москва")
	}

	suggestions, err := env.svc.GetSuggestions(ctx, "моск", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count entries for "Москва" / "москва"
	count := 0
	for _, s := range suggestions {
		if s.Text == "Москва" || s.Text == "москва" {
			count++
		}
	}
	if count > 1 {
		t.Errorf("expected at most 1 deduplicated entry for Москва, got %d", count)
	}
}

func mustUUID() uuid.UUID {
	return uuid.New()
}
