package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

type favoriteTestEnv struct {
	svc     service.FavoriteService
	bhRepo  *mock.BathhouseRepo
	favRepo *mock.FavoriteRepo
}

func newFavoriteTestEnv() *favoriteTestEnv {
	bhRepo := mock.NewBathhouseRepo()
	favRepo := mock.NewFavoriteRepo()
	svc := service.NewFavoriteService(favRepo, bhRepo)
	return &favoriteTestEnv{
		svc:     svc,
		bhRepo:  bhRepo,
		favRepo: favRepo,
	}
}

func TestFavoriteService_Toggle_Add(t *testing.T) {
	env := newFavoriteTestEnv()
	ownerID := uuid.New()
	userID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	isFav, err := env.svc.Toggle(context.Background(), userID, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isFav {
		t.Error("expected is_favorite=true after adding")
	}
}

func TestFavoriteService_Toggle_Remove(t *testing.T) {
	env := newFavoriteTestEnv()
	ownerID := uuid.New()
	userID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	// Add first
	_, _ = env.svc.Toggle(context.Background(), userID, bh.ID)

	// Toggle again to remove
	isFav, err := env.svc.Toggle(context.Background(), userID, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isFav {
		t.Error("expected is_favorite=false after removing")
	}
}

func TestFavoriteService_Toggle_BathhouseNotFound(t *testing.T) {
	env := newFavoriteTestEnv()
	userID := uuid.New()

	_, err := env.svc.Toggle(context.Background(), userID, uuid.New())
	if err == nil {
		t.Error("expected error for non-existent bathhouse")
	}
}

func TestFavoriteService_List_Empty(t *testing.T) {
	env := newFavoriteTestEnv()
	userID := uuid.New()

	result, err := env.svc.List(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("totalCount = %d, want 0", result.TotalCount)
	}
}

func TestFavoriteService_List_WithItems(t *testing.T) {
	env := newFavoriteTestEnv()
	ownerID := uuid.New()
	userID := uuid.New()

	bh1 := createBathhouse(t, env.bhRepo, ownerID)
	bh2 := createBathhouse(t, env.bhRepo, ownerID)

	_, _ = env.svc.Toggle(context.Background(), userID, bh1.ID)
	_, _ = env.svc.Toggle(context.Background(), userID, bh2.ID)

	result, err := env.svc.List(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("totalCount = %d, want 2", result.TotalCount)
	}
}

func TestFavoriteService_IsFavorite(t *testing.T) {
	env := newFavoriteTestEnv()
	ownerID := uuid.New()
	userID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	// Not favorite yet
	isFav, err := env.svc.IsFavorite(context.Background(), userID, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isFav {
		t.Error("expected is_favorite=false before adding")
	}

	// Add to favorites
	_, _ = env.svc.Toggle(context.Background(), userID, bh.ID)

	// Should be favorite now
	isFav, err = env.svc.IsFavorite(context.Background(), userID, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isFav {
		t.Error("expected is_favorite=true after adding")
	}
}

func TestFavoriteService_Toggle_DoubleAdd(t *testing.T) {
	env := newFavoriteTestEnv()
	ownerID := uuid.New()
	userID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	// Add
	isFav, err := env.svc.Toggle(context.Background(), userID, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isFav {
		t.Error("expected is_favorite=true")
	}

	// Remove
	isFav, err = env.svc.Toggle(context.Background(), userID, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isFav {
		t.Error("expected is_favorite=false")
	}

	// Add again
	isFav, err = env.svc.Toggle(context.Background(), userID, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isFav {
		t.Error("expected is_favorite=true after re-adding")
	}
}
