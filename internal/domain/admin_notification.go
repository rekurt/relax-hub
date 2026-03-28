package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AdminNotificationType represents the type of admin notification.
type AdminNotificationType string

const (
	AdminNotifAntifraudFlag          AdminNotificationType = "antifraud_flag"
	AdminNotifSLAViolation           AdminNotificationType = "sla_violation"
	AdminNotifReconciliationMismatch AdminNotificationType = "reconciliation_mismatch"
	AdminNotifFloatDrift             AdminNotificationType = "float_drift"
	AdminNotifTicketEscalation       AdminNotificationType = "ticket_escalation"
	AdminNotifDisputeOpened          AdminNotificationType = "dispute_opened"
	AdminNotifKYCPending             AdminNotificationType = "kyc_pending"
	AdminNotifSystemAlert            AdminNotificationType = "system_alert"
)

func (t AdminNotificationType) IsValid() bool {
	switch t {
	case AdminNotifAntifraudFlag, AdminNotifSLAViolation,
		AdminNotifReconciliationMismatch, AdminNotifFloatDrift,
		AdminNotifTicketEscalation, AdminNotifDisputeOpened,
		AdminNotifKYCPending, AdminNotifSystemAlert:
		return true
	}
	return false
}

// AdminNotifSeverity represents the severity level.
type AdminNotifSeverity string

const (
	AdminNotifSeverityInfo     AdminNotifSeverity = "info"
	AdminNotifSeverityWarning  AdminNotifSeverity = "warning"
	AdminNotifSeverityError    AdminNotifSeverity = "error"
	AdminNotifSeverityCritical AdminNotifSeverity = "critical"
)

func (s AdminNotifSeverity) IsValid() bool {
	switch s {
	case AdminNotifSeverityInfo, AdminNotifSeverityWarning,
		AdminNotifSeverityError, AdminNotifSeverityCritical:
		return true
	}
	return false
}

// AdminNotification is a notification targeted at admin users by role.
type AdminNotification struct {
	ID        uuid.UUID
	Role      AdminSubRole
	Severity  AdminNotifSeverity
	Type      AdminNotificationType
	Title     string
	Body      string
	Data      json.RawMessage
	IsRead    bool
	ReadAt    *time.Time
	ReadBy    *uuid.UUID
	CreatedAt time.Time
}

// AdminNotificationFilter specifies query parameters for listing admin notifications.
type AdminNotificationFilter struct {
	Role     *AdminSubRole
	Severity *AdminNotifSeverity
	Type     *AdminNotificationType
	IsRead   *bool
	Page     int
	PageSize int
}

// AdminNotifTargetRoles returns the admin sub-roles that should receive a given notification type.
func AdminNotifTargetRoles(notifType AdminNotificationType) []AdminSubRole {
	switch notifType {
	case AdminNotifAntifraudFlag:
		return []AdminSubRole{AdminSubRoleSuperAdmin, AdminSubRoleFinance}
	case AdminNotifSLAViolation:
		return []AdminSubRole{AdminSubRoleSuperAdmin, AdminSubRoleModerator}
	case AdminNotifReconciliationMismatch, AdminNotifFloatDrift:
		return []AdminSubRole{AdminSubRoleSuperAdmin, AdminSubRoleFinance}
	case AdminNotifTicketEscalation:
		return []AdminSubRole{AdminSubRoleSuperAdmin, AdminSubRoleSupportL2, AdminSubRoleSupportL3}
	case AdminNotifDisputeOpened:
		return []AdminSubRole{AdminSubRoleSuperAdmin, AdminSubRoleSupportL2, AdminSubRoleSupportL3}
	case AdminNotifKYCPending:
		return []AdminSubRole{AdminSubRoleSuperAdmin, AdminSubRoleModerator}
	case AdminNotifSystemAlert:
		return []AdminSubRole{AdminSubRoleSuperAdmin}
	default:
		return []AdminSubRole{AdminSubRoleSuperAdmin}
	}
}
