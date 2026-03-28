package domain

// AdminSubRole represents a granular admin sub-role.
type AdminSubRole string

const (
	AdminSubRoleSuperAdmin AdminSubRole = "super_admin"
	AdminSubRoleModerator  AdminSubRole = "moderator"
	AdminSubRoleSupportL1  AdminSubRole = "support_l1"
	AdminSubRoleSupportL2  AdminSubRole = "support_l2"
	AdminSubRoleSupportL3  AdminSubRole = "support_l3"
	AdminSubRoleFinance    AdminSubRole = "finance"
)

func (r AdminSubRole) IsValid() bool {
	switch r {
	case AdminSubRoleSuperAdmin, AdminSubRoleModerator,
		AdminSubRoleSupportL1, AdminSubRoleSupportL2, AdminSubRoleSupportL3,
		AdminSubRoleFinance:
		return true
	}
	return false
}

// AdminPermission represents a specific admin capability.
type AdminPermission string

const (
	PermUserManage          AdminPermission = "users.manage"
	PermBathhouseModerate   AdminPermission = "bathhouses.moderate"
	PermReviewModerate      AdminPermission = "reviews.moderate"
	PermPhotoModerate       AdminPermission = "photos.moderate"
	PermKYCModerate         AdminPermission = "kyc.moderate"
	PermComplaintManage     AdminPermission = "complaints.manage"
	PermBookingManage       AdminPermission = "bookings.manage"
	PermAnalyticsView       AdminPermission = "analytics.view"
	PermFinanceManage       AdminPermission = "finance.manage"
	PermWalletManage        AdminPermission = "wallets.manage"
	PermSettingsManage      AdminPermission = "settings.manage"
	PermFeatureFlagsManage  AdminPermission = "feature_flags.manage"
	PermForceMajeureManage  AdminPermission = "force_majeure.manage"
	PermAntiFraudManage     AdminPermission = "antifraud.manage"
	PermTicketManage        AdminPermission = "tickets.manage"
	PermDisputeManage       AdminPermission = "disputes.manage"
	PermPromoManage         AdminPermission = "promo.manage"
	PermServiceFeeManage    AdminPermission = "service_fee.manage"
	PermHolidayManage       AdminPermission = "holidays.manage"
	PermAmenityManage       AdminPermission = "amenities.manage"
	PermObjectTypeManage    AdminPermission = "object_types.manage"
	PermAuditLogView        AdminPermission = "audit_log.view"
	PermReconciliationView  AdminPermission = "reconciliation.view"
	PermAdminRolesManage    AdminPermission = "admin_roles.manage"
	PermCityManage          AdminPermission = "cities.manage"
)

// AdminRolePermissions defines the permission matrix: which sub-role has which permissions.
var AdminRolePermissions = map[AdminSubRole][]AdminPermission{
	AdminSubRoleSuperAdmin: {
		PermUserManage, PermBathhouseModerate, PermReviewModerate, PermPhotoModerate,
		PermKYCModerate, PermComplaintManage, PermBookingManage, PermAnalyticsView,
		PermFinanceManage, PermWalletManage, PermSettingsManage, PermFeatureFlagsManage,
		PermForceMajeureManage, PermAntiFraudManage, PermTicketManage, PermDisputeManage,
		PermPromoManage, PermServiceFeeManage, PermHolidayManage, PermAmenityManage,
		PermObjectTypeManage, PermAuditLogView, PermReconciliationView, PermAdminRolesManage,
		PermCityManage,
	},
	AdminSubRoleModerator: {
		PermBathhouseModerate, PermReviewModerate, PermPhotoModerate,
		PermKYCModerate, PermComplaintManage, PermAuditLogView,
	},
	AdminSubRoleSupportL1: {
		PermTicketManage, PermBookingManage,
	},
	AdminSubRoleSupportL2: {
		PermTicketManage, PermBookingManage, PermComplaintManage,
		PermDisputeManage,
	},
	AdminSubRoleSupportL3: {
		PermTicketManage, PermBookingManage, PermComplaintManage,
		PermDisputeManage, PermUserManage, PermWalletManage,
	},
	AdminSubRoleFinance: {
		PermFinanceManage, PermWalletManage, PermReconciliationView,
		PermAnalyticsView, PermServiceFeeManage, PermAuditLogView,
	},
}

// HasPermission checks whether a sub-role has a specific permission.
func (r AdminSubRole) HasPermission(perm AdminPermission) bool {
	perms, ok := AdminRolePermissions[r]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// AllAdminSubRoles returns all valid admin sub-roles.
func AllAdminSubRoles() []AdminSubRole {
	return []AdminSubRole{
		AdminSubRoleSuperAdmin,
		AdminSubRoleModerator,
		AdminSubRoleSupportL1,
		AdminSubRoleSupportL2,
		AdminSubRoleSupportL3,
		AdminSubRoleFinance,
	}
}

// AllAdminPermissions returns all defined permissions.
func AllAdminPermissions() []AdminPermission {
	return []AdminPermission{
		PermUserManage, PermBathhouseModerate, PermReviewModerate, PermPhotoModerate,
		PermKYCModerate, PermComplaintManage, PermBookingManage, PermAnalyticsView,
		PermFinanceManage, PermWalletManage, PermSettingsManage, PermFeatureFlagsManage,
		PermForceMajeureManage, PermAntiFraudManage, PermTicketManage, PermDisputeManage,
		PermPromoManage, PermServiceFeeManage, PermHolidayManage, PermAmenityManage,
		PermObjectTypeManage, PermAuditLogView, PermReconciliationView, PermAdminRolesManage,
		PermCityManage,
	}
}
