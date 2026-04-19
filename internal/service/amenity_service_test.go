package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository/mock"
)

func newAmenityService() AmenityService {
	repo := mock.NewAmenityRepo()
	return NewAmenityService(repo)
}

func TestAmenityService_CRUD(t *testing.T) {
	svc := newAmenityService()
	ctx := context.Background()

	// Create
	amenity := &domain.Amenity{
		Name:     "Бассейн",
		Icon:     "pool",
		IsActive: true,
	}
	if err := svc.Create(ctx, amenity); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if amenity.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	// GetByID
	got, err := svc.GetByID(ctx, amenity.ID)
	if err != nil {
		t.Fatalf("getByID failed: %v", err)
	}
	if got.Name != "Бассейн" {
		t.Errorf("expected name 'Бассейн', got '%s'", got.Name)
	}
	if got.Icon != "pool" {
		t.Errorf("expected icon 'pool', got '%s'", got.Icon)
	}

	// Update
	amenity.Name = "Большой бассейн"
	amenity.Icon = "pool_large"
	if err := svc.Update(ctx, amenity); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	got, _ = svc.GetByID(ctx, amenity.ID)
	if got.Name != "Большой бассейн" {
		t.Errorf("expected updated name, got '%s'", got.Name)
	}
	if got.Icon != "pool_large" {
		t.Errorf("expected updated icon, got '%s'", got.Icon)
	}

	// ListAll
	list, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("listAll failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 amenity, got %d", len(list))
	}

	// Delete
	if err := svc.Delete(ctx, amenity.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, err = svc.GetByID(ctx, amenity.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestAmenityService_Validate(t *testing.T) {
	svc := newAmenityService()
	ctx := context.Background()

	tests := []struct {
		name    string
		amenity *domain.Amenity
		wantErr bool
	}{
		{
			name:    "empty name",
			amenity: &domain.Amenity{Name: "", Icon: "test", IsActive: true},
			wantErr: true,
		},
		{
			name:    "negative sort order",
			amenity: &domain.Amenity{Name: "Test", SortOrder: -1, IsActive: true},
			wantErr: true,
		},
		{
			name:    "valid amenity",
			amenity: &domain.Amenity{Name: "Сауна", Icon: "sauna", IsActive: true},
			wantErr: false,
		},
		{
			name:    "valid with zero sort order",
			amenity: &domain.Amenity{Name: "Мангал", SortOrder: 0, IsActive: true},
			wantErr: false,
		},
		{
			name:    "valid inactive",
			amenity: &domain.Amenity{Name: "Караоке", IsActive: false},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(ctx, tt.amenity)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAmenityService_DeleteNotFound(t *testing.T) {
	svc := newAmenityService()
	ctx := context.Background()

	err := svc.Delete(ctx, uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAmenityService_UpdateNotFound(t *testing.T) {
	svc := newAmenityService()
	ctx := context.Background()

	err := svc.Update(ctx, &domain.Amenity{ID: uuid.New(), Name: "Test"})
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
