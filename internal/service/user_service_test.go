package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func TestUserService_GetByID(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := service.NewUserService(userRepo)

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
	svc := service.NewUserService(userRepo)

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

func TestUserService_Block(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := service.NewUserService(userRepo)

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
	svc := service.NewUserService(userRepo)

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
	svc := service.NewUserService(userRepo)

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
	svc := service.NewUserService(userRepo)

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("should return ErrNotFound, got: %v", err)
	}
}
