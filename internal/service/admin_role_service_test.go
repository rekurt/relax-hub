package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func setupAdminRoleTest(t *testing.T) (AdminRoleService, *mock.UserRepo, *domain.User) {
	t.Helper()
	userRepo := mock.NewUserRepo()

	adminUser := &domain.User{
		ID:           uuid.New(),
		Email:        "admin@test.com",
		Name:         "Test Admin",
		Role:         domain.RoleAdmin,
		AdminSubRole: domain.AdminSubRoleSuperAdmin,
		TwoFAMethod:  domain.TwoFATOTP,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(context.Background(), adminUser); err != nil {
		t.Fatal(err)
	}

	svc := NewAdminRoleService(userRepo)
	return svc, userRepo, adminUser
}

func TestAdminRoleService_GetAdminSubRole(t *testing.T) {
	svc, _, adminUser := setupAdminRoleTest(t)

	subRole, err := svc.GetAdminSubRole(context.Background(), adminUser.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subRole != domain.AdminSubRoleSuperAdmin {
		t.Errorf("expected super_admin, got %s", subRole)
	}
}

func TestAdminRoleService_GetAdminSubRole_DefaultForEmpty(t *testing.T) {
	userRepo := mock.NewUserRepo()
	adminUser := &domain.User{
		ID:           uuid.New(),
		Email:        "admin2@test.com",
		Name:         "Admin No Role",
		Role:         domain.RoleAdmin,
		AdminSubRole: "",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = userRepo.Create(context.Background(), adminUser)

	svc := NewAdminRoleService(userRepo)
	subRole, err := svc.GetAdminSubRole(context.Background(), adminUser.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subRole != domain.AdminSubRoleSuperAdmin {
		t.Errorf("expected super_admin default, got %s", subRole)
	}
}

func TestAdminRoleService_GetAdminSubRole_NonAdmin(t *testing.T) {
	userRepo := mock.NewUserRepo()
	clientUser := &domain.User{
		ID:        uuid.New(),
		Email:     "client@test.com",
		Name:      "Client",
		Role:      domain.RoleClient,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = userRepo.Create(context.Background(), clientUser)

	svc := NewAdminRoleService(userRepo)
	_, err := svc.GetAdminSubRole(context.Background(), clientUser.ID)
	if err == nil {
		t.Error("expected error for non-admin user")
	}
}

func TestAdminRoleService_SetAdminSubRole(t *testing.T) {
	svc, userRepo, adminUser := setupAdminRoleTest(t)

	err := svc.SetAdminSubRole(context.Background(), adminUser.ID, domain.AdminSubRoleModerator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it was updated
	updated, err := userRepo.GetByID(context.Background(), adminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.AdminSubRole != domain.AdminSubRoleModerator {
		t.Errorf("expected moderator, got %s", updated.AdminSubRole)
	}
}

func TestAdminRoleService_SetAdminSubRole_Invalid(t *testing.T) {
	svc, _, _ := setupAdminRoleTest(t)

	err := svc.SetAdminSubRole(context.Background(), uuid.New(), domain.AdminSubRole("invalid"))
	if err == nil {
		t.Error("expected error for invalid sub-role")
	}
}

func TestAdminRoleService_SetAdminSubRole_NonAdmin(t *testing.T) {
	userRepo := mock.NewUserRepo()
	clientUser := &domain.User{
		ID:        uuid.New(),
		Email:     "client@test.com",
		Name:      "Client",
		Role:      domain.RoleClient,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = userRepo.Create(context.Background(), clientUser)

	svc := NewAdminRoleService(userRepo)
	err := svc.SetAdminSubRole(context.Background(), clientUser.ID, domain.AdminSubRoleModerator)
	if err == nil {
		t.Error("expected error for non-admin user")
	}
}

func TestAdminRoleService_HasEnabled2FA(t *testing.T) {
	svc, _, adminUser := setupAdminRoleTest(t)

	has2FA, err := svc.HasEnabled2FA(context.Background(), adminUser.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has2FA {
		t.Error("admin with TOTP should have 2FA enabled")
	}
}

func TestAdminRoleService_HasEnabled2FA_None(t *testing.T) {
	userRepo := mock.NewUserRepo()
	adminUser := &domain.User{
		ID:          uuid.New(),
		Email:       "admin-no2fa@test.com",
		Name:        "Admin No 2FA",
		Role:        domain.RoleAdmin,
		TwoFAMethod: domain.TwoFANone,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = userRepo.Create(context.Background(), adminUser)

	svc := NewAdminRoleService(userRepo)
	has2FA, err := svc.HasEnabled2FA(context.Background(), adminUser.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has2FA {
		t.Error("admin without 2FA should return false")
	}
}

func TestAdminRoleService_ListAdminUsers(t *testing.T) {
	svc, userRepo, _ := setupAdminRoleTest(t)

	// Add a non-admin user
	clientUser := &domain.User{
		ID:        uuid.New(),
		Email:     "client@test.com",
		Name:      "Client User",
		Role:      domain.RoleClient,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = userRepo.Create(context.Background(), clientUser)

	// Add another admin
	moderator := &domain.User{
		ID:           uuid.New(),
		Email:        "mod@test.com",
		Name:         "Moderator",
		Role:         domain.RoleAdmin,
		AdminSubRole: domain.AdminSubRoleModerator,
		TwoFAMethod:  domain.TwoFASMS,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = userRepo.Create(context.Background(), moderator)

	admins, err := svc.ListAdminUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(admins) != 2 {
		t.Errorf("expected 2 admin users, got %d", len(admins))
	}
}

func TestAdminRoleService_GetPermissionsMatrix(t *testing.T) {
	svc, _, _ := setupAdminRoleTest(t)
	matrix := svc.GetPermissionsMatrix()
	if len(matrix) != 6 {
		t.Errorf("expected 6 roles in matrix, got %d", len(matrix))
	}
	superPerms := matrix[domain.AdminSubRoleSuperAdmin]
	if len(superPerms) < 20 {
		t.Errorf("super_admin should have at least 20 permissions, got %d", len(superPerms))
	}
}
