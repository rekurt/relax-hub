package domain

import "testing"

func TestAdminSubRole_IsValid(t *testing.T) {
	tests := []struct {
		role  AdminSubRole
		valid bool
	}{
		{AdminSubRoleSuperAdmin, true},
		{AdminSubRoleModerator, true},
		{AdminSubRoleSupportL1, true},
		{AdminSubRoleSupportL2, true},
		{AdminSubRoleSupportL3, true},
		{AdminSubRoleFinance, true},
		{AdminSubRole("unknown"), false},
		{AdminSubRole(""), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.valid {
				t.Errorf("AdminSubRole(%q).IsValid() = %v, want %v", tt.role, got, tt.valid)
			}
		})
	}
}

func TestAdminSubRole_HasPermission(t *testing.T) {
	tests := []struct {
		name     string
		role     AdminSubRole
		perm     AdminPermission
		expected bool
	}{
		{"super_admin has all permissions", AdminSubRoleSuperAdmin, PermUserManage, true},
		{"super_admin has finance", AdminSubRoleSuperAdmin, PermFinanceManage, true},
		{"super_admin has admin_roles", AdminSubRoleSuperAdmin, PermAdminRolesManage, true},
		{"moderator can moderate bathhouses", AdminSubRoleModerator, PermBathhouseModerate, true},
		{"moderator can moderate reviews", AdminSubRoleModerator, PermReviewModerate, true},
		{"moderator cannot manage finance", AdminSubRoleModerator, PermFinanceManage, false},
		{"moderator cannot manage users", AdminSubRoleModerator, PermUserManage, false},
		{"moderator cannot manage admin roles", AdminSubRoleModerator, PermAdminRolesManage, false},
		{"support_l1 can manage tickets", AdminSubRoleSupportL1, PermTicketManage, true},
		{"support_l1 can manage bookings", AdminSubRoleSupportL1, PermBookingManage, true},
		{"support_l1 cannot manage disputes", AdminSubRoleSupportL1, PermDisputeManage, false},
		{"support_l2 can manage disputes", AdminSubRoleSupportL2, PermDisputeManage, true},
		{"support_l3 can manage users", AdminSubRoleSupportL3, PermUserManage, true},
		{"support_l3 can manage wallets", AdminSubRoleSupportL3, PermWalletManage, true},
		{"finance can manage finance", AdminSubRoleFinance, PermFinanceManage, true},
		{"finance can view reconciliation", AdminSubRoleFinance, PermReconciliationView, true},
		{"finance cannot manage users", AdminSubRoleFinance, PermUserManage, false},
		{"invalid role has no permissions", AdminSubRole("invalid"), PermUserManage, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.HasPermission(tt.perm); got != tt.expected {
				t.Errorf("AdminSubRole(%q).HasPermission(%q) = %v, want %v", tt.role, tt.perm, got, tt.expected)
			}
		})
	}
}

func TestAllAdminSubRoles(t *testing.T) {
	roles := AllAdminSubRoles()
	if len(roles) != 6 {
		t.Errorf("expected 6 admin sub-roles, got %d", len(roles))
	}
	for _, r := range roles {
		if !r.IsValid() {
			t.Errorf("AllAdminSubRoles() returned invalid role: %q", r)
		}
	}
}

func TestAllAdminPermissions(t *testing.T) {
	perms := AllAdminPermissions()
	if len(perms) < 20 {
		t.Errorf("expected at least 20 permissions, got %d", len(perms))
	}
}

func TestPermissionMatrixConsistency(t *testing.T) {
	allPerms := AllAdminPermissions()
	permSet := make(map[AdminPermission]bool)
	for _, p := range allPerms {
		permSet[p] = true
	}

	for role, perms := range AdminRolePermissions {
		for _, p := range perms {
			if !permSet[p] {
				t.Errorf("role %q has permission %q which is not in AllAdminPermissions()", role, p)
			}
		}
	}

	// super_admin must have all permissions
	superPerms := AdminRolePermissions[AdminSubRoleSuperAdmin]
	superSet := make(map[AdminPermission]bool)
	for _, p := range superPerms {
		superSet[p] = true
	}
	for _, p := range allPerms {
		if !superSet[p] {
			t.Errorf("super_admin is missing permission %q", p)
		}
	}
}
