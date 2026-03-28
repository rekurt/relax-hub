package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func newObjectTypeService() ObjectTypeService {
	repo := mock.NewObjectTypeRepo()
	return NewObjectTypeService(repo)
}

func TestObjectTypeService_CRUD(t *testing.T) {
	svc := newObjectTypeService()
	ctx := context.Background()

	// Create
	objType := &domain.ObjectType{
		Name:        "Русская баня",
		Description: "Традиционная русская баня с парной",
		IsActive:    true,
	}
	if err := svc.Create(ctx, objType); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if objType.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	// GetByID
	got, err := svc.GetByID(ctx, objType.ID)
	if err != nil {
		t.Fatalf("getByID failed: %v", err)
	}
	if got.Name != "Русская баня" {
		t.Errorf("expected name 'Русская баня', got '%s'", got.Name)
	}
	if got.Description != "Традиционная русская баня с парной" {
		t.Errorf("expected description to match, got '%s'", got.Description)
	}

	// Update
	objType.Name = "Русская парная"
	objType.Description = "Обновленное описание"
	if err := svc.Update(ctx, objType); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	got, _ = svc.GetByID(ctx, objType.ID)
	if got.Name != "Русская парная" {
		t.Errorf("expected updated name, got '%s'", got.Name)
	}

	// ListAll
	list, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("listAll failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 object type, got %d", len(list))
	}

	// Delete
	if err := svc.Delete(ctx, objType.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, err = svc.GetByID(ctx, objType.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestObjectTypeService_Validate(t *testing.T) {
	svc := newObjectTypeService()
	ctx := context.Background()

	tests := []struct {
		name    string
		objType *domain.ObjectType
		wantErr bool
	}{
		{
			name:    "empty name",
			objType: &domain.ObjectType{Name: "", IsActive: true},
			wantErr: true,
		},
		{
			name:    "negative sort order",
			objType: &domain.ObjectType{Name: "Test", SortOrder: -1, IsActive: true},
			wantErr: true,
		},
		{
			name:    "valid object type",
			objType: &domain.ObjectType{Name: "Финская сауна", Description: "Сухой пар", IsActive: true},
			wantErr: false,
		},
		{
			name:    "valid without description",
			objType: &domain.ObjectType{Name: "Хаммам", IsActive: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(ctx, tt.objType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectTypeService_DeleteNotFound(t *testing.T) {
	svc := newObjectTypeService()
	ctx := context.Background()

	err := svc.Delete(ctx, uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestObjectTypeService_UpdateNotFound(t *testing.T) {
	svc := newObjectTypeService()
	ctx := context.Background()

	err := svc.Update(ctx, &domain.ObjectType{ID: uuid.New(), Name: "Test"})
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestObjectTypeService_MultipleItems(t *testing.T) {
	svc := newObjectTypeService()
	ctx := context.Background()

	types := []string{"Русская баня", "Финская сауна", "Хаммам"}
	for _, name := range types {
		if err := svc.Create(ctx, &domain.ObjectType{Name: name, IsActive: true}); err != nil {
			t.Fatalf("create failed for %s: %v", name, err)
		}
	}

	list, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("listAll failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 object types, got %d", len(list))
	}
}
