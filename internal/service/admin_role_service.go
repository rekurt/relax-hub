package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// AdminRoleService manages admin sub-roles and permissions.
type AdminRoleService interface {
	GetAdminSubRole(ctx context.Context, userID uuid.UUID) (domain.AdminSubRole, error)
	HasEnabled2FA(ctx context.Context, userID uuid.UUID) (bool, error)
	SetAdminSubRole(ctx context.Context, userID uuid.UUID, subRole domain.AdminSubRole) error
	ListAdminUsers(ctx context.Context) ([]AdminUserInfo, error)
	GetPermissionsMatrix() map[domain.AdminSubRole][]domain.AdminPermission
}

// AdminUserInfo is a lightweight admin user summary.
type AdminUserInfo struct {
	ID           uuid.UUID            `json:"id"`
	Email        string               `json:"email"`
	Name         string               `json:"name"`
	AdminSubRole domain.AdminSubRole  `json:"admin_sub_role"`
	TwoFAMethod  domain.TwoFAMethod   `json:"two_fa_method"`
	IsActive     bool                 `json:"is_active"`
}

type adminRoleService struct {
	userRepo repository.UserRepository
}

func NewAdminRoleService(userRepo repository.UserRepository) AdminRoleService {
	return &adminRoleService{userRepo: userRepo}
}

func (s *adminRoleService) GetAdminSubRole(ctx context.Context, userID uuid.UUID) (domain.AdminSubRole, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get admin sub-role: %w", err)
	}
	if user.Role != domain.RoleAdmin {
		return "", domain.ErrForbidden
	}
	if user.AdminSubRole == "" {
		return domain.AdminSubRoleSuperAdmin, nil
	}
	return user.AdminSubRole, nil
}

func (s *adminRoleService) HasEnabled2FA(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("check admin 2FA: %w", err)
	}
	return user.TwoFAMethod != domain.TwoFANone, nil
}

func (s *adminRoleService) SetAdminSubRole(ctx context.Context, userID uuid.UUID, subRole domain.AdminSubRole) error {
	if !subRole.IsValid() {
		return domain.ErrInvalidInput
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("set admin sub-role: %w", err)
	}

	if user.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	user.AdminSubRole = subRole
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update admin sub-role: %w", err)
	}
	return nil
}

func (s *adminRoleService) ListAdminUsers(ctx context.Context) ([]AdminUserInfo, error) {
	result, err := s.userRepo.List(ctx, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("list admin users: %w", err)
	}

	var admins []AdminUserInfo
	for _, u := range result.Items {
		if u.Role != domain.RoleAdmin {
			continue
		}
		admins = append(admins, AdminUserInfo{
			ID:           u.ID,
			Email:        u.Email,
			Name:         u.Name,
			AdminSubRole: u.AdminSubRole,
			TwoFAMethod:  u.TwoFAMethod,
			IsActive:     u.IsActive,
		})
	}
	return admins, nil
}

func (s *adminRoleService) GetPermissionsMatrix() map[domain.AdminSubRole][]domain.AdminPermission {
	return domain.AdminRolePermissions
}
